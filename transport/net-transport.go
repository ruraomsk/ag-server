package transport

import (
	"net"
	"rura/ag-server/logger"
	"rura/ag-server/setup"
	"time"
)

//GetMessagesFromDevice принять сообщение
func GetMessagesFromDevice(socket net.Conn, hcan chan HeaderDevice) {
	defer socket.Close()
	socket.SetReadDeadline(time.Now().Add(setup.Set.CommServer.TimeOutRead))
	var h HeaderDevice
	for {
		buf := make([]byte, 19)
		n, err := socket.Read(buf)
		if err == nil && n != len(buf) {
			logger.Error.Printf("при чтении сообщения от устройства %s прочитано %d байт нужно %d", socket.RemoteAddr().String(), n, len(buf))
			return
		}
		if err != nil {
			logger.Error.Printf("при чтении сообщения от устройства %s %s", socket.RemoteAddr().String(), err.Error())
			return
		}
		buf2 := make([]byte, buf[18]+2)
		n, err = socket.Read(buf2)
		if err == nil && n != len(buf2) {
			logger.Error.Printf("при чтении сообщения от устройства %s %s", socket.RemoteAddr().String(), err.Error())
			return
		}
		if err != nil {
			logger.Error.Printf("при чтении сообщения от устройства %s %s", socket.RemoteAddr().String(), err.Error())
			return
		}
		buffer := append(buf, buf2...)
		err = h.Parse(buffer)
		if err != nil {
			logger.Error.Printf("при раскодировании от устройства %s %s", socket.RemoteAddr().String(), err.Error())
			return

		}
		hcan <- h
	}
}

//GetMessagesFromService прием сообщений от сервера
func GetMessagesFromService(socket net.Conn, hcan chan HeaderServer) {
	defer socket.Close()
	var hs HeaderServer
	for {
		socket.SetReadDeadline(time.Now().Add(setup.Set.CommServer.TimeOutRead))
		buf := make([]byte, 13)
		n, err := socket.Read(buf)
		if err == nil && n != len(buf) {
			logger.Error.Printf("при чтении сообщения от сервера %s прочитано %d байт нужно %d", socket.RemoteAddr().String(), n, len(buf))
			return
		}
		if err != nil {
			logger.Error.Printf("при чтении сообщения от сервера %s %s", socket.RemoteAddr().String(), err.Error())
			return
		}
		buf2 := make([]byte, buf[12]+2)
		n, err = socket.Read(buf2)
		if err == nil && n != len(buf2) {
			logger.Error.Printf("при чтении сообщения от сервера %s %s", socket.RemoteAddr().String(), err.Error())
			return
		}
		if err != nil {
			logger.Error.Printf("при чтении сообщения от сервера %s %s", socket.RemoteAddr().String(), err.Error())
			return
		}
		buffer := append(buf, buf2...)
		err = hs.Parse(buffer)
		if err != nil {
			logger.Error.Printf("при раскодировании от сервера %s %s", socket.RemoteAddr().String(), err.Error())
			return

		}
		hcan <- hs
	}
}

//SendMessagesToDevice передать сообщение на устройство
func SendMessagesToDevice(socket net.Conn, hout chan HeaderServer) {
	defer socket.Close()
	for {
		hs := <-hout
		socket.SetWriteDeadline(time.Now().Add(setup.Set.CommServer.TimeOutWrite))
		buffer := hs.MakeBuffer()
		n, err := socket.Write(buffer)
		if err != nil {
			logger.Error.Printf("при передаче от устройства %s %s", socket.RemoteAddr().String(), err.Error())
			return
		}
		if n != len(buffer) {
			logger.Error.Printf("при передаче от устройства %s неверно передано байт %d %d", socket.RemoteAddr().String(), len(buffer), n)
			return
		}
	}
}

//SendMessagesToServer передать сообщение на устройство
func SendMessagesToServer(socket net.Conn, hout chan HeaderDevice) {
	defer socket.Close()
	for {
		hd := <-hout
		socket.SetWriteDeadline(time.Now().Add(setup.Set.CommServer.TimeOutWrite))
		buffer := hd.MakeBuffer()
		n, err := socket.Write(buffer)
		if err != nil {
			logger.Error.Printf("при передаче на сервер  %s %s", socket.RemoteAddr().String(), err.Error())
			return
		}
		if n != len(buffer) {
			logger.Error.Printf("при передаче на сервер %s неверно передано байт %d %d", socket.RemoteAddr().String(), len(buffer), n)
			return
		}
	}
}
