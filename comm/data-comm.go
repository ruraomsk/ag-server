package comm

import (
	"rura/ag-server/extcon"
	"sync"
	"time"
)

var mutex sync.Mutex

//Pin распредение пинов по портам
type device struct {
	id int
	//Идентификатор устройства
	inbuff  [1024]byte
	outbuf  [1024]byte
	status  int
	lastop  time.Time          //Время последней операции обмена
	context *extcon.ExtContext //Расширенный контекст для управления устройством
}
