package controller

import (
	"database/sql"
	"fmt"
	"runtime"
	"sync"
	"time"

	"github.com/ruraomsk/ag-server/controller/device"
	"github.com/ruraomsk/ag-server/controller/gui"
	"github.com/ruraomsk/ag-server/extcon"
	"github.com/ruraomsk/ag-server/logger"
	"github.com/ruraomsk/ag-server/pudge"
	"github.com/ruraomsk/ag-server/setup"

	_ "github.com/lib/pq"
)

// Имитатор котроллеров

//Devs список всех устройств

//Dmutex мютекс для всех устройств
var Dmutex sync.Mutex

func restartDevice() {
	time.Sleep(120 * time.Second)
	for _, d := range device.Devs {
		d.Mutex.Lock()
		if !d.Status {
			logger.Info.Println("Перезапускаем ", d.ID)
			go d.StartDevice()
		}
		d.Mutex.Unlock()
	}
}
func getController(id int, rq chan int, ans chan string) *pudge.Controller {
	//Вначале проверим на pudge
	ctrl := new(pudge.Controller)
	c, is := pudge.GetController(id)
	if !is {
		//Нет на pudge теперь надо проверить среди регистрированных
		rq <- id
		key := <-ans
		if len(key) == 0 {
			return nil
		}
		pudge.SetDefault(ctrl, key)
		return ctrl
	}
	ctrl = c
	ctrl.Base = true
	ctrl.CK = 1
	ctrl.NK = 1
	ctrl.PK = 1

	return ctrl
}

//Start Запуск имитатора контроллеров
func Start(context *extcon.ExtContext, rq chan int, ans chan string) {
	runtime.GOMAXPROCS(runtime.NumCPU())
	logger.Info.Println("Start controller work...")
	fmt.Println("Controller start work...")
	device.Devs = make(map[int]*device.Device)
	dbinfo := fmt.Sprintf("host=%s user=%s password=%s dbname=%s sslmode=disable",
		setup.Set.DataBase.Host, setup.Set.DataBase.User,
		setup.Set.DataBase.Password, setup.Set.DataBase.DBname)

	conDevGis, err := sql.Open("postgres", dbinfo)
	if err != nil {
		logger.Error.Printf("Запрос на открытие %s %s", dbinfo, err.Error())
		return
	}
	defer conDevGis.Close()
	w := "select idevice from public.\"cross\";"
	rows, err := conDevGis.Query(w)
	if err != nil {
		logger.Error.Println(err.Error())
		return
	}
	defer rows.Close()
	for !pudge.Works {
		time.Sleep(1 * time.Second)
	}
	count := 0
	for rows.Next() {
		// if count > 501 {
		// 	break
		// }
		dev := new(device.Device)
		rows.Scan(&dev.ID)
		if dev.ID == 496932 {
			//Real!!
			continue
		}
		dev.Controller = getController(dev.ID, rq, ans)
		if dev.Controller == nil {
			logger.Error.Printf("нет такого ID на перекрестках %d", dev.ID)
			continue
		}
		dev.CK = 1
		dev.NK = 1
		dev.PK = 1
		device.Devs[dev.ID] = dev
		go dev.StartDevice()
		count++
	}
	logger.Info.Println("Запущены имитаторы...")
	conDevGis.Close()
	go restartDevice()
	p, _ := extcon.NewContext("gui")
	go gui.Start(p)
	select {
	case <-context.Done():
		logger.Info.Println("Controller exit working...")
		fmt.Println("Controller exit working...")
		return
	}

}
