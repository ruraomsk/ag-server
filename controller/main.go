package controller

import (
	"fmt"
	"os"
	"runtime"
	"rura/ag-server/extcon"
	"rura/ag-server/setup"
	"rura/controller/gui"
	"rura/teprol/logger"
	"time"
)

// Имитатор котроллеров

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
	setup.LoadSetUp(path + "setup/setup_ctrl.json")
	logger.Info.Println("Start work...")
	stop := make(chan int)
	extcon.BackgroundInit()
	p, _ := extcon.NewContext("gui")
	go gui.Start(p, stop)

	extcon.BackgroundWork(time.Duration(1*time.Second), stop)
	logger.Info.Println("Exit working...")
	setup.WriteSetUp()
}
