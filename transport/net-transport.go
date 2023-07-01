package transport

import (
	"fmt"
	"net"
	"time"

	"github.com/ruraomsk/ag-server/debug"
	"github.com/ruraomsk/ag-server/extcon"
	"github.com/ruraomsk/ag-server/logger"
)

// Stoped глобальная перенная если истина то надо бросать работу
var Stoped = false

func GetOneMessage(socket net.Conn) (HeaderDevice, error) {
	var h HeaderDevice
	socket.SetReadDeadline(time.Now().Add(30 * time.Second))
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
	debug.DebugChan <- debug.DebugMessage{ID: h.ID, Time: time.Now(), FromTo: false, Buffer: buffer}
	return h, nil
}

// GetMessagesFromDevice принять сообщение
func GetMessagesFromDevice(socket net.Conn, hcan chan HeaderDevice, tout *time.Duration, errTcp chan net.Conn) {
	defer socket.Close()
	var h HeaderDevice
	for {
		if Stoped {
			return
		}
		socket.SetReadDeadline(time.Now().Add(*tout))
		// socket.SetReadDeadline(time.Unix(0, 0))
		buf := make([]byte, 19)
		n, err := socket.Read(buf)
		if Stoped {
			return
		}
		if err == nil && n != len(buf) {
			message := fmt.Sprintf("при чтении сообщения от устройства %s прочитано %d байт нужно %d", socket.RemoteAddr().String(), n, len(buf))
			debug.DebugChan <- debug.DebugMessage{ID: h.ID, Time: time.Now(), FromTo: false, Info: true, Buffer: []byte(message)}
			logger.Error.Printf(message)
			errTcp <- socket
			return
		}
		if err != nil {
			message := fmt.Sprintf("при чтении сообщения от устройства %s %s", socket.RemoteAddr().String(), err.Error())
			debug.DebugChan <- debug.DebugMessage{ID: h.ID, Time: time.Now(), FromTo: false, Info: true, Buffer: []byte(message)}
			logger.Error.Printf(message)
			errTcp <- socket
			return
		}
		buf2 := make([]byte, buf[18]+2)
		n, err = socket.Read(buf2)
		if Stoped {
			return
		}
		if err == nil && n != len(buf2) {
			message := fmt.Sprintf("при чтении сообщения от устройства %s прочитано %d байт нужно %d", socket.RemoteAddr().String(), n, len(buf2))
			debug.DebugChan <- debug.DebugMessage{ID: h.ID, Time: time.Now(), FromTo: false, Info: true, Buffer: []byte(message)}
			logger.Error.Printf(message)
			errTcp <- socket
			return
		}
		if err != nil {
			message := fmt.Sprintf("при чтении сообщения от устройства %s %s", socket.RemoteAddr().String(), err.Error())
			debug.DebugChan <- debug.DebugMessage{ID: h.ID, Time: time.Now(), FromTo: false, Info: true, Buffer: []byte(message)}
			logger.Error.Printf(message)
			errTcp <- socket
			return
		}
		buffer := append(buf, buf2...)
		//logger.Debug.Printf("Приняли с сокета %ss %v",socket.RemoteAddr().String(),buffer)
		err = h.Parse(buffer)
		if err != nil {
			message := fmt.Sprintf("при раскодировании от устройства %s %s", socket.RemoteAddr().String(), err.Error())
			debug.DebugChan <- debug.DebugMessage{ID: h.ID, Time: time.Now(), FromTo: false, Info: true, Buffer: []byte(message)}
			// debug.DebugChan <- debug.DebugMessage{ID: h.ID, Time: time.Now(), FromTo: false, Buffer: buffer}
			logger.Error.Printf(message)
			errTcp <- socket
			return

		}
		hcan <- h
		debug.DebugChan <- debug.DebugMessage{ID: h.ID, Time: time.Now(), FromTo: false, Buffer: buffer}
		if Stoped {
			return
		}
	}
}

// GetMessagesFromServer прием сообщений от сервера
func GetMessagesFromServer(socket net.Conn, hcan chan HeaderServer, tout *time.Duration, errTcp chan net.Conn) {
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
			errTcp <- socket
			return
		}
		if err != nil {
			logger.Error.Printf("при чтении сообщения от сервера %s %s", socket.RemoteAddr().String(), err.Error())
			errTcp <- socket
			return
		}
		buf2 := make([]byte, buf[12]+2)
		n, err = socket.Read(buf2)
		if Stoped {
			return
		}
		if err == nil && n != len(buf2) {
			logger.Error.Printf("при чтении сообщения от сервера %s прочитано неверно", socket.RemoteAddr().String())
			errTcp <- socket
			return
		}
		if err != nil {
			logger.Error.Printf("при чтении сообщения от сервера %s %s", socket.RemoteAddr().String(), err.Error())
			errTcp <- socket
			return
		}
		buffer := append(buf, buf2...)
		err = hs.Parse(buffer)
		if err != nil {
			logger.Error.Printf("при раскодировании от сервера %s %s", socket.RemoteAddr().String(), err.Error())
			errTcp <- socket
			return

		}
		hcan <- hs
	}
}

// SendMessagesToDevice передать сообщение на устройство
func SendMessagesToDevice(socket net.Conn, hout chan HeaderServer, tout *time.Duration, errTcp chan net.Conn, id int) {
	defer socket.Close()
	// timer := extcon.SetTimerClock(time.Duration(10 * time.Second))
	for {
		select {
		// case <-timer.C:
		// 	if Stoped {
		// 		return
		// 	}
		case hs := <-hout:
			// logger.Debug.Printf("Отправляем на %s %v", socket.RemoteAddr().String(), hs)
			socket.SetWriteDeadline(time.Now().Add(10 * time.Second))
			buffer := hs.MakeBuffer()
			n, err := socket.Write(buffer)
			if Stoped {
				return
			}
			if err != nil {
				message := fmt.Sprintf("при передаче от устройства %s %s", socket.RemoteAddr().String(), err.Error())
				debug.DebugChan <- debug.DebugMessage{ID: id, Time: time.Now(), FromTo: false, Info: true, Buffer: []byte(message)}
				logger.Error.Printf(message)
				errTcp <- socket
				return
			}
			if n != len(buffer) {
				message := fmt.Sprintf("при передаче от устройства %s неверно передано байт %d %d", socket.RemoteAddr().String(), len(buffer), n)
				debug.DebugChan <- debug.DebugMessage{ID: id, Time: time.Now(), FromTo: false, Info: true, Buffer: []byte(message)}
				logger.Error.Printf(message)
				errTcp <- socket
				return
			}
			debug.DebugChan <- debug.DebugMessage{ID: id, Time: time.Now(), FromTo: true, Buffer: buffer}
		}
	}
}

// SendMessagesToServer передать сообщение на устройство
func SendMessagesToServer(socket net.Conn, hout chan HeaderDevice, tout *time.Duration, errTcp chan net.Conn) {
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
				errTcp <- socket
				return
			}
			if n != len(buffer) {
				logger.Error.Printf("при передаче на сервер %s неверно передано байт %d %d", socket.RemoteAddr().String(), len(buffer), n)
				errTcp <- socket
				return
			}
		}
	}
}
