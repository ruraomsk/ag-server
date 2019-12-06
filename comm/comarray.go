package comm

import (
	"bufio"
	"encoding/json"
	"net"
	"rura/ag-server/logger"
	"rura/ag-server/setup"
	"strconv"
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
	reader := bufio.NewReader(soc)
	for {
		c, err := reader.ReadString(0)
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
		dev.CommandARM <- command
	}
}
func workerArray(soc net.Conn) {
	defer soc.Close()
	var array CommandArray
	reader := bufio.NewReader(soc)
	for {
		a, err := reader.ReadString(0)
		if err != nil {
			logger.Error.Println("При чтении привязки от сервера АРМ ", err.Error())
			return
		}
		err = json.Unmarshal([]byte(a), &array)
		if err != nil {
			logger.Error.Println("При конвератации привязки сервера АРМ ", err.Error())
			continue
		}
		dev, is := devs[array.ID]
		if !is {
			logger.Error.Printf("Команда привязки сервера АРМ нет такого id %d", array.ID)
			continue
		}
		dev.CommandArray <- array
	}

}
