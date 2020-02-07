package comm

import (
	"bufio"
	"encoding/json"
	"net"
	"strconv"

	"github.com/ruraomsk/ag-server/logger"
	"github.com/ruraomsk/ag-server/pudge"
	"github.com/ruraomsk/ag-server/setup"
)

func listenArmCommand() {
	ln, err := net.Listen("tcp", ":"+strconv.Itoa(setup.Set.CommServer.PortCommand))
	if err != nil {
		logger.Error.Printf("Ошибка открытия порта %s", err.Error())
		return
	}
	defer ln.Close()
	for {
		socket, err := ln.Accept()
		if err != nil {
			logger.Error.Printf("Ошибка accept %s", err.Error())
			continue
		}
		go workerCommand(socket)
	}
}
func listenArmArray() {
	ln, err := net.Listen("tcp", ":"+strconv.Itoa(setup.Set.CommServer.PortArray))

	if err != nil {
		logger.Error.Printf("Ошибка открытия порта %s", err.Error())
		return
	}
	defer ln.Close()
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
		err = json.Unmarshal([]byte(c), &command)
		if err != nil {
			logger.Error.Println("При конвератации команд сервера АРМ ", err.Error())
			continue
		}
		dev, is := devs[command.ID]
		if !is {
			logger.Error.Printf("Команда сервера АРМ нет такого id %d", command.ID)
			continue
		}
		logger.Info.Println("Команда %v", command)
		dev.CommandARM <- command
	}
}
func workerArray(soc net.Conn) {
	defer soc.Close()
	var state pudge.Cross
	logger.Info.Printf("Новый клиент массивов %s", soc.RemoteAddr().String())
	reader := bufio.NewReader(soc)
	for {
		a, err := reader.ReadString('\n')
		if err != nil {
			logger.Error.Println("При чтении привязки от сервера АРМ ", err.Error())
			return
		}
		err = json.Unmarshal([]byte(a), &state)
		if err != nil {
			logger.Error.Println("При конвератации привязки сервера АРМ ", err.Error())
			continue
		}
		_, is := devs[state.IDevice]
		if !is {
			logger.Error.Printf("Команда привязки сервера АРМ нет такого id %d", state.IDevice)
			continue
		}
		_, is = pudge.GetCross(state.Region, state.Area, state.ID)
		if !is {
			//Перекрестка нет нужно создать
			logger.Info.Printf("Добавлен перекресток %d %d %d", state.Region, state.Area, state.ID)
		}
		// logger.Info.Printf("Write status %v", state)
		pudge.SetCross(&state)
	}

}
