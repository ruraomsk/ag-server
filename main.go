package main

import (
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
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

//Секция инициализации программы
func init() {
	setup.Set = new(setup.Setup)
	if _, err := toml.DecodeFile("config.toml", &setup.Set); err != nil {
		fmt.Println("Can't load config file - ", err.Error())
	}

}
func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	path := setup.Set.Home
	err = logger.Init(path + "/log/ag-server")
	if err != nil {
		fmt.Println("Error opening logger subsystem ", err.Error())
		return
	}
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

	fmt.Println("\nExit ag-server working...")
}
