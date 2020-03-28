package comm

import (
	"fmt"
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/JanFant/TLServer/logger"
	"github.com/ruraomsk/ag-server/extcon"
	"github.com/ruraomsk/ag-server/pudge"
	"github.com/ruraomsk/ag-server/setup"
	"github.com/ruraomsk/ag-server/transport"
)

var devs map[int]*device
var mutex sync.Mutex

// var answare chan string
// var request chan int

//Это сервер коммуникации
//Слушает входящие сообщения и распределяет их на устройства

//GetChanArray возвращает канал для присылки массивов для данного устройства
func GetChanArray(id int) (chan CommandArray, bool) {
	d, is := devs[id]
	if !is {
		return nil, false
	}
	return d.CommandArray, true
}

//StartListen основной вход сервер коммуникаций
func StartListen(stop chan int) {
	for !pudge.Works {
		time.Sleep(1 * time.Second)
	}

	//Запускаем слушателя для команд от АРМ
	go listenArmCommand()
	// //Запускаем слушателя для массивов привязки от АРМ
	go listenArmArray()
	count := 0
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
		if !pudge.Works {
			return
		}
		count++
		if count%5 == 0 {
			logger.Info.Println("Входящих соединений", count)
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
	logger.Debug.Printf("Устройствo %s подключается...", soc.RemoteAddr().String())
	ctrl := new(pudge.Controller)
	var err error
	hout := make(chan transport.HeaderServer, 100)
	hin := make(chan transport.HeaderDevice, 100)
	defer soc.Close()
	readTout := time.Duration(setup.Set.CommServer.TimeOutRead * int64(time.Second))
	writeTout := time.Duration(setup.Set.CommServer.TimeOutWrite * int64(time.Second))
	go transport.GetMessagesFromDevice(soc, hin, &readTout)
	go transport.SendMessagesToDevice(soc, hout, &writeTout)
	hDev := <-hin
	// logger.Info.Printf("hDev %v", hDev)
	start := time.Now()
	ctrl, err = getController(hDev.ID)
	if err != nil {
		logger.Error.Printf("Устройствo %s %s", soc.RemoteAddr().String(), err.Error())
		return
	}
	if ctrl.TimeOut != 0 {
		readTout = time.Duration(ctrl.TimeOut * int64(time.Second))
	} else {
		ctrl.TimeOut = setup.Set.CommServer.TimeOutRead
		pudge.SetController(ctrl)
	}
	if ctrl.TMax != 0 {
		writeTout = time.Duration(ctrl.TMax * int64(time.Second))
	} else {
		ctrl.TMax = setup.Set.CommServer.TimeOutWrite
		pudge.SetController(ctrl)
	}
	dmess := hDev.ParseMessage()
	flag := false
	for _, m := range dmess {
		if m.Type == 0x1D {
			flag = true
			m.Get0x1DDevice(ctrl)
		}
		if m.Type == 0x10 {
			flag = true
			m.Get0x10Device(ctrl)
		}
		if m.Type == 0x12 {
			flag = true
			m.Get0x12Device(ctrl)
		}
		if m.Type == 0x1B {
			flag = true
			// m.Get0x1BDevice()
		}
		if m.Type == 0x1C {
			flag = true
			// m.Get0x1CDevice()
		}
	}
	if !flag {
		//В сообщении соединении нет 0x10 или 0x1D значит рвем связь
		logger.Error.Printf("Устройство %d неверный формат подключения", hDev.ID)
		logger.Error.Printf("Устройство %d прислало %v", hDev.ID, dmess)
		return
	}
	if time.Now().Sub(start) > time.Duration(10*time.Second) {
		logger.Info.Println("больше 10 секунд ", ctrl.ID)
	}
	//Обновим состояние в pudge
	ctrl.StatusConnection = pudge.Connected
	ctrl.LastOperation = time.Now()
	dd := new(device)
	dd.id = ctrl.ID
	dd.CommandARM = make(chan CommandARM)
	dd.CommandArray = make(chan CommandArray)
	dd.addNumber()
	dd.context, _ = extcon.NewContext("device" + strconv.Itoa(dd.id))
	mutex.Lock()
	devs[dd.id] = dd
	mutex.Unlock()
	updateController(ctrl, &hDev)
	if ctrl.Base {
		ctrl.Arrays = make([]pudge.ArrayPriv, 0)
	}
	pudge.SetController(ctrl)
	//Подтвердим что клиент прописан
	hs := transport.CreateHeaderServer(0, int(hDev.Code))
	mss := make([]transport.SubMessage, 0)
	hs.UpackMessages(mss)
	hout <- hs
	pudge.ChanLog <- pudge.RecLogCtrl{ID: ctrl.ID, LogString: "Подключен"}
	//Проверим есть ли зарегистрированный слушатель нашего id и скажем ему что
	//теперь есть новый и ему можно завершиться
	//Ждем сообщения о состоянии устройства
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
	timer := extcon.SetTimerClock(time.Duration(1 * time.Second))
	for {
		select {
		case hDev = <-hin:
			lastBase := ctrl.Base
			hs, need := updateController(ctrl, &hDev)
			if ctrl.Base && !lastBase {
				ctrl.Arrays = make([]pudge.ArrayPriv, 0)
			}
			pudge.SetController(ctrl)
			if len(hs.Message) != 0 || need {
				hout <- hs
			}
		case <-timer.C:
			if time.Now().Sub(ctrl.LastOperation) > readTout {
				//Уже пять минут нет связи с устройством
				//Прощаемся с ним %-)
				ctrl.StatusConnection = pudge.NotConnected
				pudge.SetController(ctrl)
				w := fmt.Sprintf("Устройство %d более %f не выходит на связь ", dd.id, readTout.Seconds())
				pudge.ChanLog <- pudge.RecLogCtrl{ID: ctrl.ID, LogString: w}
				logger.Error.Print(w)
				return
			}

		case <-dd.context.Done():
			transport.Stoped = true
			// pudge.ChanLog <- pudge.RecLogCtrl{ID: ctrl.ID, LogString: "Остановлен сервер"}
			logger.Info.Printf("Устройство %d приказано умереть", dd.id)

			return
		case comARM := <-dd.CommandARM:
			//Пришла команда арма
			hs, err = makeCommandToDevice(dd, comARM)
			if err != nil {
				logger.Error.Printf("При создании команды от АРМ %d %s", dd.id, err.Error())
				continue
			}
			hout <- hs

		case comArray := <-dd.CommandArray:
			//Пришла команда арма загрузки привязки
			if comArray.ID == 0 && comArray.Number == 0 {
				//Команда перейти в локальный режим
				hs := makeLocalOn(dd)
				// logger.Debug.Printf("Local on %d", dd.id)
				hout <- hs
				break
			}
			if comArray.ID == 0 && comArray.Number == 1 {
				//Команда выйти из локального режима
				hs := makeLocalOff(dd)
				// logger.Debug.Printf("Local off %d", dd.id)
				hout <- hs
				break
			}

			is := false
			for n, ap := range ctrl.Arrays {
				if ap.Number == comArray.Number && ap.NElem == comArray.NElem {
					ctrl.Arrays[n].Array = comArray.Elems
					is = true
					break
				}
			}
			if !is {
				ap := new(pudge.ArrayPriv)
				ap.Number = comArray.Number
				ap.NElem = comArray.NElem
				ap.Array = comArray.Elems
				ctrl.Arrays = append(ctrl.Arrays, *ap)
			}
			pudge.SetController(ctrl)
			hs := makeArrayToDevice(dd, comArray)
			// logger.Debug.Printf("send array %d", dd.id)

			hout <- hs
		}
	}

}

//Считывает полученную информацию от устройства и распаковывет ее в контроллер
func updateController(c *pudge.Controller, hDev *transport.HeaderDevice) (transport.HeaderServer, bool) {
	dmess := hDev.ParseMessage()
	// logger.Info.Printf("Устройство %d прислало %v", hDev.ID, dmess)
	need := false
	mutex.Lock()
	d := devs[hDev.ID]
	c.LastOperation = time.Now()
	c.StatusConnection = pudge.Connected
	defer mutex.Unlock()
	hs := transport.CreateHeaderServer(0, int(hDev.Code))
	if hDev.Number != 0 {
		mss := make([]transport.SubMessage, 0)
		var ms transport.SubMessage
		ms.Set0x01Server(int(hDev.Number))
		mss = append(mss, ms)
		hs.UpackMessages(mss)
	}
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
			c.Base = false
		case 0x07:
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
			logger.Debug.Printf("Пришла строка лога id %d %v", hDev.ID, lg)
		case 0x0f:
			//Установление связи ДК v2
			need = true
			err := mes.Get0x0FDevice(c)
			if err != nil {
				logger.Error.Printf("При разборе команды 0x0f id %d %s", hDev.ID, err.Error())
			}
		case 0x10:
			need = true
			err := mes.Get0x10Device(c)
			if err != nil {
				logger.Error.Printf("При разборе команды 0x0f id %d %s", hDev.ID, err.Error())
			}
		case 0x11:
			//Состояние оборудования v2
			err := mes.Get0x11Device(c)
			if err != nil {
				logger.Error.Printf("При разборе команды 0x11 id %d %s", hDev.ID, err.Error())
			}
		case 0x12:
			//Состояние ДК v3
			err := mes.Get0x12Device(c)
			if err != nil {
				logger.Error.Printf("При разборе команды 0x12 id %d %s", hDev.ID, err.Error())
			}
		case 0x13:
			//Массив приявязки
			var ar pudge.ArrayPriv
			err := mes.Get0x13Device(&ar)
			if err != nil {
				logger.Error.Printf("При разборе команды 0x13 id %d %s", hDev.ID, err.Error())
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
		case 0x1D:
			//Состояние подключения
			need = true
			err := mes.Get0x1DDevice(c)
			if err != nil {
				logger.Error.Printf("При разборе команды 0x1D id %d %s", hDev.ID, err.Error())
			}
		case 0x1B:
			//Состояние ПСБ V1 перелается с адресом отправителя 0x7F
			//Ничего пока не делаем
		case 0x1C:
			//Состояние подтверждения перелается с адресом отправителя 0x7F
			//Ничего пока не делаем
		default:
			logger.Error.Printf("От %d неверная команда %x", hDev.ID, mes.Type)
			return hs, false
		}
	}
	pudge.SetController(c)
	return hs, need
}
func getController(id int) (*pudge.Controller, error) {
	//Вначале проверим на pudge
	ctrl := new(pudge.Controller)
	// logger.Info.Printf("Check reg for %d", id)
	c, is := pudge.GetController(id)
	if !is {
		//Нет на pudge теперь надо проверить среди регистрированн
		strKey := pudge.IsRegistred(id)

		if len(strKey) == 0 {
			return nil, fmt.Errorf("id %d не зарегистрирован", id)
		}
		pudge.SetDefault(ctrl, strKey)
		pudge.SetController(ctrl)
		// logger.Info.Printf("id %d reg on %s", id, strKey)
		return ctrl, nil
	}
	// logger.Info.Printf("Check reg for %d closed", id)
	ctrl = c
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
		return hs, fmt.Errorf("Неверная команда от АРМ для %d  %x ", dd.id, comARM.Command)
	}
	mss = append(mss, ms)
	hs.UpackMessages(mss)
	return hs, nil
}
func makeLocalOn(dd *device) transport.HeaderServer {
	dd.addNumber()
	var ms transport.SubMessage
	//Сообщение об отключении управления
	hs := transport.CreateHeaderServer(int(dd.NumServ), 0)
	mss := make([]transport.SubMessage, 0)
	ms.Set0x02Server(false)
	mss = append(mss, ms)
	hs.UpackMessages(mss)
	// hss = append(hss, hs)
	return hs

}
func makeLocalOff(dd *device) transport.HeaderServer {
	dd.addNumber()
	var ms transport.SubMessage
	//Сообщение об отключении управления
	hs := transport.CreateHeaderServer(int(dd.NumServ), 0)
	mss := make([]transport.SubMessage, 0)
	ms.Set0x02Server(true)
	mss = append(mss, ms)
	hs.UpackMessages(mss)
	// hss = append(hss, hs)
	return hs

}
func makeArrayToDevice(dd *device, comArray CommandArray) transport.HeaderServer {
	dd.addNumber()
	var ms transport.SubMessage
	hs := transport.CreateHeaderServer(int(dd.NumServ), 0)
	mss := make([]transport.SubMessage, 0)
	ms.SetArray(comArray.Number, comArray.NElem, comArray.Elems)
	mss = append(mss, ms)
	hs.UpackMessages(mss)
	return hs
}
