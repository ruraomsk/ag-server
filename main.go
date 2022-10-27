package main

import (
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/ruraomsk/ag-server/cameras"
	"github.com/ruraomsk/ag-server/comm"
	"github.com/ruraomsk/ag-server/debug"
	"github.com/ruraomsk/ag-server/dumper"
	"github.com/ruraomsk/ag-server/loader"
	"github.com/ruraomsk/ag-server/logsys"
	"github.com/ruraomsk/ag-server/pudge"
	"github.com/ruraomsk/ag-server/sqlsave"
	"github.com/ruraomsk/ag-server/svgsave"
	"github.com/ruraomsk/ag-server/xcontrol"

	"github.com/BurntSushi/toml"
	"github.com/ruraomsk/ag-server/creator"
	"github.com/ruraomsk/ag-server/extcon"
	"github.com/ruraomsk/ag-server/logger"
	"github.com/ruraomsk/ag-server/setup"

	//pprof init

	_ "net/http/pprof"
)

var err error

func recoveryPanic() {
	if recoveryMessage := recover(); recoveryMessage != nil {
		logger.Error.Printf("PANIC:%v", recoveryMessage)
		os.Exit(-1)
	}
	// logger.Error.Println("Паники нет просто выход!")
	// os.Exit(0)
}

// Секция инициализации программы
func init() {
	setup.Set = new(setup.Setup)
	if _, err := toml.DecodeFile("config.toml", &setup.Set); err != nil {
		fmt.Println("Can't load config file - ", err.Error())
	}
}
func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	defer recoveryPanic()
	path := setup.Set.Home
	err = logger.Init(path + "/log")
	if err != nil {
		fmt.Println("Error opening logger subsystem ", err.Error())
		return
	}
	if len(os.Args) > 1 {
		if strings.Compare(os.Args[1], "create") == 0 {
			err = creator.Start(path)
			if err != nil {
				fmt.Println(err.Error())
			}
			return
		}
		if strings.Compare(os.Args[1], "createSVG") == 0 {
			err = creator.SvgCreator()
			if err != nil {
				fmt.Println(err.Error())
			}
			return
		}

		if strings.Compare(os.Args[1], "save") == 0 {
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
		if strings.Compare(os.Args[1], "update") == 0 {
			if len(os.Args[2]) == 0 || len(os.Args[3]) == 0 {
				fmt.Println("Нужно запускать с параметрами номер региона имя файла копии базы")
				return
			}
			err = creator.Update(os.Args[2], os.Args[3])
			if err != nil {
				fmt.Println(err.Error())
			}
			return
		}
		if strings.Compare(os.Args[1], "move") == 0 {
			if len(os.Args[2]) == 0 || len(os.Args[3]) == 0 {
				fmt.Println("Нужно запускать с параметрами номер региона имя файла перестановок")
				return
			}
			region, _ := strconv.Atoi(os.Args[2])
			creator.Mover(region, os.Args[3])
			return
		}

	}
	logger.Info.Println("Start ag-server work...")
	fmt.Println("Start ag-server work...")

	extcon.BackgroundInit()

	stop := make(chan interface{})
	ready := make(chan interface{})
	if setup.Set.Version == 0 {
		setup.Set.Version = 1
	}
	switch setup.Set.Version {
	case 1:
		go pudge.Start(stop)
		go comm.StartListen()
		go comm.Start(stop)
	case 2:
		// go memDB.Start(ready, stop)
		// <-ready
		// go techComm.StartListen(ready)
		// <-ready
		// go techComm.Start(ready)
		// <-ready
	default:
		fmt.Printf("Неверный номер версии программы %d", setup.Set.Version)
		return
	}
	fmt.Println("Start pudge connection...")

	if setup.Set.LogSystem.Make {
		go logsys.Start()
	}
	fmt.Println("Start logsystem...")

	if setup.Set.Camera.Make {
		go cameras.CamerasStart(setup.Set.Camera.Path)
	}

	if setup.Set.XCtrl.Switch {
		go xcontrol.Start(ready, stop)
		<-ready
	}
	fmt.Println("Start xctrl...")
	if setup.Set.Dumper.Make {
		go dumper.Start()
	}
	fmt.Println("Start dumper...")

	// if setup.Set.Statistic.Make {
	// 	go dumper.Statistics()
	// }

	if setup.Set.Loader.Make {
		go loader.RemoteLoader()
	}
	fmt.Println("Start loader...")
	if setup.Set.Saver.Make {
		go sqlsave.Start()
		go svgsave.Start()
	}
	fmt.Println("Start saver...")
	go debug.Run()
	fmt.Println("Start debugger...")
	extcon.BackgroundWork(time.Duration(1*time.Second), stop)
	logger.Info.Println("Exit ag-server working...")
	fmt.Println("\nExit ag-server working...")
}
