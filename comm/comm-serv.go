package comm

import (
	"fmt"
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/ruraomsk/TLServer/logger"
	"github.com/ruraomsk/ag-server/extcon"
	"github.com/ruraomsk/ag-server/pudge"
	"github.com/ruraomsk/ag-server/setup"
	"github.com/ruraomsk/ag-server/transport"
)

var devs map[int]*device
var mutex sync.Mutex

//var writeArch chan pudge.ArchStat
var sendPhases chan DevPhases

// var answare chan string
// var request chan int

//Это сервер коммуникации
//Слушает входящие сообщения и распределяет их на устройства

//GetChanArray возвращает канал для присылки массивов для данного устройства
func GetChanArray(id int) (chan []pudge.ArrayPriv, bool) {
	d, is := devs[id]
	if !is {
		return nil, false
	}
	return d.CommandArray, true
}

//StartListen основной вход сервер коммуникаций
func StartListen() {
	for !pudge.Works {
		time.Sleep(1 * time.Second)
	}

	//Запускаем слушателя для команд от АРМ
	go listenArmCommand()
	// //Запускаем слушателя для массивов привязки от АРМ
	go listenArmArray()
	// //Запускаем слушателя для настройки протокола
	go listenChangeProtocol()
	//writeArch = make(chan pudge.ArchStat, 1000)
	// Запускаем записывателя архива
	//go writerArch()
	// Запускаем посылку фаз
	sendPhases = make(chan DevPhases, 1000)
	go listenSendingPhazes()
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
		go newConnect(socket)
	}
}
func newConnect(soc net.Conn) {
	/*
			После установления соединения:
		1.Клиент отправляет сообщение Состояние ПБС V2, если ID клиента есть в БД сервера,
		сервер отвечает пустым сообщением 0x00, клиент, получив ответ, запускает процедуру
		начала обмена (п.2)
		2. В одном пакете клиент отправляет сообщения Установление соединения V2,Начало
		работы, Состояние оборудования , Состояние ДК V3, сервер отвечает
		подтверждением с номером принятого пакета.
	*/
	ctrl := new(pudge.Controller)
	var err error
	hout := make(chan transport.HeaderServer, 100)
	hin := make(chan transport.HeaderDevice, 100)
	defer soc.Close()
	readTout := time.Duration((setup.Set.CommServer.TimeOutRead + 60) * int64(time.Second))
	controlTout := time.Duration(setup.Set.CommServer.TimeOutRead * int64(time.Second))
	writeTout := time.Duration(setup.Set.CommServer.TimeOutWrite * int64(time.Second))
	dd := new(device)
	dd.ErrorTCP = make(chan int)
	go transport.GetMessagesFromDevice(soc, hin, &readTout, dd.ErrorTCP)
	go transport.SendMessagesToDevice(soc, hout, &writeTout, dd.ErrorTCP)
	hDev := <-hin
	logger.Debug.Printf("Устройствo %s подключается... номер %d", soc.RemoteAddr().String(), hDev.ID)
	mutex.Lock()
	ddd, ok := devs[hDev.ID]
	if ok {
		//Остановим текущее
		ddd.ExitCommand <- 1
		delete(devs, hDev.ID)
	}
	mutex.Unlock()
	start := time.Now()
	ctrl, err = getController(hDev.ID)
	if err != nil {
		logger.Error.Printf("Устройствo %s %s", soc.RemoteAddr().String(), err.Error())
		time.Sleep(3 * time.Minute)
		return
	}
	dmess := hDev.ParseMessage()
	flag := false
	//hren := false
	for _, m := range dmess {
		if m.Type == 0x1D {
			flag = true
			_ = m.Get0x1DDevice(ctrl)
			//logger.Info.Printf("Заголовок команда 0x1D id %d ", ctrl.ID)

		}
		if m.Type == 0x10 {
			flag = true
			_ = m.Get0x10Device(ctrl)
			//logger.Info.Printf("Заголовок команда 0x10 id %d ", ctrl.ID)

		}
		if m.Type == 0x12 {
			flag = true
			_ = m.Get0x12Device(ctrl)
			//logger.Info.Printf("Заголовок команда 0x12 id %d ", ctrl.ID)
		}
		if m.Type == 0x1B {
			flag = true
			//hren = true
			_ = m.Get0x1BDevice(ctrl)
			//logger.Info.Printf("Заголовок команда 0x1B id %d ", ctrl.ID)

		}
		if m.Type == 0x11 {
			flag = true
			_ = m.Get0x11Device(ctrl)
			//logger.Info.Printf("Заголовок команда 0x11 id %d ", ctrl.ID)
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
	if ctrl.Status.TObmen != 0 {
		readTout = time.Duration((int64(ctrl.Status.TObmen*60) + 60) * int64(time.Second))
		controlTout = time.Duration(int64(ctrl.Status.TObmen*60) * int64(time.Second))
		ctrl.TimeOut = int64(ctrl.Status.TObmen * 60)
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
	if time.Now().Sub(start) > time.Duration(10*time.Second) {
		logger.Info.Println("больше 10 секунд ", ctrl.ID)
	}
	//Обновим состояние в pudge
	ctrl.StatusConnection = true
	ctrl.LastOperation = time.Now()
	ctrl.IPHost = soc.RemoteAddr().String()
	dd.id = ctrl.ID
	dd.CommandARM = make(chan CommandARM)
	dd.CommandArray = make(chan []pudge.ArrayPriv)
	dd.ChangeProtocol = make(chan ChangeProtocol)
	dd.ExitCommand = make(chan int)
	//dd.Messages=make(map[int]transport.HeaderServer)
	dd.addNumber()
	dd.context, _ = extcon.NewContext("device" + strconv.Itoa(dd.id))
	mutex.Lock()
	devs[dd.id] = dd
	mutex.Unlock()
	updateController(ctrl, &hDev)
	pudge.ChanLog <- pudge.RecLogCtrl{ID: ctrl.ID, Type: -1, Time: time.Now(), LogString: "Подключен"}
	if ctrl.Base {
		ctrl.Arrays = make([]pudge.ArrayPriv, 0)
	}
	pudge.SetController(ctrl)
	//Подтвердим что клиент прописан
	var hs transport.HeaderServer
	hs = transport.CreateHeaderServer(0, 0)
	//if hren {
	//	hs = transport.CreateHeaderServer(0, 0)
	//
	//} else {
	//	hs = transport.CreateHeaderServer(0, int(hDev.Code))
	//}
	mss := make([]transport.SubMessage, 0)
	_ = hs.UpackMessages(mss)
	//time.Sleep(1 * time.Second)
	hout <- hs
	logger.Info.Printf("Подключено устройство: id %d ", ctrl.ID)
	//hs = transport.CreateHeaderServer(0, int(hDev.Code))
	//var ms transport.SubMessage
	//mss = make([]transport.SubMessage, 0)
	//ms.Set0x03Server()
	//mss = append(mss, ms)
	//_ = hs.UpackMessages(mss)
	//hout <- hs

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
	tick15min := time.NewTicker(15 * time.Minute)
	tick1hour := time.NewTicker(1 * time.Hour)
	timer := extcon.SetTimerClock(time.Duration(1 * time.Second))
	for {
		select {
		case <-tick1hour.C:
			ctrl.Traffic.LastFromDevice1Hour = ctrl.Traffic.FromDevice1Hour
			ctrl.Traffic.LastToDevice1Hour = ctrl.Traffic.ToDevice1Hour
			ctrl.Traffic.FromDevice1Hour = 0
			ctrl.Traffic.ToDevice1Hour = 0
			pudge.SetController(ctrl)
		case <-tick15min.C:
			ctrl.Traffic.LastFromDevice15Min = ctrl.Traffic.FromDevice15Min
			ctrl.Traffic.LastToDevice15Min = ctrl.Traffic.ToDevice15Min
			ctrl.Traffic.FromDevice15Min = 0
			ctrl.Traffic.ToDevice15Min = 0
			pudge.SetController(ctrl)
		case hDev = <-hin:
			ctrl.Traffic.FromDevice15Min += uint64(hDev.Length)
			ctrl.Traffic.FromDevice1Hour += uint64(hDev.Length)
			lastBase := ctrl.Base
			hs, need := updateController(ctrl, &hDev)
			if ctrl.Base && !lastBase {
				ctrl.Arrays = make([]pudge.ArrayPriv, 0)
			}
			if len(hs.Message) != 0 || need {
				l := 13 + len(hs.Message) + 4
				ctrl.Traffic.ToDevice15Min += uint64(l)
				ctrl.Traffic.ToDevice1Hour += uint64(l)
				ctrl.LastOperation = time.Now()
				hout <- hs
			} else {
				mutex.Lock()
				if dd.WaitNum != 0 {
					l := 13 + len(dd.LastMessage.Message) + 4
					ctrl.Traffic.ToDevice15Min += uint64(l)
					ctrl.Traffic.ToDevice1Hour += uint64(l)
					ctrl.LastOperation = time.Now()
					hout <- dd.LastMessage
					logger.Debug.Printf("Повторная передача на %d %v", dd.id, dd.LastMessage.Message)

				} else {
					if dd.Messages.Size() != 0 {
						dd.LastMessage = dd.Messages.Pop()
						dd.WaitNum = dd.LastMessage.Number
						l := 13 + len(dd.LastMessage.Message) + 4
						ctrl.Traffic.ToDevice15Min += uint64(l)
						ctrl.Traffic.ToDevice1Hour += uint64(l)
						ctrl.LastOperation = time.Now()
						hout <- dd.LastMessage
						logger.Debug.Printf("Передача на ответ устройства на %d %v", dd.id, dd.LastMessage.Message)
					}
				}
				mutex.Unlock()
			}
			pudge.SetController(ctrl)
		case <-dd.ErrorTCP:
			ctrl.StatusConnection = false
			pudge.SetController(ctrl)
			w := fmt.Sprintf("Устройство %d отключается ошибки ввода/вывода ", dd.id)
			pudge.ChanLog <- pudge.RecLogCtrl{ID: ctrl.ID, Type: -1, Time: time.Now(), LogString: w}
			logger.Error.Print(w)
			mutex.Lock()
			delete(devs, ctrl.ID)
			mutex.Unlock()
			return
		case <-timer.C:
			if time.Now().Sub(ctrl.LastOperation) < readTout && time.Now().Sub(ctrl.LastOperation) > controlTout {
				hs, err = makeAlive(dd)
				if err != nil {
					logger.Error.Printf("При создании команды keepALive %d %s", dd.id, err.Error())
					continue
				}
				hout <- hs
			}
			if time.Now().Sub(ctrl.LastOperation) > readTout {
				//Уже пять минут нет связи с устройством
				//Прощаемся с ним %-)
				ctrl.StatusConnection = false
				pudge.SetController(ctrl)
				w := fmt.Sprintf("Устройство %d более %f не выходит на связь ", dd.id, readTout.Seconds())
				pudge.ChanLog <- pudge.RecLogCtrl{ID: ctrl.ID, Type: -1, Time: time.Now(), LogString: w}
				logger.Error.Print(w)
				mutex.Lock()
				delete(devs, ctrl.ID)
				mutex.Unlock()
				return
			}
			if ctrl.Status.StatusV220 != 0 {
				ctrl.StatusConnection = false
				w := ""
				ctrl.DK.EDK = 11
				if ctrl.Status.StatusV220 == 25 {
					w = fmt.Sprintf("Устройство %d авария 220В ", dd.id)
					ctrl.DK.DDK = 3
				} else {
					w = fmt.Sprintf("Устройство %d выключено ", dd.id)
					ctrl.DK.DDK = 5
				}
				pudge.SetController(ctrl)
				pudge.ChanLog <- pudge.RecLogCtrl{ID: ctrl.ID, Type: -1, Time: time.Now(), LogString: w}
				logger.Error.Print(w)
				mutex.Lock()
				delete(devs, ctrl.ID)
				mutex.Unlock()
				return
			}
			mutex.Lock()
			if dd.WaitNum == 0 && dd.Messages.Size() != 0 {
				dd.LastMessage = dd.Messages.Pop()
				dd.WaitNum = dd.LastMessage.Number
				l := 13 + len(dd.LastMessage.Message) + 4
				ctrl.Traffic.ToDevice15Min += uint64(l)
				ctrl.Traffic.ToDevice1Hour += uint64(l)
				ctrl.LastOperation = time.Now()
				logger.Debug.Printf("В простое передали на %d %v", dd.id, dd.LastMessage.Message)
				hout <- dd.LastMessage
				pudge.SetController(ctrl)
			}
			mutex.Unlock()
		case <-dd.context.Done():
			transport.Stoped = true
			pudge.ChanLog <- pudge.RecLogCtrl{ID: ctrl.ID, Type: -1, Time: time.Now(), LogString: "Остановлен сервер"}
			logger.Info.Printf("Устройство %d приказано умереть", dd.id)
			return
		case <-dd.ExitCommand:
			//transport.Stoped = true
			pudge.ChanLog <- pudge.RecLogCtrl{ID: ctrl.ID, Type: 1, Time: time.Now(), LogString: "Новое подключение"}
			logger.Info.Printf("Устройство %d появилось новое подключение", dd.id)
			return
		case changeProtocol := <-dd.ChangeProtocol:
			hs, err := makeChangeProtocol(dd, changeProtocol)
			if err != nil {
				logger.Error.Printf("При создании команды измения протокола для %d %s", dd.id, err.Error())
				continue
			}
			mutex.Lock()
			dd.Messages.Push(hs)
			mutex.Unlock()

		case comARM := <-dd.CommandARM:
			//Пришла команда арма
			hs, err = makeCommandToDevice(dd, comARM)
			if err != nil {
				logger.Error.Printf("При создании команды от АРМ %d %s", dd.id, err.Error())
				continue
			}
			mutex.Lock()
			dd.Messages.Push(hs)
			mutex.Unlock()

		case comArray := <-dd.CommandArray:
			//Пришла команда арма загрузки привязки
			if comArray[0].Number == 0 {
				//Команда перейти в локальный режим
				hs := makeLocalOn(dd)
				mutex.Lock()
				dd.Messages.Push(hs)
				mutex.Unlock()
				break
			}
			if comArray[0].Number == -1 {
				//Команда выйти из локального режима

				hs := makeLocalOff(dd)
				mutex.Lock()
				dd.Messages.Push(hs)
				mutex.Unlock()
				break
			}
			for _, arp := range comArray {
				is := false
				for n, ap := range ctrl.Arrays {
					if ap.Number == arp.Number && ap.NElem == arp.NElem {
						ctrl.Arrays[n].Array = arp.Array
						is = true
						break
					}
				}
				if !is {
					ap := new(pudge.ArrayPriv)
					ap.Number = arp.Number
					ap.NElem = arp.NElem
					ap.Array = arp.Array
					ctrl.Arrays = append(ctrl.Arrays, *ap)
				}
			}
			pudge.SetController(ctrl)
			hs := makeArrayToDevice(dd, comArray)
			mutex.Lock()
			dd.Messages.Push(hs)
			mutex.Unlock()
		}
	}

}

//Считывает полученную информацию от устройства и распаковывет ее в контроллер
func updateController(c *pudge.Controller, hDev *transport.HeaderDevice) (transport.HeaderServer, bool) {
	dmess := hDev.ParseMessage()
	// logger.Info.Printf("Устройство %d прислало %v", hDev.ID, dmess)
	need := false
	mutex.Lock()
	defer mutex.Unlock()
	//d := devs[hDev.ID]
	c.LastOperation = time.Now()
	c.TimeDevice = hDev.Time
	c.StatusConnection = true
	d := devs[hDev.ID]
	hs := transport.CreateHeaderServer(0, 1)
	if hDev.Number != 0 {
		mss := make([]transport.SubMessage, 0)
		var ms transport.SubMessage
		ms.Set0x01Server(int(hDev.Number))
		mss = append(mss, ms)
		_ = hs.UpackMessages(mss)
	}
	for _, mes := range dmess {
		switch mes.Type {
		case 0x00:
			//Пустое сообщение ничего не делаем
		case 0x01:
			num, _, _, _, _ := mes.Get0x01Device()
			if num != 0 {
				if int(d.WaitNum) == num {
					d.WaitNum = 0
				}
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
			//logger.Info.Printf("Пришла статистика от %d %02d:%02d",hDev.ID,st.Hour,st.Min)
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
					//logger.Info.Printf("Пришла статистика %d %v",c.ID,stt)
					for _, s := range stt.Datas {
						if s.Chanel < 1 || s.Chanel > len(c.Input.S) {
							continue
						}
						if s.Status != 0 {
							c.Input.S[s.Chanel-1] = true
						} else {
							c.Input.S[s.Chanel-1] = false
						}
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
			if c.StatusCommandDU.IsReqSFDK1 || c.StatusCommandDU.IsReqSFDK2 {
				sendPhases <- DevPhases{ID: c.ID, DK: c.DK}
			}
		case 0x10:
			need = true
			err := mes.Get0x10Device(c)
			//logger.Info.Printf("Пришла команда 0x10 id %d ", hDev.ID)
			if err != nil {
				logger.Error.Printf("При разборе команды 0x10 id %d %s", hDev.ID, err.Error())
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
			//logger.Debug.Printf("Команда 0x12 от %d Переход %d %b",hDev.ID,c.DK.EDK,c.DK.PDK)
			if err != nil {
				logger.Error.Printf("При разборе команды 0x12 id %d %s", hDev.ID, err.Error())
			}
			if c.StatusCommandDU.IsReqSFDK1 || c.StatusCommandDU.IsReqSFDK2 {
				sendPhases <- DevPhases{ID: c.ID, DK: c.DK}
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
			need = true
			err := mes.Get0x1BDevice(c)
			if err != nil {
				logger.Error.Printf("При разборе команды 0x1B id %d %s", hDev.ID, err.Error())
			}
		case 0x1C:
			//Состояние подтверждения перелается с адресом отправителя 0x7F
			//Ничего пока не делаем
			//logger.Info.Printf("Пришла команда 0x1c id %d ", hDev.ID)
			need = true
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
	key := pudge.IsRegistred(id)

	if key == nil {
		return nil, fmt.Errorf("id %d не зарегистрирован", id)
	}
	// logger.Info.Printf("Check reg for %d", id)
	c, is := pudge.GetController(id)
	if !is {
		ctrl := new(pudge.Controller)
		pudge.SetDefault(ctrl, *key)
		pudge.SetController(ctrl)
		return ctrl, nil
	}

	return c, nil
}
func makeChangeProtocol(dd *device, protocol ChangeProtocol) (transport.HeaderServer, error) {
	dd.addNumber()
	hs := transport.CreateHeaderServer(int(dd.NumServ), 0x7f)
	mss := make([]transport.SubMessage, 0)
	var ms transport.SubMessage
	if protocol.F0x32 {
		_ = ms.Set0x32Server(protocol.IP, protocol.Port)
		mss = append(mss, ms)
	}
	if protocol.F0x33 {
		ms.Set0x33Server(protocol.Long)
		mss = append(mss, ms)
	}
	if protocol.F0x34 {
		ms.Set0x34Server(protocol.Type)
		mss = append(mss, ms)
	}
	if protocol.F0x35 {
		ms.Set0x35Server(protocol.Interval, protocol.Ignore)
		mss = append(mss, ms)
	}
	_ = hs.UpackMessages(mss)
	return hs, nil
}
func makeAlive(dd *device) (transport.HeaderServer, error) {
	//dd.addNumber()
	hs := transport.CreateHeaderServer(0, 1)
	mss := make([]transport.SubMessage, 0)
	var ms transport.SubMessage
	ms.Set0x03Server()
	mss = append(mss, ms)
	_ = hs.UpackMessages(mss)
	return hs, nil

}
func makeCommandToDevice(dd *device, comARM CommandARM) (transport.HeaderServer, error) {
	dd.addNumber()
	hs := transport.CreateHeaderServer(int(dd.NumServ), 1)
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
	_ = hs.UpackMessages(mss)
	mutex.Lock()
	//dd.Messages[int(dd.NumServ)]=hs
	mutex.Unlock()
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
	_ = hs.UpackMessages(mss)
	// hss = append(hss, hs)
	//mutex.Lock()
	//dd.Messages[int(dd.NumServ)]=hs
	//mutex.Unlock()
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
	_ = hs.UpackMessages(mss)
	// hss = append(hss, hs)
	//mutex.Lock()
	//dd.Messages[int(dd.NumServ)]=hs
	//mutex.Unlock()
	return hs

}
func makeArrayToDevice(dd *device, comArrays []pudge.ArrayPriv) transport.HeaderServer {
	dd.addNumber()
	hs := transport.CreateHeaderServer(int(dd.NumServ), 0)
	mss := make([]transport.SubMessage, 0)
	for _, arp := range comArrays {
		ms := new(transport.SubMessage)
		ms.SetArray(arp.Number, arp.NElem, arp.Array)
		mss = append(mss, *ms)
		logger.Info.Printf("Передали на устройство %d привязку %v", dd.id, arp.Array)
	}
	_ = hs.UpackMessages(mss)
	mutex.Lock()
	//dd.Messages[int(dd.NumServ)]=hs
	mutex.Unlock()
	return hs
}
