package techComm

import (
	"github.com/ruraomsk/TLServer/logger"
	"github.com/ruraomsk/ag-server/memDB"
	"sync"
	"time"
)

// Пакет управления СФДК
var mapDevices = make(map[int]int)
var dMutex sync.Mutex

func sfdkControl() {
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
		memDB.TableDevices.Lock()
		for _, c := range memDB.GetListControllers() {
			ctrl, err := memDB.GetController(c)
			if err != nil {
				continue
			}
			if !ctrl.StatusConnection {
				continue
			}
			dMutex.Lock()
			count, is := mapDevices[c]
			dMutex.Unlock()
			if !is {
				if ctrl.StatusCommandDU.IsReqSFDK1 && ctrl.StatusConnection {
					//logger.Info.Printf("нет в контроле но есть команда id %d count %d", ctrl.ID,count)
					setOffSfdk(ctrl.ID)
					memDB.SetController(ctrl)
				}
				continue
			}
			if count == 0 && ctrl.StatusCommandDU.IsReqSFDK1 && ctrl.StatusConnection {
				//logger.Info.Printf("есть в контроле и есть команда id %d count %d", ctrl.ID,count)
				setOffSfdk(ctrl.ID)
				ctrl.StatusCommandDU.IsReqSFDK1 = false
				memDB.SetController(ctrl)
				continue
			}
			if count != 0 && !ctrl.StatusCommandDU.IsReqSFDK1 && ctrl.StatusConnection {
				//logger.Info.Printf("есть в контроле и нет команды id %d count %d", ctrl.ID,count)
				setOnSfdk(ctrl.ID)
				ctrl.StatusCommandDU.IsReqSFDK1 = true
				memDB.SetController(ctrl)
				continue
			}
		}
		memDB.TableDevices.Unlock()
	}
}
func setOnSfdk(id int) {
	//logger.Info.Printf("Включаем СФДК для %d", id)
	deckhand, err := getChanCommand(id)
	if err != nil {
		logger.Info.Printf("Отключено устройство %d", id)
		return
	}
	deckhand <- CommandARM{ID: id, Command: 4, Params: 1, User: "СФДК"}
}
func setOffSfdk(id int) {
	//logger.Info.Printf("Отключаем СФДК для %d", id)
	deckhand, err := getChanCommand(id)
	if err != nil {
		logger.Info.Printf("Отключено устройство %d", id)
		return
	}
	deckhand <- CommandARM{ID: id, Command: 4, Params: 0, User: "СФДК"}
}
