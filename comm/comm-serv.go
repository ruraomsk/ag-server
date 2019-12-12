package comm

import (
	"fmt"
	"net"
	"rura/ag-server/extcon"
	"rura/ag-server/logger"
	"rura/ag-server/pudge"
	"rura/ag-server/setup"
	"rura/ag-server/transport"
	"strconv"
	"sync"
	"time"
)

var devs map[int]*device
var mutex sync.Mutex

//Это сервер коммуникации
//Слушает входящие сообщения и распределяет их на устройства

//StartListen основной вход сервер коммуникаций
func StartListen(stop chan int) {
	//Запускаем слушателя для команд от АРМ
	go listenArmCommand()
	//Запускаем слушателя для массивов привязки от АРМ
	go listenArmArray()

	devs = make(map[int]*device)
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
		go newConnect(socket, stop)
	}
}
func newConnect(soc net.Conn, stop chan int) {
	/*
			После установления соединения:
		1.Клиент отправляет сообщение Состояние ПБС V2, если ID клиента есть в БД сервера,
		сервер отвечает пустым сообщением 0x00, клиент, получив ответ, запускает процедуру
		начала обмена (п.2)
		2. В одном пакете клиент отправляет сообщения Установление соединения V2,Начало
		работы, Состояние оборудования , Состояние ДК V3, сервер отвечает
		подтверждением с номером принятого пакета.
	*/
	var hDev transport.HeaderDevice
	var ctrl pudge.Controller
	var err error
	defer soc.Close()

	hDev, err = transport.GetMessageFromDevice(soc)
	if err != nil {
		logger.Error.Printf("При приеме первого соединения от устройства %s %s", soc.LocalAddr().String(), err.Error())
		return
	}
	ctrl, err = getController(hDev.ID)
	if err != nil {
		logger.Error.Printf("Устройств %s %s", soc.LocalAddr().String(), err.Error())
		return
	}

	dmess := hDev.ParseMessage()
	flag := false
	for _, m := range dmess {
		if m.Type == 0x10 {
			flag = true
			m.Get0x10Device(&ctrl)
		}
	}
	if !flag {
		//В сообщении соединении нет 0x10 значит рвем связь
		logger.Error.Printf("Устройство %d неверный формат подключения", hDev.ID)
		return
	}
	//Обновим состояние в pudge
	ctrl.StatusConnection = pudge.Connected
	ctrl.LastOperation = time.Now()
	pudge.SetController(ctrl)
	//Готовим пустое сообщение
	hs := transport.CreateHeaderServer(0, 0)
	err = transport.SendMessageToDevice(soc, hs)
	if err != nil {
		logger.Error.Printf("При передаче %s", err.Error())
		return
	}
	//Проверим есть ли зарегистрированный слушатель нашего id и скажем ему что
	//теперь есть новый и ему можно завершиться
	//Ждем сообщения о состоянии устройства
	hDev, err = transport.GetMessageFromDevice(soc)
	if err != nil {
		logger.Error.Printf("При ожидании состояния устройства %s", err.Error())
		return
	}
	dd := new(device)
	dd.id = ctrl.ID
	dd.CommandARM = make(chan CommandARM)
	dd.CommandArray = make(chan CommandArray)
	dd.addNumber()
	dd.context, _ = extcon.NewContext("device" + strconv.Itoa(dd.id))
	mutex.Lock()
	devs[dd.id] = dd
	mutex.Unlock()
	updateController(&ctrl, &hDev)
	pudge.SetController(ctrl)
	//С этого момента начинается основной цикл работы
	/*
	   3. В процессе работы, при изменении состояния ДК или оборудования, клиент отправляет
	   сообщение &quot;Состояние оборудования V2&quot;, &quot;Состояние ДК V3&quot;, (при одновременном
	   изменении, оба сообщения отправляются в одном пакете), сервер отвечает
	   подтверждением с номером принятого пакета.
	   4. Если в течение интервала контроля связи не было событий, требующих передачи
	   состояния, по истечении интервала передается &quot;Состояние ДК V3&quot; с текущим состоянием,
	   сервер отвечает подтверждением с номером принятого пакета.
	   5. При изменении состояния ПБС, передается &quot;Состояние ПБС V2&quot;, сервер отвечает
	   пустым сообщением 0x00.
	   6. Если у клиента активен режим сбора статистики, по завершению периода усреднения,
	   передаются сообщения &quot;Завершение периода усреднения&quot; и &quot;Передача статистики&quot;,
	   сервер отвечает подтверждением с номером принятого пакета.
	   7. При необходимости изменения параметров платы ПБС (IP адрес, порт, режим обмена)
	   сервер передает клиенту соответствующее сообщение, клиент отвечает сообщениями
	   &quot;Состояние ПБС V2&quot; и &quot;Подтверждение от ПБС&quot; в одном пакете, сервер в ответ передает
	   сообщение 0x00.
	   8. При передаче массивов привязки, сервер передает сообщение &quot;Управление УСДК –
	   Отключить&quot;, далее массивы привязки, объединенные в сообщения, по завершению
	   &quot;Управление УСДК – Включить&quot;. Клиент подтверждает каждое принятое сообщение.
	*/
	for {
		is, err := transport.GetMaybeMessageFromDevice(soc, &hDev)
		if err != nil {
			logger.Error.Printf("При приеме сообщения от %d %s", dd.id, err.Error())
			ctrl.StatusConnection = pudge.NotConnected
			pudge.SetController(ctrl)
			return
		}
		if is {
			updateController(&ctrl, &hDev)
			pudge.SetController(ctrl)
		}
		timer := extcon.SetTimerClock(time.Duration(10 * time.Second))
		select {
		case <-timer:
			if time.Now().Sub(ctrl.LastOperation) > setup.Set.CommServer.KeepAlive {
				//Уже пять минут нет связи с устройством
				//Прощаемся с ним %-)
				ctrl.StatusConnection = pudge.NotConnected
				logger.Info.Printf("Устройство %d более положенного не выходит на связь", dd.id)
				return
			}
		case <-dd.context.Done():
			logger.Info.Printf("Устройство %d приказано умереть", dd.id)
			return
		case comARM := <-dd.CommandARM:
			//Пришла команда арма
			hs, err = makeCommandToDevice(dd, comARM)
			if err != nil {
				logger.Error.Printf("При создании команды от АРМ %d %s", dd.id, err.Error())
				continue
			}
			err = transport.SendMessageToDevice(soc, hs)
			if err != nil {
				logger.Error.Printf("При передаче %d %s", dd.id, err.Error())
				return
			}

		case comArray := <-dd.CommandArray:
			//Пришла команда арма загрузки привязки

			hss := makeArrayToDevice(dd, comArray)
			for _, h := range hss {
				err = transport.SendMessageToDevice(soc, h)
				if err != nil {
					logger.Error.Printf("При передаче %d %s", dd.id, err.Error())
					return
				}
				time.Sleep(10 * time.Second)
			}

		}
	}

}

