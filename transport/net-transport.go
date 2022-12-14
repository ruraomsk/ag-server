package transport

import (
	"fmt"
	"net"
	"time"

	"github.com/ruraomsk/TLServer/logger"
	"github.com/ruraomsk/ag-server/extcon"
)

//Stoped глобальная перенная если истина то надо бросать работу
var Stoped = false

func GetOneMessage(socket net.Conn) (HeaderDevice, error) {
	var h HeaderDevice
	socket.SetReadDeadline(time.Now().Add(10 * time.Minute))
	buf := make([]byte, 19)
	n, err := socket.Read(buf)
	if err == nil && n != len(buf) {
		return h, fmt.Errorf("при чтении сообщения от устройства %s прочитано %d байт нужно %d", socket.RemoteAddr().String(), n, len(buf))
	}
	if err != nil {
		return h, fmt.Errorf("при чтении сообщения от устройства %s %s", socket.RemoteAddr().String(), err.Error())
	}
	buf2 := make([]byte, buf[18]+2)
	n, err = socket.Read(buf2)
	if err == nil && n != len(buf2) {
		return h, fmt.Errorf("при чтении сообщения от устройства %s прочитано %d байт нужно %d", socket.RemoteAddr().String(), n, len(buf2))
	}
	if err != nil {
		return h, fmt.Errorf("при чтении сообщения от устройства %s %s", socket.RemoteAddr().String(), err.Error())
	}
	buffer := append(buf, buf2...)
	//logger.Debug.Printf("Приняли с сокета %ss %v",socket.RemoteAddr().String(),buffer)
	err = h.Parse(buffer)
	if err != nil {
		return h, fmt.Errorf("при раскодировании от устройства %s %s", socket.RemoteAddr().String(), err.Error())
	}
	return h, nil
}

//GetMessagesFromDevice принять сообщение
func GetMessagesFromDevice(socket net.Conn, hcan chan HeaderDevice, tout *time.Duration, errTcp chan int) {
	defer socket.Close()
	var h HeaderDevice
	for {
		if Stoped {
			return
		}
		socket.SetReadDeadline(time.Now().Add(15 * time.Minute))
		buf := make([]byte, 19)
		n, err := socket.Read(buf)
		if Stoped {
			return
		}
		if err == nil && n != len(buf) {
			logger.Error.Printf("при чтении сообщения от устройства %s прочитано %d байт нужно %d", socket.RemoteAddr().String(), n, len(buf))
			errTcp <- 1
			return
		}
		if err != nil {
			logger.Error.Printf("при чтении сообщения от устройства %s %s", socket.RemoteAddr().String(), err.Error())
			errTcp <- 1
			return
		}
		buf2 := make([]byte, buf[18]+2)
		n, err = socket.Read(buf2)
		if Stoped {
			return
		}
		if err == nil && n != len(buf2) {
			logger.Error.Printf("при чтении сообщения от устройства %s прочитано %d байт нужно %d", socket.RemoteAddr().String(), n, len(buf2))
			errTcp <- 1
			return
		}
		if err != nil {
			logger.Error.Printf("при чтении сообщения от устройства %s %s", socket.RemoteAddr().String(), err.Error())
			errTcp <- 1
			return
		}
		buffer := append(buf, buf2...)
		//logger.Debug.Printf("Приняли с сокета %ss %v",socket.RemoteAddr().String(),buffer)
		err = h.Parse(buffer)
		if err != nil {
			logger.Error.Printf("при раскодировании от устройства %s %s", socket.RemoteAddr().String(), err.Error())
			errTcp <- 1
			return

		}
		hcan <- h
		if Stoped {
			return
		}
	}
}

//GetMessagesFromService прием сообщений от сервера
func GetMessagesFromService(socket net.Conn, hcan chan HeaderServer, tout *time.Duration) {
	defer socket.Close()
	var hs HeaderServer
	for {
		if Stoped {
			return
		}
		socket.SetReadDeadline(time.Now().Add(*tout))
		buf := make([]byte, 13)
		n, err := socket.Read(buf)
		if Stoped {
			return
		}
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
		if Stoped {
			return
		}
		if err == nil && n != len(buf2) {
			logger.Error.Printf("при чтении сообщения от сервера %s прочитано неверно", socket.RemoteAddr().String())
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
func SendMessagesToDevice(socket net.Conn, hout chan HeaderServer, tout *time.Duration, errTcp chan int) {
	defer socket.Close()
	timer := extcon.SetTimerClock(time.Duration(1 * time.Second))
	for {
		select {
		case <-timer.C:
			if Stoped {
				return
			}
		case hs := <-hout:
			// logger.Debug.Printf("Отправляем на %s %v", socket.RemoteAddr().String(), hs)
			socket.SetWriteDeadline(time.Now().Add(*tout))
			buffer := hs.MakeBuffer()
			n, err := socket.Write(buffer)
			if Stoped {
				return
			}
			if err != nil {
				logger.Error.Printf("при передаче от устройства %s %s", socket.RemoteAddr().String(), err.Error())
				errTcp <- 0
				return
			}
			if n != len(buffer) {
				logger.Error.Printf("при передаче от устройства %s неверно передано байт %d %d", socket.RemoteAddr().String(), len(buffer), n)
				errTcp <- 0
				return
			}
		}
	}
}

//SendMessagesToServer передать сообщение на устройство
func SendMessagesToServer(socket net.Conn, hout chan HeaderDevice, tout *time.Duration) {
	defer socket.Close()
	timer := extcon.SetTimerClock(time.Duration(1 * time.Second))
	for {
		select {
		case <-timer.C:
			if Stoped {
				return
			}
		case hd := <-hout:
			if Stoped {
				return
			}
			socket.SetWriteDeadline(time.Now().Add(*tout))
			buffer := hd.MakeBuffer()
			n, err := socket.Write(buffer)
			if Stoped {
				return
			}
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
