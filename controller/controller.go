package controller

import (
	"database/sql"
	"fmt"
	"runtime"
	"rura/ag-server/controller/device"
	"rura/ag-server/controller/gui"
	"rura/ag-server/extcon"
	"rura/ag-server/logger"
	"rura/ag-server/pudge"
	"rura/ag-server/setup"
	"sync"
	"time"

	_ "github.com/lib/pq"
)

// Имитатор котроллеров

//Devs список всех устройств

//Dmutex мютекс для всех устройств
var Dmutex sync.Mutex

func restartDevice() {
	time.Sleep(60 * time.Second)
	for _, d := range device.Devs {
		d.Mutex.Lock()
		if !d.Status {
			logger.Info.Println("Перезапускаем ", d.ID)
			go d.StartDevice()
		}
		d.Mutex.Unlock()
	}
}

//Start Запуск имитатора контроллеров
func Start(context *extcon.ExtContext) {
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
	// count := 0
	for rows.Next() {
		dev := new(device.Device)
		rows.Scan(&dev.ID)
		c, is := pudge.GetController(dev.ID)
		if !is {
			device.SetDefault(&c)
			c.ID = dev.ID
		}
		dev.Controller = &c
		device.Devs[dev.ID] = dev
		go dev.StartDevice()
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
