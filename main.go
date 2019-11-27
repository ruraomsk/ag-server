package main

import (
	"fmt"
	"os"
	"runtime"
	"rura/ag-server/extcon"
	"rura/ag-server/logger"
	"rura/ag-server/pudge"
	"rura/ag-server/setup"
	"time"
)

var err error

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	path, err := os.UserHomeDir()
	if err != nil {
		fmt.Println("Error opening system ", err.Error())
		return
	}
	err = logger.Init(path + "/log/ag-server")
	if err != nil {
		fmt.Println("Error opening logger subsystem ", err.Error())
		return
	}
	logger.Info.Println("Start work...")
	fmt.Println("Start work...")
	err = setup.LoadSetUp(path + "/setup/setup_ag.json")
	if err != nil {
		fmt.Printf("Ошибки в настройке %s", err.Error())
		return
	}
	stop := make(chan int)
	extcon.BackgroundInit()
	p, _ := extcon.NewContext("pudge")
	go pudge.Start(p, stop)

	extcon.BackgroundWork(time.Duration(1*time.Second), stop)
	logger.Info.Println("Exit working...")

	setup.WriteSetUp()
	fmt.Println("\nExit working...")
}
