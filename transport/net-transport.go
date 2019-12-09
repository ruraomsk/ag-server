package transport

import (
	"fmt"
	"net"
	"rura/ag-server/setup"
	"strings"
	"time"
)

//GetMessageFromDevice принять сообщение от устройства в любом случае
func GetMessageFromDevice(socket net.Conn) (HeaderDevice, error) {
	var h HeaderDevice
	buf := make([]byte, 19)
	n, err := socket.Read(buf)
	if err == nil && n != len(buf) {
		err = fmt.Errorf("при чтении сообщения от устройства прочитано %d байт нужно %d", n, len(buf))
	}
	if err != nil {
		return h, err
	}
	buf2 := make([]byte, buf[18]+2)
	n, err = socket.Read(buf2)
	if err == nil && n != len(buf2) {
		err = fmt.Errorf("при чтении сообщения от устройства прочитано %d байт нужно %d", n, len(buf2))
	}
	if err != nil {
		return h, err
	}
	buffer := append(buf, buf2...)
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

//GetMaybeMessageFromDevice принять сообщение от устройства если оно есть
//Если за заднный интервал не пришло сообщение то вернет false,nil
//Если были ошибки при приеме то вернет false,error
//Если прием произошел то вернет true,nil и заполненный HeaderDevice
func GetMaybeMessageFromDevice(socket net.Conn, h *HeaderDevice) (bool, error) {
	socket.SetReadDeadline(time.Now().Add(setup.Set.CommServer.TimeOutRead))
	buf := make([]byte, 19)
	n, err := socket.Read(buf)
	if err != nil && strings.Contains(err.Error(), "i/o timeout") {
		return false, nil
	}
	if err == nil && n != len(buf) {
		err = fmt.Errorf("при чтении сообщения от устройства прочитано %d байт нужно %d", n, len(buf))
	}
	if err != nil {
		return false, err
	}
	buf2 := make([]byte, buf[18]+2)
	n, err = socket.Read(buf2)
	if err != nil && strings.Contains(err.Error(), "i/o timeout") {
		return false, nil
	}
	if err == nil && n != len(buf2) {
		err = fmt.Errorf("при чтении сообщения от устройства прочитано %d байт нужно %d", n, len(buf2))
	}
	if err != nil {
		return false, err
	}
	buffer := append(buf, buf2...)
	err = h.Parse(buffer)
	return true, err
}
