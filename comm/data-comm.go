package comm

import (
	"fmt"
	"net"
	"rura/ag-server/extcon"
	"rura/ag-server/logger"
	"strconv"
	"sync"
	"time"
)

var mapPins map[int]Pin
var mutex sync.Mutex

//Pin распредение пинов по портам
type Pin struct {
	Port    int `json:"port"` //Номер порта на прием привязанный к данному пину
	ID      int `json:"id"`   //Устройство прикрепленное к порту
	inbuff  [1024]byte
	outbuf  [1024]byte
	Status  int                `json:"scomm"`
	Lastop  time.Time          //Время последней операции обмена
	context *extcon.ExtContext // Расширенный контекст для управления портом
}

func getId(conn net.Conn) (int, error) {
	return 0, nil

}
func (p *Pin) loop() {
	ln, err := net.Listen("tcp", ":"+strconv.Itoa(p.Port))
	if err != nil {
		logger.Error.Printf("Not listen port %d %s", p.Port, err.Error())
		return
	}
	for {
		conn, err := ln.Accept()
		if err != nil {
			logger.Error.Printf("Not accept port %d %s", p.Port, err.Error())
			return
		}
		id, err := getId(conn)
		if err != nil {
			logger.Error.Printf("Not recive id port %d %s", p.Port, err.Error())
			continue
		}
		fmt.Println(id)
		// if p.Status<0

	}

	// time.Sleep(1 * time.Second)
}

func addPin(port int) bool {

	_, is := mapPins[port]
	if is {
		// Pin уже обслуживается принимай решение по проблеме
		return false
	}
	var p Pin
	p.Port = port
	p.Lastop = time.Unix(0, 0)
	p.Status = -1
	mutex.Lock()
	mapPins[port] = p
	mutex.Unlock()
	go p.loop()
	return true
}
