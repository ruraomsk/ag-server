package main

import (
	"database/sql"
	"fmt"
	"os"
	"runtime"
	"rura/ag-server/controller/device"
	"rura/ag-server/extcon"
	"rura/ag-server/logger"
	"rura/ag-server/setup"
	"sync"
	"time"
)

// Имитатор котроллеров
var devs map[int]*device.Device
var mutex sync.Mutex

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	path, err := os.UserHomeDir()
	if err != nil {
		fmt.Println("Error opening system ", err.Error())
		return
	}
	err = logger.Init(path + "/log/controller")
	if err != nil {
		fmt.Println("Error opening logger subsystem ", err.Error())
		return
	}
	setup.LoadSetUp(path + "/setup/setup_ag.json")
	logger.Info.Println("Start work...")
	fmt.Println("Controller start work...")
	devs = make(map[int]*device.Device)
	dbinfo := fmt.Sprintf("host=%s user=%s password=%s dbname=%s sslmode=disable",
		setup.Set.DataBase.Host, setup.Set.DataBase.User,
		setup.Set.DataBase.Password, setup.Set.DataBase.DBname)

	conDevGis, err := sql.Open("postgres", dbinfo)
	if err != nil {
		logger.Error.Printf("Запрос на открытие %s %s", dbinfo, err.Error())
		return
	}
	defer conDevGis.Close()
	w := "select id from dev_gis;"
	rows, err := conDevGis.Query(w)
	if err != nil {
		logger.Error.Println(err.Error())
		return
	}
	defer rows.Close()
	for rows.Next() {
		dev := new(device.Device)
		rows.Scan(&dev.ID)
		devs[dev.ID] = dev
		go dev.StartDevice()
	}
	logger.Info.Println("Запущены имитаторы...")

	stop := make(chan int)
	extcon.BackgroundInit()
	// p, _ := extcon.NewContext("gui")
	// go gui.Start(p, stop)

	extcon.BackgroundWork(time.Duration(1*time.Second), stop)
	logger.Info.Println("Exit working...")
	fmt.Println("Controller exit working...")
}
