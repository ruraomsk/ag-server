package main

import "testing"

import "github.com/ruraomsk/ag-server/transport"

import "strings"

var sbd transport.SubMessage

func Test_0x01S(t *testing.T) {
	sbd.Set0x01Server(127)
	i := sbd.Get0x01Server()
	if i != 127 {
		t.Errorf("Ошибка! %d", i)
	}
}
func Test_0x02S(t *testing.T) {
	sb.Set0x02Server(true)
	i := sb.Get0x02Server()
	if i != true {
		t.Errorf("Ошибка! %v", i)
	}
}
func Test_0x03S(t *testing.T) {
	sb.Set0x03Server()
	if sb.Message[0] != 3 {
		t.Errorf("Ошибка! ")
	}
}
func Test_0x04S(t *testing.T) {
	sb.Set0x04Server(true, false)
	i := sb.Get0x04Server()
	if i[0] != true {
		t.Errorf("Ошибка! %v", i)
	}
	if i[1] != false {
		t.Errorf("Ошибка! %v", i)
	}
}

func Test_0x05S(t *testing.T) {
	sb.Set0x05Server(17)
	i := sb.Get0x05Server()
	if i != 17 {
		t.Errorf("Ошибка! %v", i)
	}
}
func Test_0x06S(t *testing.T) {
	sb.Set0x06Server(17)
	i := sb.Get0x06Server()
	if i != 17 {
		t.Errorf("Ошибка! %v", i)
	}
}
func Test_0x07S(t *testing.T) {
	sb.Set0x07Server(17)
	i := sb.Get0x07Server()
	if i != 17 {
		t.Errorf("Ошибка! %v", i)
	}
}
func Test_0x09S(t *testing.T) {
	sb.Set0x09Server(17)
	i := sb.Get0x09Server()
	if i != 17 {
		t.Errorf("Ошибка! %v", i)
	}
}
func Test_0x0AS(t *testing.T) {
	sb.Set0x0AServer(17)
	i := sb.Get0x0AServer()
	if i != 17 {
		t.Errorf("Ошибка! %v", i)
	}
}
func Test_0x0BS(t *testing.T) {
	sb.Set0x0BServer(17, 77)
	i := sb.Get0x0BServer()
	if i[0] != 17 {
		t.Errorf("Ошибка! %v", i)
	}
	if i[1] != 77 {
		t.Errorf("Ошибка! %v", i)
	}
}
func Test_0x32S(t *testing.T) {
	err := sb.Set0x32Server("127.0.0.1", 999)
	if err != nil {
		t.Errorf("Ошибка! %s", err.Error())
	}
	ip, port, err := sb.Get0x32Server()
	if err != nil {
		t.Errorf("Ошибка! %s", err.Error())
	}
	if port != 999 {
		t.Errorf("Ошибка port! %v", port)
	}
	ipe := "127.000.000.001"
	if strings.Compare(ip, ipe) != 0 {
		t.Errorf("Ошибка! <%s> <%s> %d %d", ip, ipe, len(ip), len(ipe))
	}
}
func Test_0x33S(t *testing.T) {
	sb.Set0x33Server(317)
	i := sb.Get0x33Server()
	if i != 317 {
		t.Errorf("Ошибка! %v", i)
	}
}
func Test_0x34S(t *testing.T) {
	sb.Set0x34Server(true)
	i := sb.Get0x34Server()
	if i != true {
		t.Errorf("Ошибка! %v", i)
	}
}
func Test_0x35S(t *testing.T) {
	sb.Set0x35Server(17, true)
	i, ii := sb.Get0x35Server()
	if i != 17 {
		t.Errorf("Ошибка! %v", i)
	}
	if ii != true {
		t.Errorf("Ошибка! %v", ii)
	}
}
