package transport

import (
	"github.com/ruraomsk/ag-server/logger"
	"github.com/ruraomsk/ag-server/setup"
	"net"
	"time"
)

//GetMessagesFromDevice принять сообщение
func GetMessagesFromDevice(socket net.Conn, hcan chan HeaderDevice, status *bool) {
	defer socket.Close()
	*status = true
	var h HeaderDevice
	for {
		// socket.SetReadDeadline(time.Now().Add(setup.Set.CommServer.TimeOutRead))
		socket.SetReadDeadline(time.Now().Add(time.Duration(5 * time.Minute)))
		buf := make([]byte, 19)
		n, err := socket.Read(buf)
		if err == nil && n != len(buf) {
			logger.Error.Printf("при чтении сообщения от устройства %s прочитано %d байт нужно %d", socket.RemoteAddr().String(), n, len(buf))
			hcan <- h
			*status = false
			return
		}
		if err != nil {
			logger.Error.Printf("при чтении сообщения от устройства %s %s", socket.RemoteAddr().String(), err.Error())
			hcan <- h
			*status = false
			return
		}
		buf2 := make([]byte, buf[18]+2)
		n, err = socket.Read(buf2)
		if err == nil && n != len(buf2) {
			logger.Error.Printf("при чтении сообщения от устройства %s прочитано %d байт нужно %d", socket.RemoteAddr().String(), n, len(buf2))
			hcan <- h
			*status = false
			return
		}
		if err != nil {
			logger.Error.Printf("при чтении сообщения от устройства %s %s", socket.RemoteAddr().String(), err.Error())
			hcan <- h
			*status = false
			return
		}
		buffer := append(buf, buf2...)
		err = h.Parse(buffer)
		if err != nil {
			logger.Error.Printf("при раскодировании от устройства %s %s", socket.RemoteAddr().String(), err.Error())
			hcan <- h
			*status = false
			return

		}
		// logger.Info.Printf("in %v", h)
		hcan <- h
	}
}

//GetMessagesFromService прием сообщений от сервера
func GetMessagesFromService(socket net.Conn, hcan chan HeaderServer) {
	defer socket.Close()
	var hs HeaderServer
	for {
		// socket.SetReadDeadline(time.Now().Add(setup.Set.CommServer.TimeOutRead))
		// socket.SetReadDeadline(time.Now().Add(time.Duration(5 * time.Minute)))
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
func SendMessagesToDevice(socket net.Conn, hout chan HeaderServer, status *bool) {
	defer socket.Close()
	*status = true
	for {
		select {
		case hs := <-hout:
			socket.SetWriteDeadline(time.Now().Add(setup.Set.CommServer.TimeOutWrite))
			buffer := hs.MakeBuffer()
			n, err := socket.Write(buffer)
			if err != nil {
				logger.Error.Printf("при передаче от устройства %s %s", socket.RemoteAddr().String(), err.Error())
				*status = false
				return
			}
			if n != len(buffer) {
				logger.Error.Printf("при передаче от устройства %s неверно передано байт %d %d", socket.RemoteAddr().String(), len(buffer), n)
				*status = false
				return
			}
			// logger.Info.Printf("out %v", hs)

		}
	}
}

//SendMessagesToServer передать сообщение на устройство
func SendMessagesToServer(socket net.Conn, hout chan HeaderDevice) {
	defer socket.Close()
	for {
		select {
		case hd := <-hout:
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
}
