package pudge

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/ruraomsk/ag-server/extcon"
	"github.com/ruraomsk/ag-server/logger"
	"github.com/ruraomsk/ag-server/setup"
	"math/rand"
	"strconv"
	"sync"
	"time"

	_ "github.com/lib/pq"
)

var mutex sync.Mutex
var controllers map[int]*Controller
var crosses map[string]*Cross

//Works флаг готовности pudge
var Works bool

var conDBSave *sql.DB
var conCross *sql.DB
var dbinfo string
var err error

//GetCross возвращает копию перекрестка
func GetCross(region, area, id int) (Cross, bool) {
	mutex.Lock()
	defer mutex.Unlock()
	reg := Region{Region: region, Area: area, ID: id}
	c, is := crosses[reg.ToKey()]
	return *c, is
}

//GetCrosses возвращает все перекрестки
func GetCrosses() []Region {
	mutex.Lock()
	defer mutex.Unlock()
	r := make([]Region, 0)
	for _, cr := range crosses {
		reg := Region{Region: cr.Region, Area: cr.Area, ID: cr.ID}
		r = append(r, reg)
	}
	return r
}
func getNameCross(idevice int) string {
	for _, c := range crosses {
		if c.IDevice == idevice {
			return c.Name
		}
	}
	return ""
}

//GetController возвращает копию Контроллера
func GetController(id int) (*Controller, bool) {
	mutex.Lock()
	defer mutex.Unlock()
	c, is := controllers[id]
	return c, is
}

//SetCrossNewDevice деляет новую привязку контроллера
func SetCrossNewDevice(reg Region, idevice int) error {
	mutex.Lock()
	defer mutex.Unlock()
	c, is := crosses[reg.ToKey()]
	if !is {
		return fmt.Errorf("нет такого перекрестка %v", reg)
	}
	c.IDevice = idevice
	c.WriteToDB = true
	crosses[reg.ToKey()] = c
	return nil
}

//SetCross обновляет состояние перекрестка
func SetCross(c *Cross) {
	mutex.Lock()
	defer mutex.Unlock()
	reg := Region{Region: c.Region, Area: c.Area, ID: c.ID}
	c.WriteToDB = true
	crosses[reg.ToKey()] = c
	return
}

//SetController Записывает новое состояние контроллера и если есть изменения то записывает его в лог
func SetController(c *Controller) {
	mutex.Lock()
	defer mutex.Unlock()
	insert := false
	_, is := controllers[c.ID]
	if !is {
		insert = true
		controllers[c.ID] = c
	}
	js, _ := json.Marshal(c)
	if insert {
		c.WriteToDB = false
		controllers[c.ID] = c
		w := "insert into " + setup.Set.Pudge.TableSave + " (id,device) values(" + strconv.Itoa(c.ID) + ",'" + string(js) + "');"
		_, err := conDBSave.Exec(w)
		if err != nil {
			logger.Error.Printf("For insert to controller %d %s", c.ID, err.Error())
			return
		}
	} else {
		c.WriteToDB = true
		controllers[c.ID] = c
	}
}

//Start главная процедура управления состоянием котроллеров
func Start(context *extcon.ExtContext, stop chan int, rq chan int, ans chan string) {
	// Создаем каналы и переменные
	rand.Seed(int64(1234))
	Works = false
	defer mutex.Unlock()
	controllers = make(map[int]*Controller)
	crosses = make(map[string]*Cross)
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
	defer conDBSave.Close()
	if err = conDBSave.Ping(); err != nil {
		logger.Error.Printf("Ping %s", err.Error())
		return
	}
	conCross, err = sql.Open("postgres", dbinfo)
	if err != nil {
		logger.Error.Printf("Запрос на открытие %s %s", dbinfo, err.Error())
		stop <- 1
		return
	}
	defer conCross.Close()
	err = loadDBase()
	if err != nil {
		logger.Error.Printf("save %s", err.Error())
		stop <- 1
		return
	}
	Works = true
	timer := extcon.SetTimerClock(time.Duration(setup.Set.Pudge.StepSave) * time.Second)
	for true {
		select {
		case tim := <-timer.C:
			if time.Now().Sub(tim) > time.Duration(setup.Set.Pudge.StepSave)*time.Second {
				logger.Info.Println("Добавьте время для обновления БД")
			}
			// logger.Info.Println("timer")
			setStatusCross()
			saveDBase()
		case <-context.Done():
			saveDBase()
			return
		case id := <-rq:
			ans <- isRegistred(id)
		}
	}
}
func toReturnControllers(mgs []int) {
	var ret Controllers
	ret.Contrs = make([]Controller, 0)
	mutex.Lock()
	for _, i := range mgs {
		ret.Contrs = append(ret.Contrs, *controllers[i])
	}
	mutex.Unlock()
}
