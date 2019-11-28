package pudge

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"rura/ag-server/extcon"
	"rura/ag-server/logger"
	"rura/ag-server/setup"
	"strconv"
	"sync"
	"time"

	"github.com/lib/pq"
)

var mutex sync.Mutex
var mapContrs map[int]Controller

//InWriteServARM тут принимаем запросы от сервера АРМ для исполнения сервером Коммуникации
var InWriteServARM chan CommandARM

//ToServComm туда отправляем запросы от сервера АРМ для исполнения после логгирования в БД
var ToServComm chan CommandARM
var conDBLog *sql.DB
var conDBSave *sql.DB
var err error

//GetController возвращает копию Контроллера
func GetController(id int) (Controller, bool) {
	var c Controller
	mutex.Lock()
	c, is := mapContrs[id]
	mutex.Unlock()
	if !is {
		return c, false
	}
	return c, true
}

//SetController Записывает новое состояние контроллера и если есть изменения то записывает его в лог
func SetController(c Controller) {
	mutex.Lock()
	write := false
	cc, is := mapContrs[c.ID]
	if !is {
		write = true
	} else {
		write = cc.isChanged(c)
	}
	mapContrs[c.ID] = c
	mutex.Unlock()
	if write {
		t := time.Now()
		js, _ := json.Marshal(c)
		w := "insert into " + setup.Set.Pudge.TableLog + " (tm,flag,id,txt) values('" + string(pq.FormatTimestamp(t)) +
			"',2," + strconv.Itoa(c.ID) + ",'" + string(js) + "');"
		_, err := conDBLog.Exec(w)
		if err != nil {
			logger.Error.Printf("For wtite log to controller %s", err.Error())
			return
		}
		_, err = conDBSave.Exec("insert into " + setup.Set.Pudge.TableSave + "(id,device) values(" + strconv.Itoa(c.ID) + ",'" +
			string(js) + "');")
		if err != nil {
			logger.Error.Printf("For wtite log to controller %s", err.Error())
			return
		}
	}
}

//Start главная процедура управления состоянием котроллеров
func Start(context *extcon.ExtContext, stop chan int) {
	// Создаем каналы и переменные
	mapContrs = make(map[int]Controller)
	InWriteServARM = make(chan CommandARM)
	ToServComm = make(chan CommandARM)
	dbinfo := fmt.Sprintf("host=%s user=%s password=%s dbname=%s sslmode=disable",
		setup.Set.DataBase.Host, setup.Set.DataBase.User,
		setup.Set.DataBase.Password, setup.Set.DataBase.DBname)
	conDBLog, err = sql.Open("postgres", dbinfo)
	if err != nil {
		logger.Error.Printf("Запрос на открытие %s %s", dbinfo, err.Error())
		context.Cancel()
		stop <- 1
		return
	}
	defer conDBLog.Close()
	if err = conDBLog.Ping(); err != nil {
		logger.Error.Printf("Ping %s", err.Error())
		context.Cancel()
		stop <- 1
		return
	}
	conDBSave, err = sql.Open("postgres", dbinfo)
	if err != nil {
		logger.Error.Printf("Запрос на открытие %s %s", dbinfo, err.Error())
		context.Cancel()
		stop <- 1
		return
	}
	defer conDBSave.Close()
	if err = conDBSave.Ping(); err != nil {
		logger.Error.Printf("Ping %s", err.Error())
		context.Cancel()
		stop <- 1
		return
	}
	err = loadSave()
	if err != nil {
		logger.Error.Printf("save %s", err.Error())
		context.Cancel()
		stop <- 1
		return
	}
	context.SetTimeOut(time.Duration(setup.Set.Pudge.StepSave) * time.Second)
	for true {
		select {
		case msgARM := <-InWriteServARM:
			//Запротоколлировать приход команды
			//И передать для исполнения
			toLogCommad(msgARM)
			ToServComm <- msgARM
		case <-context.Done():
			if context.GetStatus() == "timeout" {
				saveSave()
				context.SetTimeOut(time.Duration(setup.Set.Pudge.StepSave) * time.Second)
			} else {
				saveSave()
				context.Cancel()
				return
			}
		}
	}

}
func toReturnControllers(mgs []int) {
	var ret Controllers
	ret.Contrs = make([]Controller, 0)
	mutex.Lock()
	for _, i := range mgs {
		ret.Contrs = append(ret.Contrs, mapContrs[i])
	}
	mutex.Unlock()
}
func toLogCommad(msg CommandARM) {
	t := time.Now()
	js, _ := json.Marshal(msg)
	w := "insert into " + setup.Set.Pudge.TableLog + " (tm,flag,id,json) values('" + string(pq.FormatTimestamp(t)) +
		"',1," + string(msg.ID) + ",'" + string(js) + "');"
	_, err := conDBLog.Exec(w)
	if err != nil {
		logger.Error.Printf("For wtite log to command %s", err.Error())
		return
	}
}
func loadSave() error {
	rows, err := conDBSave.Query("Select * from " + setup.Set.Pudge.TableSave + ";")
	if err != nil {
		return err
	}
	defer rows.Close()
	mutex.Lock()
	defer mutex.Unlock()
	var id int
	var js []byte
	var c Controller
	for rows.Next() {
		err = rows.Scan(&id, &js)
		if err != nil {
			return err
		}
		err = json.Unmarshal(js, &c)
		if err != nil {
			return err
		}
		c.WriteToDB = false
		mapContrs[id] = c
	}
	return nil
}
func saveSave() error {
	mutex.Lock()
	copyMap := make(map[int]Controller)
	for _, c := range mapContrs {
		copyMap[c.ID] = c
	}
	mutex.Unlock()
	for _, c := range copyMap {
		if c.WriteToDB {
			js, _ := json.Marshal(c)
			_, err = conDBSave.Exec("update into " + setup.Set.Pudge.TableSave + "(id,device) values(" + strconv.Itoa(c.ID) + ",'" +
				string(js) + "');")
			mutex.Lock()
			cc := mapContrs[c.ID]
			cc.WriteToDB = false
			mapContrs[c.ID] = cc
			mutex.Unlock()
		}
	}
	return nil
}
