package main

import (
	"fmt"
	"github.com/ruraomsk/ag-server/dumper"
	"github.com/ruraomsk/ag-server/loader"
	"github.com/ruraomsk/ag-server/sqlsave"
	"github.com/ruraomsk/ag-server/svgsave"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/ruraomsk/TLServer/logger"
	"github.com/ruraomsk/ag-server/comm"
	"github.com/ruraomsk/ag-server/controller"
	"github.com/ruraomsk/ag-server/creator"
	"github.com/ruraomsk/ag-server/extcon"
	"github.com/ruraomsk/ag-server/inspect"
	"github.com/ruraomsk/ag-server/pudge"
	"github.com/ruraomsk/ag-server/setup"
	"github.com/ruraomsk/ag-server/xcontrol"

	//pprof init

	_ "net/http/pprof"
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
	err = logger.Init(path + "/log")
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
			if len(os.Args[2]) == 0 {
				fmt.Println("Нужно запускать с параметром all для всех регионов или указать код региона")
				return
			}
			err = creator.SaveAll(path+"/save", os.Args[2])
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
	go pudge.Start(p, stop)
	go comm.StartListen()
	if len(os.Args) > 1 {
		if strings.Contains(os.Args[1], "simul") {
			c, _ := extcon.NewContext("controller")
			go controller.Start(c)
		}
	}
	i, _ := extcon.NewContext("inspector")
	go inspect.Start(i, stop)
	x, _ := extcon.NewContext("xcontrol")
	go xcontrol.Start(x, stop)
	go dumper.Start()
	go dumper.Statistics()
	go loader.RemoteLoader()
	if setup.Set.Saver.Make {
		go sqlsave.Start()
		go svgsave.Start()
	}
	extcon.BackgroundWork(time.Duration(1*time.Second), stop)
	logger.Info.Println("Exit ag-server working...")
	fmt.Println("\nExit ag-server working...")
}
