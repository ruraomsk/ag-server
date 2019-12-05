package comm

import (
	"rura/ag-server/extcon"
)

type device struct {
	id int
	//Идентификатор устройства
	context *extcon.ExtContext //Расширенный контекст для управления устройством
	NumDev  uint8              //Номер сообщения для подтверждения
	NumServ uint8              //Номер сообщения от сервера
	WaitNum uint8              //Номер ожидаемого сообщения
}

func (d *device) addNumber() {
	d.NumServ++
	if d.NumServ > 250 {
		d.NumServ = 1
	}
}
