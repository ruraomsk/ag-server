package main

import (
	"fmt"
	"os"
	"runtime"
	"rura/ag-server/logger"
	"rura/ag-server/setup"
	"rura/creator/create"
)

//Создает все таблицы в постгресс
//Дополнительно производит загрузку из внешних текстовых файлов
//Если таковые будут))
var err error

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	path, err := os.UserHomeDir()
	if err != nil {
		fmt.Println("Error opening system ", err.Error())
		return
	}
	err = logger.Init(path + "/log/creator")
	if err != nil {
		fmt.Println("Error opening logger subsystem ", err.Error())
		return
	}
	logger.Info.Println("Start work...")
	setup.LoadSetUp(path + "/setup/setup_ag.json")
	err = create.SQLCreate(path + "/setup")
	if err != nil {
		fmt.Println("Найдены ошибки проверьте log file")
		return
	}
	logger.Info.Println("Exit working...")
	setup.WriteSetUp()
}
