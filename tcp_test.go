package main

import (
	"net"
	"rura/ag-server/pudge"
	"rura/ag-server/transport"
	"testing"
	"time"
)

var rServer chan transport.HeaderServer
var rDevice chan transport.HeaderDevice

func startListen(t *testing.T) {
	count := 0
	ln, err := net.Listen("tcp", ":4000")

	if err != nil {
		t.Errorf("Ошибка %d открытия порта %s", count, err.Error())
		return
	}
	defer ln.Close()
	for {

		socket, err := ln.Accept()
		if err != nil {
			t.Errorf("Ошибка %d accept %s", count, err.Error())
			continue
		}
		count++
		go goclient(socket)
	}
}
func goclient(soc net.Conn) {
	for {
		buffer := make([]byte, 1024)
		soc.Read(buffer)
	}
}
func startServer(t *testing.T) {
	//Запускаем слушателя для команд от АРМ
	ln, err := net.Listen("tcp", ":2000")

	if err != nil {
		t.Errorf("Ошибка открытия порта %s", err.Error())
		return
	}
	defer ln.Close()
	for {
		socket, err := ln.Accept()
		if err != nil {
			t.Errorf("Ошибка accept %s", err.Error())
			continue
		}
		buf := make([]byte, 13)
		n, err := socket.Read(buf)
		if err == nil && n != len(buf) {
			t.Errorf("при чтении сообщения от сервера прочитано %d байт нужно %d", n, len(buf))
		}
		if err != nil {
			t.Errorf("Ошибка чтения от сервера %s", err.Error())
		}
		buf2 := make([]byte, buf[12]+2)
		n, err = socket.Read(buf2)
		if err == nil && n != len(buf2) {
			t.Errorf("при чтении сообщения от сервера прочитано %d байт нужно %d", n, len(buf2))
		}
		if err != nil {
			t.Errorf("Ошибка чтения от сервера %s", err.Error())
		}
		buffer := append(buf, buf2...)
		var hs transport.HeaderServer
		err = hs.Parse(buffer)
		if err != nil {
			t.Errorf("при разборе  сообщения от сервера %s", err.Error())
		}
		rServer <- hs

	}
}
func startDevice(t *testing.T) {
	//Запускаем слушателя для команд от АРМ
	ln, err := net.Listen("tcp", ":3000")

	if err != nil {
		t.Errorf("Ошибка открытия порта %s", err.Error())
		return
	}
	defer ln.Close()
	for {
		socket, err := ln.Accept()
		if err != nil {
			t.Errorf("Ошибка accept %s", err.Error())
			continue
		}
		buf := make([]byte, 19)
		n, err := socket.Read(buf)
		if err == nil && n != len(buf) {
			t.Errorf("при чтении сообщения от устройства прочитано %d байт нужно %d", n, len(buf))
		}
		if err != nil {
			t.Errorf("Ошибка чтения от сервера %s", err.Error())
		}
		buf2 := make([]byte, buf[18]+2)
		n, err = socket.Read(buf2)
		if err == nil && n != len(buf2) {
			t.Errorf("при чтении сообщения от устройства прочитано %d байт нужно %d", n, len(buf2))
		}
		if err != nil {
			t.Errorf("Ошибка чтения от устройства %s", err.Error())
		}
		buffer := append(buf, buf2...)
		var hd transport.HeaderDevice
		err = hd.Parse(buffer)
		if err != nil {
			t.Errorf("при разборе  сообщения от устройства %s", err.Error())
		}
		rDevice <- hd

	}
}
func Test_MaxListens(t *testing.T) {
	go startListen(t)
	time.Sleep(5 * time.Second)
	for i := 0; i < 4000; i++ {
		s, err := net.Dial("tcp", ":4000")
		if err != nil {
			t.Errorf("Ошибка %d соединения с портом %s", i, err.Error())
			return
		}
		defer s.Close()

	}
}
func Test_TcpServer(t *testing.T) {
	rServer = make(chan transport.HeaderServer)
	go startServer(t)
	var hs transport.HeaderServer
	hs.IDServer = uint8(0xa7)
	hs.Time = time.Now()
	hs.Code = 0x7f
	hs.Number = 1
	var ms transport.SubMessage
	mss := make([]transport.SubMessage, 0)
	ms.Set0x01Server(10)
	mss = append(mss, ms)
	ms.Set0x02Server(true)
	mss = append(mss, ms)
	ms.Set0x03Server()
	mss = append(mss, ms)
	ms.Set0x04Server(true, true)
	mss = append(mss, ms)
	ms.Set0x05Server(11)
	mss = append(mss, ms)
	ms.Set0x06Server(3)
	mss = append(mss, ms)
	ms.Set0x07Server(12)
	mss = append(mss, ms)
	ms.Set0x09Server(13)
	mss = append(mss, ms)
	ms.Set0x0AServer(14)
	mss = append(mss, ms)
	hs.UpackMessages(mss)
	buffer := hs.MakeBuffer()
	time.Sleep(5 * time.Second)
	s, err := net.Dial("tcp", ":"+"2000")
	if err != nil {
		t.Errorf("Ошибка соединения с портом %s", err.Error())
		return
	}
	defer s.Close()
	n, err := s.Write(buffer)
	if err != nil {
		t.Errorf("Ошибка передачи %s", err.Error())
		return
	}
	if n != len(buffer) {
		t.Errorf("Передано %d нужно %d", n, len(buffer))
		return
	}
	// var nhs transport.HeaderServer
	nhs := <-rServer

	err = nhs.Parse(buffer)
	if err != nil {
		t.Error(err.Error())
	}
	if !hs.Compare(&nhs) {
		t.Errorf("Не совпали HeaderServer \n%v \n%v\n", hs, nhs)
	}
	smess := hs.ParseMessage()
	nsmess := nhs.ParseMessage()
	for n, mes := range smess {
		nmes := nsmess[n]
		if !nmes.Compare(&mes) {
			t.Error(nmes.ToString(), " not equal ", mes.ToString())
		}
	}

}
func Test_TcpDevice(t *testing.T) {
	rDevice = make(chan transport.HeaderDevice)
	go startDevice(t)
	var hd transport.HeaderDevice
	hd = transport.CreateHeaderDevice(128978, 30, 0, 0xAC)
	var ms transport.SubMessage
	mss := make([]transport.SubMessage, 0)
	ms.Set0x00Device()
	mss = append(mss, ms)
	ms.Set0x01Device(1, 2, 3, 4, 5)
	mss = append(mss, ms)
	ms.Set0x04Device(10, 11, 9, 12)
	mss = append(mss, ms)
	ms.Set0x07Device(7, 8, 9, 10)
	mss = append(mss, ms)
	var st pudge.Statistic
	ms.Set0x09Device(&st)
	mss = append(mss, ms)
	ms.Set0x0ADevice(&st)
	mss = append(mss, ms)
	var c pudge.Controller
	ms.Set0x0FDevice(&c)
	mss = append(mss, ms)
	ms.Set0x10Device(&c)
	mss = append(mss, ms)
	ms.Set0x11Device(&c)
	mss = append(mss, ms)
	ms.Set0x12Device(&c)
	mss = append(mss, ms)
	ms.Set0x1DDevice(&c)
	mss = append(mss, ms)
	var ar pudge.ArrayPriv
	ar.Array = make([]int, 0)
	// ms.Set0x13Device(&ar)
	// mss = append(mss, ms)

	hd.UpackMessages(mss)
	buffer := hd.MakeBuffer()
	time.Sleep(5 * time.Second)
	s, err := net.Dial("tcp", ":"+"3000")
	if err != nil {
		t.Errorf("Ошибка соединения с портом %s", err.Error())
		return
	}
	defer s.Close()
	n, err := s.Write(buffer)
	if err != nil {
		t.Errorf("Ошибка передачи %s", err.Error())
		return
	}
	if n != len(buffer) {
		t.Errorf("Передано %d нужно %d", n, len(buffer))
		return
	}
	nhd := <-rDevice
	err = nhd.Parse(buffer)
	if err != nil {
		t.Error(err.Error())
	}
	if !hd.Compare(&nhd) {
		t.Errorf("Не совпали \n%v\n%v\n ", hd, nhd)
	}
	smess := hd.ParseMessage()
	nsmess := nhd.ParseMessage()
	for n, mes := range smess {
		nmes := nsmess[n]
		if !nmes.Compare(&mes) {
			t.Error(nmes.ToString(), " not equal ", mes.ToString())
		}
	}

}
