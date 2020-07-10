package comm

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"strconv"

	"github.com/JanFant/TLServer/logger"
	"github.com/ruraomsk/ag-server/pudge"
	"github.com/ruraomsk/ag-server/setup"
)

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
func workerCommand(soc net.Conn) {
	defer soc.Close()
	var command CommandARM
	logger.Info.Printf("Новый клиент комманд %s", soc.RemoteAddr().String())
	reader := bufio.NewReader(soc)
	for {
		c, err := reader.ReadString('\n')
		if err != nil {
			logger.Error.Println("При чтении команд сервера АРМ ", err.Error())
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
			logger.Error.Printf("Команда сервера АРМ %v нет такого id %d", command, command.ID)
			continue
		}
		logger.Info.Printf("Команда %v", command)
		if command.Command == 1 {
			//Принудительная отправка всех массивов
			ctrl, _ := pudge.GetController(command.ID)
			ctrl.Arrays = make([]pudge.ArrayPriv, 0)
			w := fmt.Sprintf("Пользователь %s  заказал перезагрузку всех массивов", command.User)
			ctrl.LastLogString = w
			pudge.ChanLog <- pudge.RecLogCtrl{ID: command.ID, LogString: w}
			pudge.SetController(ctrl)

			logger.Info.Printf("id %d массив привязок поставлен на перезагрузку", command.ID)
		} else {
			ctrl, _ := pudge.GetController(command.ID)
			w := fmt.Sprintf("Пользователь %s  %s", command.User, getDescription(command))
			ctrl.LastLogString = w
			pudge.ChanLog <- pudge.RecLogCtrl{ID: command.ID, LogString: w}
			pudge.SetController(ctrl)
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
			_, is := pudge.GetCross(state.State.Region, state.State.Area, state.State.ID)
			if !is {
				//Перекрестка нет
				logger.Info.Printf("Попытка удаления удаленного перекрестка %d %d %d", state.State.Region, state.State.Area, state.State.ID)
				continue
			}
			logger.Info.Printf("Удаление перекрестка %d %d %d", state.State.Region, state.State.Area, state.State.ID)
			pudge.DeleteCross(state.State.Region, state.State.Area, state.State.ID)
			w := fmt.Sprintf("Пользователь %s удаление перекрестка %d %d %d", state.User, state.State.Region, state.State.Area, state.State.ID)
			pudge.ChanLog <- pudge.RecLogCtrl{ID: state.State.IDevice, LogString: w}
			ctrl, is := pudge.GetController(state.State.IDevice)
			if is {
				ctrl.LastLogString = w
				pudge.SetController(ctrl)
			}
			continue
		}
		_, is := pudge.GetCross(state.State.Region, state.State.Area, state.State.ID)
		if !is {
			//Перекрестка нет нужно создать
			logger.Info.Printf("Добавлен перекресток %d %d %d", state.State.Region, state.State.Area, state.State.ID)
			state.State.StatusDevice = 18
			ctrl, is := pudge.GetController(state.State.IDevice)
			w := fmt.Sprintf("Пользователь %s добаление перекрестка %d %d %d", state.User, state.State.Region, state.State.Area, state.State.ID)
			if is {
				ctrl.LastLogString = w
				pudge.SetController(ctrl)
			}
			logger.Info.Print(w)
			pudge.ChanLog <- pudge.RecLogCtrl{ID: state.State.IDevice, LogString: w}
		}
		// logger.Info.Printf("Write new state ")

		pudge.SetCross(&state.State)
		ctrl, is := pudge.GetController(state.State.IDevice)
		w := fmt.Sprintf("Пользователь %s изменил перекресток %d %d %d", state.User, state.State.Region, state.State.Area, state.State.ID)
		logger.Info.Print(w)
		if is {
			ctrl.LastLogString = w
			pudge.SetController(ctrl)
		}
		pudge.ChanLog <- pudge.RecLogCtrl{ID: state.State.IDevice, LogString: w}
	}
}
func workerProtocol(soc net.Conn) {
	defer soc.Close()
	var protocol ChangeProtocol
	logger.Info.Printf("Новый клиент комманд %s", soc.RemoteAddr().String())
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
		logger.Info.Printf("Команда %v", protocol)
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
			return "Отправлена команда Переход на автоматическое регулирование ПК"
		}
		return "Отправлена команда Сменить ПК на " + strconv.Itoa(toSend.Params)
	case 6:
		if toSend.Params == 0 {
			return "Отправлена команда Переход на автоматическое регулирование СК"
		}
		return "Отправлена команда Сменить CК на " + strconv.Itoa(toSend.Params)
	case 7:
		if toSend.Params == 0 {
			return "Отправлена команда Переход на автоматическое регулирование НК"
		}
		return "Отправлена команда Сменить НК на " + strconv.Itoa(toSend.Params)
	}
	switch toSend.Params {
	case 0:
		return "Отправлена команда Локальный режим"
	case 9:
		return "Отправлена команда Координированное управление"
	case 10:
		return "Отправлена команда Включить жёлтое мигание"
	case 11:
		return "Отправлена команда Отключить светофоры"
	}
	return "Отправлена команда Включить фазу №" + strconv.Itoa(toSend.Params)
}
