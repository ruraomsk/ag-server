package pudge

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/ruraomsk/TLServer/logger"
	"github.com/ruraomsk/ag-server/extcon"
	"github.com/ruraomsk/ag-server/setup"

	//Инициализатор постргресса
	_ "github.com/lib/pq"
)

//var mutexCross sync.Mutex
var mutexCtrl sync.Mutex
var controllers map[int]*Controller
var crosses map[string]*Cross
var statuses map[int]string
var controls map[int]bool
var nowstatus map[string]string

//Works флаг готовности pudge
var Works bool

//ChanLog канал приема сообщений логов устройств
var ChanLog chan RecLogCtrl
var conDBSave *sql.DB
var conCross *sql.DB
var conLog *sql.DB

var Reload chan interface{}
var dbinfo string
var err error

func Lock() {
	mutexCtrl.Lock()
}
func Unclock() {
	mutexCtrl.Unlock()

}

//GetCrossLessLock возвращает копию перекрестка
func GetCrossLessLock(region, area, id int) (*Cross, bool) {
	reg := Region{Region: region, Area: area, ID: id}
	c, is := crosses[reg.ToKey()]
	if !is {
		return nil, is
	}
	return c, is
}

//GetCross возвращает копию перекрестка
func GetCross(region, area, id int) (Cross, bool) {
	mutexCtrl.Lock()
	defer mutexCtrl.Unlock()
	reg := Region{Region: region, Area: area, ID: id}
	c, is := crosses[reg.ToKey()]
	if !is {
		cc := new(Cross)
		return *cc, is
	}
	return *c, is
}

//GetCrosses возвращает все перекрестки
func GetCrosses() []Region {
	mutexCtrl.Lock()
	defer mutexCtrl.Unlock()
	r := make([]Region, 0)
	for _, cr := range crosses {
		reg := Region{Region: cr.Region, Area: cr.Area, ID: cr.ID}
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
	mutexCtrl.Lock()
	defer mutexCtrl.Unlock()
	for _, c := range crosses {
		if c.IDevice == idevice {
			return c.Name
		}
	}
	return ""
}

//DeleteCross Удаляет перекресток
func DeleteCross(region, area, id int) {
	mutexCtrl.Lock()
	reg := Region{Region: region, Area: area, ID: id}
	delete(crosses, reg.ToKey())
	mutexCtrl.Unlock()
	w := fmt.Sprintf("DELETE FROM public.\"cross\" WHERE region=%d and area=%d and id=%d;", region, area, id)
	_, err = conCross.Exec(w)

	if err != nil {
		logger.Error.Printf("Error %s  %s\n", w, err.Error())
	}
	return
}

//GetController возвращает копию Контроллера
func GetController(id int) (*Controller, bool) {
	mutexCtrl.Lock()
	defer mutexCtrl.Unlock()
	c, is := controllers[id]
	return c, is
}

//SetCrossNewDevice деляет новую привязку контроллера
func SetCrossNewDevice(reg Region, idevice int) error {
	mutexCtrl.Lock()
	defer mutexCtrl.Unlock()
	c, is := crosses[reg.ToKey()]
	if !is {
		return fmt.Errorf("нет такого перекрестка %v", reg)
	}
	c.IDevice = idevice
	c.WriteToDB = true
	//crosses[reg.ToKey()] = c
	return nil
}

//SetCross обновляет состояние перекрестка
func SetCrossLessLock(c *Cross) {
	reg := Region{Region: c.Region, Area: c.Area, ID: c.ID}
	_, is := crosses[reg.ToKey()]
	if !is {
		c.WriteToDB = false
		crosses[reg.ToKey()] = c
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
		crosses[reg.ToKey()] = c
		logger.Debug.Printf("Записано изменение %s", reg.ToKey())
	}
	return
}

//SetCross обновляет состояние перекрестка
func SetCross(c *Cross) {
	reg := Region{Region: c.Region, Area: c.Area, ID: c.ID}
	mutexCtrl.Lock()
	defer mutexCtrl.Unlock()
	_, is := crosses[reg.ToKey()]
	if !is {
		c.WriteToDB = false
		crosses[reg.ToKey()] = c
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
		crosses[reg.ToKey()] = c
		logger.Debug.Printf("Записано изменение %s", reg.ToKey())
	}
	return
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
	//mutexCtrl.Unlock()
	if insert {
		js, _ := json.Marshal(c)
		w := "insert into devices (id,device) values(" + strconv.Itoa(c.ID) + ",'" + string(js) + "');"
		_, err := conDBSave.Exec(w)
		if err != nil {
			logger.Error.Printf("For insert to controller %d %s", c.ID, err.Error())
			return
		}
	} else {
		//mutexCtrl.Lock()
		c.WriteToDB = true
		//mutexCtrl.Unlock()
	}
	// logger.Debug.Printf("end setController %d", c.ID)
}

//Start главная процедура управления состоянием котроллеров
func Start(stop chan interface{}) {
	// Создаем каналы и переменные
	Works = false
	//defer mutexCross.Unlock()
	//defer mutexCtrl.Unlock()
	context, _ := extcon.NewContext("pudge")
	controllers = make(map[int]*Controller)
	crosses = make(map[string]*Cross)
	statuses = make(map[int]string)
	ChanLog = make(chan RecLogCtrl, 1000)
	controls = make(map[int]bool)
	nowstatus = make(map[string]string)
	Reload = make(chan interface{})
	dbinfo = fmt.Sprintf("host=%s user=%s password=%s dbname=%s sslmode=disable",
		setup.Set.DataBase.Host, setup.Set.DataBase.User,
		setup.Set.DataBase.Password, setup.Set.DataBase.DBname)
	conDBSave, err = sql.Open("postgres", dbinfo)
	if err != nil {
		logger.Error.Printf("Запрос на открытие %s %s", dbinfo, err.Error())
		stop <- 1
		return
	}
	defer conDBSave.Close()
	if err = conDBSave.Ping(); err != nil {
		logger.Error.Printf("Ping %s", err.Error())
		stop <- 1
		return
	}
	conCross, err = sql.Open("postgres", dbinfo)
	if err != nil {
		logger.Error.Printf("Запрос на открытие %s %s", dbinfo, err.Error())
		stop <- 1
		return
	}
	defer conCross.Close()
	needHistoryCross(conCross)
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
	defer conLog.Close()
	go writeLog()
	Works = true
	timer := extcon.SetTimerClock(time.Duration(setup.Set.StepPudge) * time.Second)
	for true {
		select {
		case <-Reload:
			mutexCtrl.Lock()
			err = loadDBase()
			if err != nil {
				logger.Error.Printf("load %s", err.Error())
				stop <- 1
				return
			}
			mutexCtrl.Unlock()

		case tim := <-timer.C:
			if time.Now().Sub(tim) > time.Duration(setup.Set.StepPudge)*time.Second {
				logger.Info.Println("Добавьте время для обновления БД")
			}
			setStatusCross()
			saveDBase()
		case <-context.Done():
			Works = false
			for _, d := range controllers {
				if d.IsConnected() {
					ChanLog <- RecLogCtrl{ID: d.ID, Type: -1, Time: time.Now(), LogString: "Остановлен сервер"}
				}
			}
			time.Sleep(5 * time.Second)
			saveDBase()
			logger.Info.Println("Останов обновления БД")
			time.Sleep(5 * time.Second)
			return
		}
	}
}
