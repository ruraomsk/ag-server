package pudge

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/ruraomsk/ag-server/extcon"
	"github.com/ruraomsk/ag-server/logger"
	"github.com/ruraomsk/ag-server/setup"

	//Инициализатор постргресса
	_ "github.com/lib/pq"
)

//var mutexCross sync.Mutex
var mutexCtrl sync.Mutex
var controllers map[int]*Controller

var crosses map[Region]*Cross
var statuses map[int]string
var controls map[int]bool
var nowstatus map[Region]string
var firstLoad = true

//Works флаг готовности pudge
var Works bool

//ChanLog канал приема сообщений логов устройств
var ChanLog chan LogRecord
var conDBSave *sql.DB
var conCross *sql.DB
var conLog *sql.DB
var clearLog *sql.DB
var XTCommand chan CommandXT

var Reload chan interface{}
var dbinfo string
var err error

// func Lock() {
// 	mutexCtrl.Lock()
// }
// func Unclock() {
// 	mutexCtrl.Unlock()

// }

//GetCross возвращает копию перекрестка
func GetCross(reg Region) (Cross, bool) {
	mutexCtrl.Lock()
	defer mutexCtrl.Unlock()
	c, is := crosses[reg]
	if !is {
		return *NewCross(), is
	}
	return *c, is
}
func (c *Cross) GetWorkPhases() []int {
	result := make([]int, 0)
	rph := make(map[int]int)
	for _, om := range c.Arrays.MonthSets.MonthSets {
		for _, nw := range om.Days {
			for _, w := range c.Arrays.WeekSets.GetWeek(nw) {
				pks := c.Arrays.DaySets.GetPKs(w)
				for _, pk := range pks {
					for _, ph := range c.Arrays.SetDK.GetPhases(pk) {
						rph[ph] = ph
					}
				}
			}
		}
	}
	for p := range rph {
		result = append(result, p)
	}
	sort.Ints(result)
	return result
}

//GetCrosses возвращает все перекрестки
func GetCrosses() []Region {
	mutexCtrl.Lock()
	defer mutexCtrl.Unlock()
	r := make([]Region, 0)
	for reg := range crosses {
		r = append(r, reg)
	}
	return r
}
func GetControllers() []int {
	result := make([]int, 0)
	mutexCtrl.Lock()
	defer mutexCtrl.Unlock()
	for _, c := range controllers {
		result = append(result, c.ID)
	}
	return result
}
func getNameCross(idevice int) string {
	// mutexCtrl.Lock()
	// defer mutexCtrl.Unlock()
	for _, c := range crosses {
		if c.IDevice == idevice {
			return c.Name
		}
	}
	return ""
}

//DeleteCross Удаляет перекресток
func DeleteCross(reg Region) {
	mutexCtrl.Lock()
	delete(crosses, reg)
	mutexCtrl.Unlock()
	w := fmt.Sprintf("DELETE FROM public.\"cross\" WHERE region=%d and area=%d and id=%d;", reg.Region, reg.Area, reg.ID)
	_, err = conCross.Exec(w)
	if err != nil {
		logger.Error.Printf("Error %s  %s\n", w, err.Error())
	}
}

//GetController возвращает копию Контроллера
func GetController(id int) (*Controller, bool) {
	mutexCtrl.Lock()
	defer mutexCtrl.Unlock()
	c, is := controllers[id]
	return c, is
}

//SetCross обновляет состояние перекрестка
func SetCross(c Cross) {
	// _, f, l, _ := runtime.Caller(1)
	mutexCtrl.Lock()
	defer mutexCtrl.Unlock()
	reg := Region{Region: c.Region, Area: c.Area, ID: c.ID}
	_, is := crosses[reg]
	if !is {
		c.WriteToDB = false
		crosses[reg] = &c
		js, _ := json.Marshal(c)
		w := fmt.Sprintf("insert into public.\"cross\" (region,area,subarea,id,dgis,describ,idevice,status,state) values(%d,%d,%d,%d,point(%s),'%s',%d,%d,'%s');",
			c.Region, c.Area, c.SubArea, c.ID, c.Dgis, c.Name, c.IDevice, c.StatusDevice, string(js))
		_, err = conCross.Exec(w)

		if err != nil {
			logger.Error.Printf("Error %s  %s\n", w, err.Error())
			return
		}
	} else {
		c.WriteToDB = true
		crosses[reg] = &c
		// logger.Debug.Printf("%s-%d %v %d", f, l, reg, c.IDevice)
	}
}

