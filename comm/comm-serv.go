package comm

import (
	"net"
	"rura/ag-server/extcon"
	"rura/ag-server/logger"
	"rura/ag-server/pudge"
	"rura/ag-server/setup"
	"strconv"
	"sync"
	"time"
)

var mapDevs map[int]device
var mutex sync.Mutex

//Это сервер коммуникации
//Слушает входящие сообщения и распределяет их на устройства

//clearID читает из буфера ID аозвращает чистый id и тип устройства
// Тип устройства закодирован в первом сисмоле ID
func clearID(bufID []byte) (int, int) {
	return 1, 1
}
func createDevFromController(devc pudge.Controller) device {
	var dev device
	dev.id = devc.ID
	dev.context, _ = extcon.NewContext("dev" + strconv.Itoa(dev.id))
	return dev
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
		logger.Error.Printf("Пришла неверная длина ID %d с устройcтва на %s %s", len, com.RemoteAddr().String(), err.Error())
		return
	}
	// Проверим была ли уже с этим устройством связь
	id, typeDevice := clearID(bufID)
	dev, is := mapDevs[id]
	if is {
		// Если есть обмен с этим устройством остановим его
		dev.context.Cancel()
		for {
			time.Sleep(100 * time.Millisecond)
			if !dev.context.IsExecuted() {
				break
			}
		}
	}
	// Проверим есть ли он в pudge?
	devc, is := pudge.GetController(id)
	if !is {
		//Абсолютно первое подключения за все время системы
		devc, err = pudge.CreateController(id, typeDevice)
		if err != nil {
			//Такого устройства мы вообще не знаем
			logger.Error.Printf("Неизвестный  ID %d с устройcтва на %s %s", id, com.RemoteAddr().String(), err.Error())
			return
		}
		//devc теперь содежит заполненный контроллер для запуска обмена
		//поместим его в базу
		devc.Comment = "Первое подключение"
	} else {
		devc.Comment = "Переподключение"
	}
	pudge.SetController(devc)
	dev = createDevFromController(devc)
	dev.com = com
	mutex.Lock()
	mapDevs[dev.id] = dev
	mutex.Unlock()

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
