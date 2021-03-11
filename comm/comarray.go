package comm

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/ruraomsk/TLServer/logger"
	"github.com/ruraomsk/ag-server/extcon"
	"github.com/ruraomsk/ag-server/pudge"
	"github.com/ruraomsk/ag-server/setup"
)

//DevPhases для передачи фаз
type DevPhases struct {
	ID int      `json:"idevice"`
	DK pudge.DK `json:"dk"`
}

var countSenders = 0
var connectMap map[string]net.Conn
var cMutex sync.Mutex

func listenArmCommand() {
	ln, err := net.Listen("tcp", ":"+strconv.Itoa(setup.Set.CommServer.PortCommand))
	if err != nil {
		logger.Error.Printf("Ошибка открытия порта %s", err.Error())
		return
	}
	//defer ln.Close()
	for {
		socket, err := ln.Accept()
		if err != nil {
			logger.Error.Printf("Ошибка accept %s", err.Error())
			continue
		}
		go workerCommand(socket)
	}
}
func listenChangeProtocol() {
	ln, err := net.Listen("tcp", ":"+strconv.Itoa(setup.Set.CommServer.PortProtocol))
	if err != nil {
		logger.Error.Printf("Ошибка открытия порта %s", err.Error())
		return
	}
	//defer ln.Close()
	for {
		socket, err := ln.Accept()
		if err != nil {
			logger.Error.Printf("Ошибка accept %s", err.Error())
			continue
		}
		go workerProtocol(socket)
	}
}

