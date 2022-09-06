package comm

import (
	"sync"
	"time"

	"github.com/ruraomsk/ag-server/logger"
	"github.com/ruraomsk/ag-server/pudge"
)

// Пакет управления СФДК
var mapDevices map[int]int
var dMutex sync.Mutex

func sfdkControl() {
	mapDevices = make(map[int]int)
	for {
		com := <-cSFDK
		if com.Command != 4 {
			logger.Error.Printf("Неверная команда %v", com)
			continue
		}
		//logger.Info.Printf("Пришла команда %v", com)
		dMutex.Lock()
		count, is := mapDevices[com.ID]
		if !is {
			count = 0
		}
		if com.Params < 0 {
			//logger.Info.Printf("СФ ДК отменен из Сервера Связи для %d",com.ID)
			setOffSfdk(com.ID)
			count = 0
			com.Params = 0
		}
		if com.Params == 0 {
			count--
			if count < 0 {
				count = 0
			}
		} else {
			count++
		}
		mapDevices[com.ID] = count
		//logger.Info.Printf("id %d %d count %d", com.ID,com.Params,count)
		dMutex.Unlock()
	}
}
func routeSgdk() {
	for {
		time.Sleep(1 * time.Second)
		for _, c := range pudge.GetControllers() {
			dMutex.Lock()
			count, is := mapDevices[c]
			dMutex.Unlock()
			if !is {
				ctrl, is := pudge.GetController(c)
				if !is {
					continue
				}
				if ctrl.StatusCommandDU.IsReqSFDK1 && ctrl.StatusConnection {
					//logger.Info.Printf("нет в контроле но есть команда id %d count %d", ctrl.ID,count)
					setOffSfdk(ctrl.ID)
				}
				continue
			}
			ctrl, is := pudge.GetController(c)
			if !is {
				continue
			}
			if count == 0 && ctrl.StatusCommandDU.IsReqSFDK1 && ctrl.StatusConnection {
				//logger.Info.Printf("есть в контроле и есть команда id %d count %d", ctrl.ID,count)
				setOffSfdk(ctrl.ID)
				ctrl.StatusCommandDU.IsReqSFDK1 = false
				pudge.GetController(c)
				continue
			}
			if count != 0 && !ctrl.StatusCommandDU.IsReqSFDK1 && ctrl.StatusConnection {
				//logger.Info.Printf("есть в контроле и нет команды id %d count %d", ctrl.ID,count)
				setOnSfdk(ctrl.ID)
				ctrl.StatusCommandDU.IsReqSFDK1 = true
				pudge.GetController(c)
				continue
			}
		}
	}
}
func setOnSfdk(id int) {
	//logger.Info.Printf("Включаем СФДК для %d", id)
	dev, is := getDevice(id)
	if !is {
		return
	}
	dev.CommandARM <- pudge.CommandARM{ID: id, Command: 4, Params: 1, User: "СФДК"}
}
func setOffSfdk(id int) {
	//logger.Info.Printf("Отключаем СФДК для %d", id)
	dev, is := getDevice(id)
	if !is {
		return
	}
	dev.CommandARM <- pudge.CommandARM{ID: id, Command: 4, Params: 0, User: "СФДК"}
}
