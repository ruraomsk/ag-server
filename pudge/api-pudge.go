package pudge

import (
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/lib/pq"
	"rura/ag-server/extcon"
	"rura/ag-server/logger"
	"rura/ag-server/setup"
	"strconv"
	"sync"
	"time"
)

var mutex sync.Mutex
var mapContrs map[int]Controller

var conDBLog *sql.DB
var conDBSave *sql.DB
var conDevGis *sql.DB
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
	defer mutex.Unlock()
	insert := false
	_, is := mapContrs[c.ID]
	if !is {
		insert = true
	}
	mapContrs[c.ID] = c
	c.WriteToDB = false
	js, _ := json.Marshal(c)
	if insert {
		w := "insert into " + setup.Set.Pudge.TableSave + " (id,device) values(" + strconv.Itoa(c.ID) + ",'" + string(js) + "');"
		_, err := conDBSave.Exec(w)
		if err != nil {
			logger.Error.Printf("For insert to controller %s", err.Error())
			return
		}
	} else {
		_, err = conDBSave.Exec("update  " + setup.Set.Pudge.TableSave + " set device='" + string(js) + "' where id=" + strconv.Itoa(c.ID) + ";")
		if err != nil {
			logger.Error.Printf("For update to controller %s", err.Error())
			return
		}
	}
}

//Start главная процедура управления состоянием котроллеров
func Start(context *extcon.ExtContext, stop chan int) {
	// Создаем каналы и переменные
	mapContrs = make(map[int]Controller)
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
	conDevGis, err = sql.Open("postgres", dbinfo)
	if err != nil {
		logger.Error.Printf("Запрос на открытие %s %s", dbinfo, err.Error())
		context.Cancel()
		stop <- 1
		return
	}
	defer conDevGis.Close()

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
		case <-context.Done():
			if context.GetStatus() == "timeout" {
				saveSave()
				logger.Info.Println("Save DB")
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
	defer mutex.Unlock()
	for _, c := range mapContrs {
		if !c.WriteToDB {
			continue
		}
		js, _ := json.Marshal(c)
		_, err = conDBSave.Exec("update  " + setup.Set.Pudge.TableSave + " set device='" + string(js) + "' where id=" + strconv.Itoa(c.ID) + ";")
		if err != nil {
			logger.Error.Printf("For update save to controller %s", err.Error())
			break
		}
		c.WriteToDB = false
	}
	return nil
}
