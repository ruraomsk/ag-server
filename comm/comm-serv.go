package comm

import (
	"net"
	"rura/ag-server/extcon"
	"rura/ag-server/logger"
	"rura/ag-server/pudge"
	"rura/ag-server/setup"
	"strconv"
	"time"
)

var mapDevs map[int]device

//Это сервер коммуникации
//Слушает входящие сообщения и распределяет их на устройства

//clearID читает из буфера ID аозвращает чистый id и тип устройства
// Тип устройства закодирован в первом сисмоле ID
func clearID(bufID []byte) (int, int) {
	return 1, 1
}
func getTempID(com net.Conn) {
	defer com.Close()
	bufID := make([]byte, setup.Set.CommServer.LenID)
	com.SetReadDeadline(time.Now().Add(time.Duration(setup.Set.CommServer.TimeOutRead)))
	len, err := com.Read(bufID)
	if err != nil {
		logger.Error.Printf("Ошибка чтения ID с устройтва на %s %s", com.RemoteAddr().String(), err.Error())
		return

	}
	if len != setup.Set.CommServer.LenID {
		logger.Error.Printf("Пришла неверная длина ID %d с устройтва на %s %s", len, com.RemoteAddr().String(), err.Error())
		return
	}
	// Проверим была ли уже с этим устройством связь
	id, typeDevice := clearID(bufID)
	dev, is := mapDevs[id]
	if !is {
		// С момента запуска сервера еще не было обменов
		// Проверим есть ли он в pudge?
		cdev, ispudge := pudge.GetController(id)
		if !ispudge {
			//Абсолютно первое подключения за все время системы
		}
	}

}

//Start основной вход сервер коммуникаций
func Start(context *extcon.ExtContext) {
	ln, err := net.Listen("tcp", ":"+strconv.Itoa(setup.Set.CommServer.Port))

	if err != nil {
		logger.Error.Printf("Ошибка открытия порта %s", err.Error())
		return
	}
	defer ln.Close()
	for {
		com, err := ln.Accept()
		if err != nil {
			logger.Error.Printf("Ошибка accept %s", err.Error())
			continue
		}
		go getTempID(com)
	}
}
