package main

import (
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/ruraomsk/ag-server/comm"
	"github.com/ruraomsk/ag-server/controller"
	"github.com/ruraomsk/ag-server/creator"
	"github.com/ruraomsk/ag-server/extcon"
	"github.com/ruraomsk/ag-server/inspect"
	"github.com/ruraomsk/ag-server/logger"
	"github.com/ruraomsk/ag-server/pudge"
	"github.com/ruraomsk/ag-server/setup"
)

var err error

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	path, err := os.UserHomeDir()
	if runtime.GOOS != "linux" {
		path = "D:/asud"
	}
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
				fmt.Println(err.Error())
			}
			return
		}
		if strings.Contains(os.Args[1], "save") {
			err = creator.SaveAll(path)
			if err != nil {
				fmt.Println(err.Error())
			}
			return
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
	if len(os.Args) > 1 {
		if strings.Contains(os.Args[1], "simul") {
			time.Sleep(5 * time.Second)
			c, _ := extcon.NewContext("controller")
			go controller.Start(c, rq, ans)
		}
	}
	i, _ := extcon.NewContext("inspector")
	go inspect.Start(i, stop)

	extcon.BackgroundWork(time.Duration(1*time.Second), stop)
	logger.Info.Println("Exit ag-server working...")

	setup.WriteSetUp()
	fmt.Println("\nExit ag-server working...")
}
