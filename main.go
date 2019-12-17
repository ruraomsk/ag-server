package main

import (
	"fmt"
	"os"
	"runtime"
	"rura/ag-server/comm"
	"rura/ag-server/controller"
	"rura/ag-server/creator"
	"rura/ag-server/extcon"
	"rura/ag-server/logger"
	"rura/ag-server/pudge"
	"rura/ag-server/setup"
	"strings"
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
	err = setup.LoadSetUp(path + "/setup/setup_ag.json")
	if len(os.Args) > 1 {
		if strings.Contains(os.Args[1], "create") {
			err = creator.Start(path)
			if err != nil {
				return
			}
		}
	}
	logger.Info.Println("Start ag-server work...")
	fmt.Println("Start ag-server work...")
	if err != nil {
		fmt.Printf("Ошибки в настройке %s", err.Error())
		return
	}
	stop := make(chan int)
	extcon.BackgroundInit()
	p, _ := extcon.NewContext("pudge")
	rq := make(chan int)
	ans := make(chan string)
	go pudge.Start(p, stop, rq, ans)
	go comm.StartListen(stop, rq, ans)
	time.Sleep(5 * time.Second)
	c, _ := extcon.NewContext("controller")
	go controller.Start(c, rq, ans)
	extcon.BackgroundWork(time.Duration(1*time.Second), stop)
	logger.Info.Println("Exit ag-server working...")

	setup.WriteSetUp()
	fmt.Println("\nExit ag-server working...")
}
