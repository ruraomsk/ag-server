package transport

import (
	"fmt"
	"net"
	"rura/ag-server/setup"
	"time"
)

//GetMessageFromDevice принять сообщение от устройства в любом случае
func GetMessageFromDevice(socket net.Conn) (HeaderDevice, error) {
	var h HeaderDevice
	buffer := make([]byte, 1024)
	socket.SetReadDeadline(time.Now().Add(setup.Set.CommServer.TimeOutRead))
	len, err := socket.Read(buffer)
	if err != nil {
		return h, err
	}
	if len == 0 {
		return h, fmt.Errorf("прочитано ноль байт от устройства %s", socket.LocalAddr().String())
	}
	err = h.Parse(buffer)
	return h, err
}

//SendMessageToDevice передать сообщение на устройство
func SendMessageToDevice(socket net.Conn, hs HeaderServer) error {
	socket.SetWriteDeadline(time.Now().Add(setup.Set.CommServer.TimeOutWrite))
	buffer := hs.MakeBuffer()
	n, err := socket.Write(buffer)
	if err != nil {
		return err
	}
	if n != len(buffer) {
		return fmt.Errorf("передано %d байт вместо %d на устройство %s", n, len(buffer), socket.LocalAddr().String())
	}
	return nil
}
