package techComm

import (
	"github.com/ruraomsk/TLServer/logger"
	"github.com/ruraomsk/ag-server/memDB"
	"github.com/ruraomsk/ag-server/setup"
	"github.com/ruraomsk/ag-server/transport"
	"net"
	"strconv"
)

func StartListen() {
	//Запускаем слушателя для команд от АРМ
	go listenArmCommand()
	// //Запускаем слушателя для массивов привязки от АРМ
	go listenArmArray()
	// //Запускаем слушателя для настройки протокола
	go listenChangeProtocol()
	//writeArch = make(chan pudge.ArchStat, 1000)
	go listenSendingPhazes()
	ln, err := net.Listen("tcp", ":"+strconv.Itoa(setup.Set.CommServer.Port))

	if err != nil {
		logger.Error.Printf("Ошибка открытия порта %s", err.Error())
		return
	}
	defer ln.Close()
	for {
		socket, err := ln.Accept()
		if err != nil {
			logger.Error.Printf("Ошибка accept %s", err.Error())
			continue
		}
		go newConnect(socket)
	}
}
func newConnect(socket net.Conn) {
	//Разобраться с подключением и если все ок запустить обработчика
	hDev, err := transport.GetOneMessage(socket)
	if err != nil {
		logger.Error.Print(err.Error())
		socket.Close()
		return
	}
	logger.Info.Printf("Устройствo %s подключается... номер %d", socket.RemoteAddr().String(), hDev.ID)
	if isDeviceWork(hDev.ID) {
		logger.Info.Printf("Повторное подключение %d", hDev.ID)
		stopDevice(hDev.ID)
	}
	ctrl, err := getController(hDev.ID)
	if err != nil {
		logger.Error.Printf("Устройствo %s %s", socket.RemoteAddr().String(), err.Error())
		return
	}
	dmess := hDev.ParseMessage()
	flag := false
	for _, m := range dmess {
		switch m.Type {
		case 0x1D:
			{
				flag = true
				_ = m.Get0x1DDevice(&ctrl)
			}
		case 0x10:
			{
				flag = true
				_ = m.Get0x10Device(&ctrl)

			}
		case 0x12:
			{
				flag = true
				_ = m.Get0x12Device(&ctrl)
			}
		case 0x1B:
			{
				flag = true
				_ = m.Get0x1BDevice(&ctrl)

			}
		case 0x11:
			{
				flag = true
				_ = m.Get0x11Device(&ctrl)
			}
		case 0x1C:
			{
				flag = true
			}

		}
	}
	if !flag {
		//В сообщении соединении нет 0x10 или 0x1D значит рвем связь
		logger.Error.Printf("Устройство %d неверный формат подключения", hDev.ID)
		logger.Error.Printf("Устройство %d прислало %v", hDev.ID, dmess)
		return
	}
	ctrl.StatusConnection = true
	memDB.SetController(ctrl)
	d := newDevice(ctrl, socket)
	addDevice(d)
	//Все ок можно создать обработчика и начинать обмен с устройством
	go d.Worker(hDev)
}
