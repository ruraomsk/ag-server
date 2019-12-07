package comm

import (
	"rura/ag-server/extcon"
)

//CommandARM Команды от Сервера АРМ
type CommandARM struct {
	ID      int `json:"id"`
	UserID  int `json:"user"`
	Command int `json:"cmd"`
	Params  int `json:"param"`
}

//CommandArray Привязка от Сервера АРМ
type CommandArray struct {
	ID     int   `json:"id"`
	UserID int   `json:"user"`
	Number int   `json:"number"`
	Elems  []int `json:"elems"`
}

type device struct {
	id int
	//Идентификатор устройства
	context      *extcon.ExtContext //Расширенный контекст для управления устройством
	NumDev       uint8              //Номер сообщения для подтверждения
	NumServ      uint8              //Номер сообщения от сервера
	WaitNum      uint8              //Номер ожидаемого сообщения
	CommandARM   chan CommandARM
	CommandArray chan CommandArray
}

func (d *device) addNumber() {
	d.NumServ++
	if d.NumServ > 250 {
		d.NumServ = 1
	}
}
