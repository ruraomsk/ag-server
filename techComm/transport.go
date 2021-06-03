package techComm

import (
	"github.com/ruraomsk/TLServer/logger"
	"github.com/ruraomsk/ag-server/extcon"
	"github.com/ruraomsk/ag-server/transport"
	"time"
)

//SendMessagesToDevice передать сообщение на устройство
func (d *Device) SendMessagesToDevice() {
	timer := extcon.SetTimerClock(1 * time.Second)
	for {
		select {
		case <-timer.C:
			if !d.Work {
				return
			}
		case hs := <-d.hOut:
			// logger.Debug.Printf("Отправляем на %s %v", socket.RemoteAddr().String(), hs)
			_ = d.Socket.SetWriteDeadline(time.Now().Add(d.tOut))
			buffer := hs.MakeBuffer()
			n, err := d.Socket.Write(buffer)
			if !d.Work {
				return
			}
			if err != nil {
				logger.Error.Printf("при передаче от устройства %s %s", d.Socket.RemoteAddr().String(), err.Error())
				d.ErrorTCP <- 0
				return
			}
			if n != len(buffer) {
				logger.Error.Printf("при передаче от устройства %s неверно передано байт %d %d", d.Socket.RemoteAddr().String(), len(buffer), n)
				d.ErrorTCP <- 0
				return
			}
			d.trafficOut(len(buffer))
		}
	}
}

//GetMessagesFromDevice принять сообщение
func (d *Device) GetMessagesFromDevice() {
	var h transport.HeaderDevice
	for {
		if !d.Work {
			return
		}
		_ = d.Socket.SetReadDeadline(time.Now().Add(d.tIn))
		buf := make([]byte, 19)
		n, err := d.Socket.Read(buf)
		if !d.Work {
			return
		}
		if err == nil && n != len(buf) {
			logger.Error.Printf("при чтении сообщения от устройства %s прочитано %d байт нужно %d", d.Socket.RemoteAddr().String(), n, len(buf))
			d.ErrorTCP <- 1
			return
		}
		if err != nil {
			logger.Error.Printf("при чтении сообщения от устройства %s %s", d.Socket.RemoteAddr().String(), err.Error())
			d.ErrorTCP <- 1
			return
		}
		buf2 := make([]byte, buf[18]+2)
		n, err = d.Socket.Read(buf2)
		if !d.Work {
			return
		}
		if err == nil && n != len(buf2) {
			logger.Error.Printf("при чтении сообщения от устройства %s прочитано %d байт нужно %d", d.Socket.RemoteAddr().String(), n, len(buf2))
			d.ErrorTCP <- 1
			return
		}
		if err != nil {
			logger.Error.Printf("при чтении сообщения от устройства %s %s", d.Socket.RemoteAddr().String(), err.Error())
			d.ErrorTCP <- 1
			return
		}
		buffer := append(buf, buf2...)
		d.trafficIn(len(buffer))
		//logger.Debug.Printf("Приняли с сокета %ss %v",socket.RemoteAddr().String(),buffer)
		err = h.Parse(buffer)
		if err != nil {
			logger.Error.Printf("при раскодировании от устройства %s %s", d.Socket.RemoteAddr().String(), err.Error())
			d.ErrorTCP <- 1
			return

		}
		d.hIn <- h
	}
}