//SetController Записывает новое состояние контроллера и если есть изменения то записывает его в лог
func SetController(c *Controller) {
	// logger.Debug.Printf("start setController %d", c.ID)
	insert := false
	mutexCtrl.Lock()
	defer mutexCtrl.Unlock()
	_, is := controllers[c.ID]
	if !is {
		insert = true
		c.WriteToDB = false
		controllers[c.ID] = c
	}
	if insert {
		js, _ := json.Marshal(c)
		w := "insert into devices (id,device) values(" + strconv.Itoa(c.ID) + ",'" + string(js) + "');"
		_, err := conDBSave.Exec(w)
		if err != nil {
			logger.Error.Printf("For insert to controller %d %s", c.ID, err.Error())
			return
		}
	} else {
		c.WriteToDB = true
	}
}

//Start главная процедура управления состоянием котроллеров
func Start(stop chan interface{}) {
	// Создаем каналы и переменные
	Works = false
	context, _ := extcon.NewContext("pudge")
	controllers = make(map[int]*Controller)
	crosses = make(map[Region]*Cross)
	statuses = make(map[int]string)
	ChanLog = make(chan LogRecord, 1000)
	XTCommand = make(chan CommandXT, 100)
	controls = make(map[int]bool)
	nowstatus = make(map[Region]string)
	Reload = make(chan interface{})
	dbinfo = fmt.Sprintf("host=%s user=%s password=%s dbname=%s sslmode=disable",
		setup.Set.DataBase.Host, setup.Set.DataBase.User,
		setup.Set.DataBase.Password, setup.Set.DataBase.DBname)

	for {
		conDBSave, err = sql.Open("postgres", dbinfo)
		if err != nil {
			logger.Error.Printf("Запрос на открытие %s %s", dbinfo, err.Error())
			time.Sleep(time.Second * 10)
			continue
		}
		if err = conDBSave.Ping(); err != nil {
			logger.Error.Printf("Ping %s", err.Error())
			time.Sleep(time.Second * 10)
			continue
		}
		break
	}
	defer conDBSave.Close()
	conCross, err = sql.Open("postgres", dbinfo)
	if err != nil {
		logger.Error.Printf("Запрос на открытие %s %s", dbinfo, err.Error())
		stop <- 1
		return
	}
	defer conCross.Close()
	_ = needHistoryCross(conCross)
	err = loadDBase()
	if err != nil {
		logger.Error.Printf("load %s", err.Error())
		stop <- 1
		return
	}
	conLog, err = sql.Open("postgres", dbinfo)
	if err != nil {
		logger.Error.Printf("Запрос на открытие %s %s", dbinfo, err.Error())
		stop <- 1
		return
	}
	clearLog, err = sql.Open("postgres", dbinfo)
	if err != nil {
		logger.Error.Printf("Запрос на открытие %s %s", dbinfo, err.Error())
		stop <- 1
		return
	}
	defer conLog.Close()
	go writeLog()
	Works = true
	go StatisticStart()
	timer := extcon.SetTimerClock(time.Duration(setup.Set.StepPudge) * time.Second)
	for {
		select {
		case <-Reload:
			logger.Info.Println("reload pudge start")
			err = loadDBase()
			if err != nil {
				logger.Error.Printf("load %s", err.Error())
				stop <- 1
				return
			}
			logger.Info.Println("reload pudge stop")

		case tim := <-timer.C:
			if time.Since(tim) > time.Duration(setup.Set.StepPudge)*time.Second {
				logger.Info.Println("Добавьте время для обновления БД")
			}
			setStatusCross()
			_ = saveDBase()
		case <-context.Done():
			Works = false
			// for _, d := range controllers {
			// 	if d.IsConnected() {

			// 		ChanLog <- LogRecord{ID: d.ID, Type: 1, Time: time.Now(), Journal: UserDeviceStatus("Останов сервера", -2, 0)}
			// 	}
			// }
			time.Sleep(5 * time.Second)
			_ = saveDBase()
			logger.Info.Println("Останов обновления БД")
			time.Sleep(5 * time.Second)
			return
		}
	}
}