func listenArmArray() {
	ln, err := net.Listen("tcp", ":"+strconv.Itoa(setup.Set.CommServer.PortArray))

	if err != nil {
		logger.Error.Printf("Ошибка открытия порта %s", err.Error())
		return
	}
	//defer ln.Close()
	for {
		socket, err := ln.Accept()
		if err != nil {
			logger.Error.Printf("Ошибка accept %s", err.Error())
			continue
		}
		go workerArray(socket)
	}
}
func listenSendingPhazes() {
	ln, err := net.Listen("tcp", ":"+strconv.Itoa(setup.Set.CommServer.PortDevices))
	connectMap = make(map[string]net.Conn)
	go workerDevices()
	if err != nil {
		logger.Error.Printf("Ошибка открытия порта %s", err.Error())
		return
	}
	//defer ln.Close()
	for {
		socket, err := ln.Accept()
		if err != nil {
			logger.Error.Printf("Ошибка accept %s", err.Error())
			continue
		}
		logger.Info.Printf("Новый клиент фаз устройства %s", socket.RemoteAddr().String())
		cMutex.Lock()
		connectMap[socket.RemoteAddr().String()] = socket
		cMutex.Unlock()
	}
}
func workerDevices() {
	timer := extcon.SetTimerClock(time.Duration(10 * time.Second))
	// writer := bufio.NewWriter(soc)
	for {
		select {
		case <-timer.C:
			for _, soc := range connectMap {
				_, err := fmt.Fprintf(soc, "0\n")
				if err != nil {
					logger.Error.Printf("Ошибка передачи tcp %s %s", soc.RemoteAddr().String(), err.Error())
					cMutex.Lock()
					delete(connectMap, soc.RemoteAddr().String())
					cMutex.Unlock()
				}
			}
		case d := <-sendPhases:
			array, err := json.Marshal(&d)
			if err != nil {
				logger.Error.Printf("Ошибка json %s", err.Error())
				continue
			}
			for _, soc := range connectMap {
				_, err = fmt.Fprintf(soc, string(array)+"\n")
				if err != nil {
					logger.Error.Printf("Ошибка передачи tcp %s %s", soc.RemoteAddr().String(), err.Error())
					cMutex.Lock()
					delete(connectMap, soc.RemoteAddr().String())
					cMutex.Unlock()
				}
			}
		}
	}

}
func workerCommand(soc net.Conn) {
	defer soc.Close()
	var command CommandARM
	logger.Info.Printf("Новый клиент комманд %s", soc.RemoteAddr().String())
	reader := bufio.NewReader(soc)
	for {
		c, err := reader.ReadString('\n')
		if err != nil {
			logger.Error.Printf("При чтении команд сервера АРМ %s $s ", soc.RemoteAddr().String(), err.Error())
			return
		}
		if c[0:1] == "0" {
			// logger.Info.Println("Keep alive")
			continue
		}
		err = json.Unmarshal([]byte(c), &command)
		if err != nil {
			logger.Error.Println("При конвератации команд сервера АРМ ", err.Error())
			continue
		}
		dev, is := devs[command.ID]
		if !is {
			if strings.Compare(command.User, setup.Set.XCtrl.NameUser) != 0 {
				logger.Error.Printf("Команда сервера АРМ %v нет такого id %d", command, command.ID)
			}
			continue
		}
		logger.Info.Printf("Команда %v %s", command, soc.RemoteAddr().String())
		if command.Command == 1 {
			//Принудительная отправка всех массивов
			w := fmt.Sprintf(" %s  заказал перезагрузку всех массивов", command.User)
			pudge.ChanLog <- pudge.RecLogCtrl{ID: command.ID, Type: -1, Time: time.Now(), LogString: w}
			ctrl, _ := pudge.GetController(command.ID)
			ctrl.Arrays = make([]pudge.ArrayPriv, 0)
			pudge.SetController(ctrl)
			logger.Info.Printf("id %d массив привязок поставлен на перезагрузку", command.ID)
		} else {
			if command.Command != 4 {
				w := fmt.Sprintf("%s  %s", command.User, getDescription(command))
				pudge.ChanLog <- pudge.RecLogCtrl{ID: command.ID, Type: -1, Time: time.Now(), LogString: w}
			}
			dev.CommandARM <- command
		}
	}
}
func workerArray(soc net.Conn) {
	defer soc.Close()
	var state pudge.UserCross
	logger.Info.Printf("Новый клиент массивов %s", soc.RemoteAddr().String())
	reader := bufio.NewReader(soc)
	for {
		a, err := reader.ReadString('\n')
		if err != nil {
			logger.Error.Println("При чтении привязки от сервера АРМ ", err.Error())
			return
		}
		//fmt.Printf("=%v=",a)
		if a[0:1] == "0" {
			// logger.Info.Println("Keep alive")
			continue
		}

		err = json.Unmarshal([]byte(a), &state)
		if err != nil {
			logger.Error.Printf("При конвератации привязки сервера АРМ %s %s", a, err.Error())
			continue
		}
		// logger.Error.Println("Пришло state")
		if state.State.IDevice < 0 {
			//Удаление перекрестка
			last, is := pudge.GetCross(state.State.Region, state.State.Area, state.State.ID)
			if !is {
				//Перекрестка нет
				logger.Info.Printf("Попытка удаления удаленного перекрестка %d %d %d", state.State.Region, state.State.Area, state.State.ID)
				continue
			}
			logger.Info.Printf("Удаление перекрестка %d %d %d %d", state.State.Region, state.State.Area, state.State.ID, last.IDevice)
			w := fmt.Sprintf("%s удаление перекрестка %d %d %d ", state.User, state.State.Region, state.State.Area, state.State.ID)
			pudge.ChanLog <- pudge.RecLogCtrl{ID: last.IDevice, Type: 0, Time: time.Now(), LogString: w}
			time.Sleep(1 * time.Second)
			pudge.DeleteCross(state.State.Region, state.State.Area, state.State.ID)
			continue
		}
		old, is := pudge.GetCross(state.State.Region, state.State.Area, state.State.ID)
		if !is {
			//Перекрестка нет нужно создать
			logger.Info.Printf("Добавлен перекресток %d %d %d", state.State.Region, state.State.Area, state.State.ID)
			state.State.StatusDevice = 18
			w := fmt.Sprintf(" %s добавил перекрестка %d %d %d %d", state.User, state.State.Region, state.State.Area, state.State.ID, state.State.IDevice)
			logger.Info.Print(w)
			pudge.SetCross(&state.State)
			time.Sleep(1 * time.Second)
			pudge.ChanLog <- pudge.RecLogCtrl{ID: state.State.IDevice, Type: 0, Time: time.Now(), LogString: w}
			continue
		}
		// logger.Info.Printf("Write new state ")
		if old.IDevice != state.State.IDevice {
			logger.Info.Printf("Отключаем старое устройство %d ", old.IDevice)
			d, ok := devs[old.IDevice]
			if ok {
				//Остановим текущее
				d.ExitCommand <- 1
				delete(devs, old.IDevice)
			}
		}
		pudge.SetCross(&state.State)
		w := fmt.Sprintf("%s изменил перекресток %d %d %d", state.User, state.State.Region, state.State.Area, state.State.ID)
		logger.Info.Print(w)
		pudge.ChanLog <- pudge.RecLogCtrl{ID: state.State.IDevice, Type: 0, Time: time.Now(), LogString: w}
	}
}
func workerProtocol(soc net.Conn) {
	defer soc.Close()
	var protocol ChangeProtocol
	logger.Info.Printf("Новый клиент протокола %s", soc.RemoteAddr().String())
	reader := bufio.NewReader(soc)
	for {
		c, err := reader.ReadString('\n')
		if err != nil {
			logger.Error.Println("При чтении изменения протокола от АРМ ", err.Error())
			return
		}
		if c[0:1] == "0" {
			// logger.Info.Println("Keep alive")
			continue
		}
		err = json.Unmarshal([]byte(c), &protocol)
		if err != nil {
			logger.Error.Println("При конвертации изменения протокола АРМ ", err.Error())
			continue
		}
		dev, is := devs[protocol.ID]
		if !is {
			logger.Error.Printf("Команда протокола АРМ %v нет такого id %d", protocol, protocol.ID)
			continue
		}
		w := fmt.Sprintf("%s send command %v", protocol.User, protocol)
		logger.Info.Print(w)
		pudge.ChanLog <- pudge.RecLogCtrl{ID: protocol.ID, Type: 1, Time: time.Now(), LogString: w}
		dev.ChangeProtocol <- protocol
	}
}

func getDescription(toSend CommandARM) string {
	switch toSend.Command {
	case 4:
		if toSend.Params == 1 {
			return "Отправлен запрос на смену фаз"
		}
		return "Отключить запрос на смену фаз"
	case 5:
		if toSend.Params == 0 {
			return "Переход на автоматическое регулирование ПК"
		}
		return "Сменить ПК на " + strconv.Itoa(toSend.Params)
	case 6:
		if toSend.Params == 0 {
			return "Переход на автоматическое регулирование СК"
		}
		return "Сменить CК на " + strconv.Itoa(toSend.Params)
	case 7:
		if toSend.Params == 0 {
			return "Переход на автоматическое регулирование НК"
		}
		return "Сменить НК на " + strconv.Itoa(toSend.Params)
	}
	switch toSend.Params {
	case 0:
		return "Переход в Локальный режим"
	case 9:
		return "Переход в  Координированное управление"
	case 10:
		return "Включить жёлтое мигание"
	case 11:
		return "Отключить светофоры"
	}
	return "Включить фазу №" + strconv.Itoa(toSend.Params)
}
