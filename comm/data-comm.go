package comm

import (
	"net"
	"rura/ag-server/extcon"
	"time"
)

type device struct {
	id int
	//Идентификатор устройства
	inbuff  [1024]byte
	outbuf  [1024]byte
	status  int
	lastop  time.Time          //Время последней операции обмена
	context *extcon.ExtContext //Расширенный контекст для управления устройством
	com     net.Conn
}