//Считывает полученную информацию от устройства и распаковывет ее в контроллер
func updateController(c *pudge.Controller, hDev *transport.HeaderDevice) {
	dmess := hDev.ParseMessage()
	mutex.Lock()
	d := devs[hDev.ID]
	c.LastOperation = time.Now()
	c.StatusConnection = pudge.Connected
	defer mutex.Unlock()
	for _, mes := range dmess {
		switch mes.Type {
		case 0x00:
			//Пустое сообщение ничего не делаем
		case 0x01:
			num, _, _, _, _ := mes.Get0x01Device()

			if num == int(d.WaitNum) {
				d.WaitNum = 0
			}
		case 0x04:
			c.StatusDevice = 1
			c.Base = false
		case 0x07:
			c.StatusDevice = 1
			c.Base = true
		case 0x09:
			//Устройство закончило сбор статистики проверим если есть такая то обновим ее заголовок
			//если нет то добавим новый заголовок в массив статистики
			st, err := mes.Get0x09Device()
			if err != nil {
				logger.Error.Printf("Разбор x09 от устройства %d %s", hDev.ID, err.Error())
				continue
			}
			flag := false
			for n, stt := range c.Statistics {
				if stt.Period == st.Period {
					flag = true
					c.Statistics[n] = st
					break
				}
			}
			if !flag {
				c.Statistics = append(c.Statistics, st)
			}
		case 0x0a:
			//Собственно статистика пришла
			nper := int(mes.Message[1])
			flag := false
			for n, stt := range c.Statistics {
				if stt.Period == nper {
					flag = true
					err := mes.Get0x0ADevice(&stt)
					if err != nil {
						logger.Error.Printf("При разборе команды 0x0a id %d %s", hDev.ID, err.Error())
					}
					c.Statistics[n] = stt
					break
				}
			}
			if !flag {
				logger.Error.Printf("При разборе команды 0x0a id %d нет такого периода %d", hDev.ID, nper)
			}
		case 0x0B:
			//Прием сохраненного журнала от устройства
			var lg pudge.LogLine
			err := mes.Get0x0BDevice(&lg)
			if err != nil {
				logger.Error.Printf("При разборе команды 0x0B id %d %s", hDev.ID, err.Error())
				continue
			}
			c.LogLines = append(c.LogLines, lg)
		case 0x0f:
			//Установление связи ДК v2
			err := mes.Get0x0FDevice(c)
			if err != nil {
				logger.Error.Printf("При разборе команды 0x0f id %d %s", hDev.ID, err.Error())
				continue
			}
		case 0x10:
			logger.Error.Printf("Повторная выдача команды 0x10 id %d ", hDev.ID)
		case 0x11:
			//Состояние оборудования v2
			err := mes.Get0x11Device(c)
			if err != nil {
				logger.Error.Printf("При разборе команды 0x11 id %d %s", hDev.ID, err.Error())
				continue
			}
		case 0x12:
			//Состояние ДК v3
			err := mes.Get0x12Device(c)
			if err != nil {
				logger.Error.Printf("При разборе команды 0x12 id %d %s", hDev.ID, err.Error())
				continue
			}
		case 0x13:
			//Массив приявязки
			var ar pudge.ArrayPriv
			err := mes.Get0x13Device(&ar)
			if err != nil {
				logger.Error.Printf("При разборе команды 0x13 id %d %s", hDev.ID, err.Error())
				continue
			}
			flag := false
			for n, a := range c.Arrays {
				if a.Number == ar.Number {
					//Заменим массив
					c.Arrays[n] = ar
					flag = true
				}
			}
			if !flag {
				c.Arrays = append(c.Arrays, ar)
			}
		default:
			logger.Error.Printf("От %d неверная команда %x", hDev.ID, mes.Type)
		}

	}
	return
}
func getController(id int) (pudge.Controller, error) {
	//Вначале проверим на pudge
	ctrl, is := pudge.GetController(id)
	if !is {
		//Нет на pudge теперь надо проверить среди регистрированных
		is, name := pudge.IsRegistred(id)
		if !is {
			return ctrl, fmt.Errorf("id %d не зарегистрирован", id)
		}
		ctrl = pudge.CreateEmptyController(id)
		ctrl.Name = name
	}
	return ctrl, nil
}
func makeCommandToDevice(dd *device, comARM CommandARM) (transport.HeaderServer, error) {
	dd.addNumber()
	hs := transport.CreateHeaderServer(int(dd.NumServ), 0)
	mss := make([]transport.SubMessage, 0)
	var ms transport.SubMessage
	switch comARM.Command {
	case 0x02:
		//Управление УСДК
		ms.Set0x02Server(comARM.Command == 2)
	case 0x03:
		//Запрос состояния устройства
		ms.Set0x03Server()
	case 0x04:
		//Запрос на смену фаз
		d1 := (comARM.Params & 1) != 0
		d2 := (comARM.Params & 2) != 0
		ms.Set0x04Server(d1, d2)
	case 0x05:
		//Смена плана координации
		ms.Set0x05Server(comARM.Params)
	case 0x06:
		//Смена карты выбора по времении суток
		ms.Set0x06Server(comARM.Params)
	case 0x07:
		//Смена недельной карты
		ms.Set0x07Server(comARM.Params)
	case 0x09:
		//Диспетчерское управление ДК1
		ms.Set0x09Server(comARM.Params)
	case 0x0A:
		//Диспетчерское управление ДК2
		ms.Set0x0AServer(comARM.Params)
	default:
		return hs, fmt.Errorf("Неверная команда от АРМ для %d user %d %x ", dd.id, comARM.UserID, comARM.Command)
	}
	mss = append(mss, ms)
	hs.UpackMessages(mss)
	return hs, nil
}
func makeArrayToDevice(dd *device, comArray CommandArray) []transport.HeaderServer {
	hss := make([]transport.HeaderServer, 0)
	var ms transport.SubMessage
	//Сообщение об отключении управления
	hs := transport.CreateHeaderServer(0, 0)
	mss := make([]transport.SubMessage, 0)
	ms.Set0x02Server(false)
	mss = append(mss, ms)
	hs.UpackMessages(mss)
	hss = append(hss, hs)
	//Собственно массив привязки
	dd.addNumber()
	hs = transport.CreateHeaderServer(int(dd.NumServ), 0)
	mss = make([]transport.SubMessage, 0)
	ms.SetArray(comArray.Number, comArray.Elems)
	mss = append(mss, ms)
	hs.UpackMessages(mss)
	hss = append(hss, hs)
	//Сообщение о включении управления
	hs = transport.CreateHeaderServer(0, 0)
	mss = make([]transport.SubMessage, 0)
	ms.Set0x02Server(true)
	mss = append(mss, ms)
	hs.UpackMessages(mss)
	hss = append(hss, hs)
	return hss
}
