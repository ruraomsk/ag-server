package comm

import (
	"fmt"
	"net"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/ruraomsk/ag-server/binding"
	"github.com/ruraomsk/ag-server/debug"
	"github.com/ruraomsk/ag-server/extcon"
	"github.com/ruraomsk/ag-server/logger"
	"github.com/ruraomsk/ag-server/pudge"
	"github.com/ruraomsk/ag-server/setup"
	"github.com/ruraomsk/ag-server/transport"
)

var Devs map[int]*Device
var Mutex sync.Mutex

// var writeArch chan pudge.ArchStat
var sendPhases chan DevPhases

// var answare chan string
// var request chan int

//Это сервер коммуникации
//Слушает входящие сообщения и распределяет их на устройства

// GetChanArray возвращает канал для присылки массивов для данного устройства
func GetChanArray(id int) (chan []pudge.ArrayPriv, bool) {
	d, is := getDevice(id)
	if !is {
		return nil, false
	}
	return d.CommandArray, true
}

// StartListen основной вход сервер коммуникаций
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
	Devs = make(map[int]*Device)
	ln, err := net.Listen("tcp", ":"+strconv.Itoa(setup.Set.CommServer.Port))

	if err != nil {
		logger.Error.Printf("Ошибка открытия порта %s", err.Error())
		panic("Скорее всего запущен еще один сервер")
	}
	defer ln.Close()
	if setup.Set.CommServer.MaxCon == 0 {
		setup.Set.CommServer.MaxCon = 10000
	}
	for {
		sock, err := ln.Accept()
		if err != nil {
			logger.Error.Printf("Ошибка accept %s", err.Error())
			continue
		}
		if !pudge.Works {
			return
		}
		count++
		if count > setup.Set.CommServer.MaxCon {
			os.Exit(100)
		}
		if count%5 == 0 {
			logger.Info.Println("Входящих соединений", count)
		}
		go newConnect(sock)
	}
}
func recoveryPanic() {
	if recoveryMessage := recover(); recoveryMessage != nil {
		logger.Error.Printf("PANIC:%v", recoveryMessage)
		stackSlice := make([]byte, 32000)
		s := runtime.Stack(stackSlice, false)
		logger.Error.Printf("Trace\n%s ", stackSlice[0:s])
		os.Exit(-1)
	}
	// logger.Error.Println("Паники нет просто выход!")
	// os.Exit(0)
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
	//ctrl := new(pudge.Controller)
	var err error
	logger.Debug.Printf("Incoming %s", soc.RemoteAddr().String())
	hout := make(chan transport.HeaderServer, 1)
	hin := make(chan transport.HeaderDevice, 1)
	defer recoveryPanic()
	defer soc.Close()

	readTout := time.Duration((setup.Set.CommServer.TimeOutRead + 60) * int64(time.Second))
	controlTout := time.Duration((setup.Set.CommServer.TimeOutRead - 30) * int64(time.Second))
	writeTout := time.Duration(setup.Set.CommServer.TimeOutWrite * int64(time.Second))
	dd := new(Device)
	dd.LastToDevice = time.Now()
	dd.ErrorTCP = make(chan net.Conn)
	dd.Socket = soc
	dd.WaitNum = 0
suka:
	hDev, err := transport.GetOneMessage(soc)
	if err != nil {
		logger.Error.Print(err.Error())
		return
	}
	w := fmt.Sprintf("Устройствo %s подключается... номер %d", soc.RemoteAddr().String(), hDev.ID)
	logger.Info.Printf(w)
	debug.DebugChan <- debug.DebugMessage{ID: hDev.ID, Time: time.Now(), FromTo: false, Info: true, Buffer: []byte(w)}
	_, ok := getDevice(hDev.ID)
	if ok {
		//Остановим текущее
		killDevice(hDev.ID)
	}
	start := time.Now()
	ctrl, reg, err := getController(hDev.ID)
	if err != nil {
		logger.Error.Printf("Устройствo %s %s", soc.RemoteAddr().String(), err.Error())
		time.Sleep(time.Minute)
		goto suka
	}
	dmess := hDev.ParseMessage()
	flag := false
	//hren := false
	for _, m := range dmess {
		if m.Type == 0x1D {
			flag = true
			if defaultEthernet(hDev, ctrl) {
				_ = m.Get0x1DDevice(ctrl)
			}
			// logger.Info.Printf("Заголовок команда 0x1D id %d ", ctrl.ID)

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
			if defaultEthernet(hDev, ctrl) {
				_ = m.Get0x1BDevice(ctrl)
			}

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
		controlTout = time.Duration((int64(ctrl.Status.TObmen*60) - 30) * int64(time.Second))
		ctrl.TimeOut = int64(ctrl.Status.TObmen * 60)
	} else {
		ctrl.TimeOut = setup.Set.CommServer.TimeOutRead
	}
	if ctrl.TMax != 0 {
		writeTout = time.Duration(ctrl.TMax * int64(time.Second))
	} else {
		ctrl.TMax = setup.Set.CommServer.TimeOutWrite
		pudge.SetController(ctrl)
	}
	if time.Since(start) > time.Duration(10*time.Second) {
		logger.Info.Println("больше 10 секунд ", ctrl.ID)
	}
	//Обновим состояние в pudge
	ctrl.StatusConnection = true
	ctrl.LastMyOperation = time.Now()
	ctrl.ConnectTime = time.Now()
	dd.LastToDevice = time.Now()
	ctrl.IPHost = soc.RemoteAddr().String()
	dd.Id = ctrl.ID
	dd.NumDev = hDev.Code
	dd.Region = reg
	dd.CommandARM = make(chan pudge.CommandARM, 1024)
	dd.CommandArray = make(chan []pudge.ArrayPriv, 1024)
	dd.ChangeProtocol = make(chan ChangeProtocol)
	dd.ExitCommand = make(chan int, 10)
	dd.WaitNum = 0
	dd.CountLost = 0
	//dd.Messages=make(map[int]transport.HeaderServer)
	dd.addNumber()
	dd.context, _ = extcon.NewContext("device" + strconv.Itoa(dd.Id))
	Mutex.Lock()
	Devs[dd.Id] = dd
	Mutex.Unlock()
	updateController(ctrl, &hDev, dd)
	pudge.ChanLog <- pudge.LogRecord{ID: ctrl.ID, Region: dd.Region, Type: 1, Time: time.Now(), Journal: pudge.SetDeviceStatus(ctrl.ID)}
	if ctrl.Base {
		ctrl.Arrays = pudge.MakeArrays(*binding.NewArrays())
	}
	pudge.SetController(ctrl)
	//Запускаем ввод вывод
	go transport.GetMessagesFromDevice(soc, hin, &readTout, dd.ErrorTCP)
	go transport.SendMessagesToDevice(soc, hout, &writeTout, dd.ErrorTCP, dd.Id)
	//Подтвердим что клиент прописан
	var hs transport.HeaderServer
	hs = transport.CreateHeaderServer(0, 0)
	mss := make([]transport.SubMessage, 0)
	_ = hs.UpackMessages(mss)
	hout <- hs
	pudge.ChanLog <- pudge.LogRecord{ID: ctrl.ID, Region: dd.Region, Type: 1, Time: time.Now(), Journal: pudge.UserDeviceStatus("Сервер", -4, 0)}
	logger.Info.Printf("Подключено устройство: id %d ", ctrl.ID)
	time.Sleep(1 * time.Second)

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
	tick1hour := time.NewTicker(1 * time.Hour)
	tickControlTobm := time.NewTicker(controlTout)
	timer := extcon.SetTimerClock(time.Second)
	replay := time.NewTicker(20 * time.Second)
	for {
		select {
		case <-tick1hour.C:
			ctrl, is := pudge.GetController(dd.Id)
			if !is {
				logger.Error.Printf("id %d нет в базe", dd.Id)
				continue
			}
			if ctrl == nil {
				logger.Error.Printf("id %d нет в базe", dd.Id)
				continue
			}
			ctrl.Traffic.LastFromDevice1Hour = ctrl.Traffic.FromDevice1Hour
			ctrl.Traffic.LastToDevice1Hour = ctrl.Traffic.ToDevice1Hour
			ctrl.Traffic.FromDevice1Hour = 0
			ctrl.Traffic.ToDevice1Hour = 0
			pudge.SetController(ctrl)
		case hDev = <-hin:
			ctrl, _ = pudge.GetController(dd.Id)
			if ctrl == nil {
				logger.Error.Printf("id %d нет в базe", dd.Id)
				continue
			}
			dd.hDev = hDev
			dd.LastToDevice = time.Now()
			ctrl.Traffic.FromDevice1Hour += uint64(hDev.Length)
			lastBase := ctrl.Base
			hs, need := updateController(ctrl, &hDev, dd)
			if ctrl.Base && !lastBase {
				ctrl.Arrays = pudge.MakeArrays(*binding.NewArrays())
			}
			if dd.Id == setup.Set.CommServer.IdDebug {
				continue
			}
			// pudge.SetController(ctrl)
			if need {
				if hs.Code == 0x7f {
					l := 13 + len(hs.Message) + 4
					ctrl.Traffic.ToDevice1Hour += uint64(l)
					ctrl.LastMyOperation = time.Now()
					// dd.LastToDevice = time.Now()
					hout <- hs
					dd.LastMessage = hs
					dd.WaitNum = hs.Number
					dd.CountLost = 0
				} else {
					if hs.Number == 0 {
						//Это только ответ для Жени
						if dd.Messages.Size() != 0 {
							l := dd.Messages.Pop()
							hs.Number = l.Number
							mss := hs.SubMessages
							mss = append(mss, l.SubMessages...)
							hs.UpackMessages(mss)
						}
					}
					l := 13 + len(hs.Message) + 4
					ctrl.Traffic.ToDevice1Hour += uint64(l)
					ctrl.LastMyOperation = time.Now()
					// dd.LastToDevice = time.Now()
					hout <- hs
					dd.LastMessage = hs
					dd.WaitNum = hs.Number
					dd.CountLost = 0
				}
			}
			if dd.WaitNum == 0 {
				sendForWait(dd, hout)
			}
			pudge.SetController(ctrl)
		case errSocket := <-dd.ErrorTCP:
			if strings.Compare(errSocket.RemoteAddr().String(), dd.Socket.RemoteAddr().String()) != 0 {
				continue
			}
			ctrl, _ = pudge.GetController(dd.Id)
			if ctrl == nil {
				logger.Error.Printf("id %d нет в базe", dd.Id)
				continue
			}
			ctrl.StatusConnection = false
			ctrl.LastMyOperation = time.Now()
			pudge.SetController(ctrl)
			w := fmt.Sprintf("Контроллер %d ошибки обмена", dd.Id)
			pudge.ChanLog <- pudge.LogRecord{ID: ctrl.ID, Region: dd.Region, Type: 1, Time: time.Now(),
				Journal: pudge.UserDeviceStatus("Сервер", -3, 0)}
			logger.Error.Print(w)
			killDevice(dd.Id)
			timer.Stop()
			time.Sleep(time.Second)
			return
		case <-tickControlTobm.C:
			if dd.Messages.Size() == 0 {
				//logger.Debug.Printf("keepAlive %d", dd.id)
				hs, _ = makeAlive(dd)
				dd.Messages.Push(hs)
				sendForWait(dd, hout)
			}
		case <-timer.C:
			//Если я в Devs
			var is bool
			_, is = getDevice(dd.Id)
			if !is {
				// logger.Info.Printf("Добавляем %d", dd.Id)
				Mutex.Lock()
				Devs[dd.Id] = dd
				Mutex.Unlock()
				// logger.Info.Printf("Добавили %d", dd.Id)
			}
			if dd.Id == setup.Set.CommServer.IdDebug {
				continue
			}
			ctrl, is = pudge.GetController(dd.Id)
			if !is {
				break
			}
			if time.Since(dd.LastToDevice) > readTout {
				//Уже пять минут нет связи с устройством
				//Прощаемся с ним %-)
				ctrl.StatusConnection = false
				ctrl.LastMyOperation = time.Now()
				pudge.SetController(ctrl)
				w := fmt.Sprintf("Устройство %d более %f не выходит на связь ", dd.Id, readTout.Seconds())
				pudge.ChanLog <- pudge.LogRecord{ID: ctrl.ID, Region: dd.Region, Type: 1, Time: time.Now(),
					Journal: pudge.UserDeviceStatus("Сервер", -3, 0)}
				logger.Error.Print(w)
				killDevice(dd.Id)
				timer.Stop()
				time.Sleep(1 * time.Second)
				return
			}
			if ctrl.Status.StatusV220 != 0 {
				ctrl.StatusConnection = false
				ctrl.LastMyOperation = time.Now()
				w := ""
				ctrl.DK.EDK = 11
				if ctrl.Status.StatusV220 == 25 {
					w = fmt.Sprintf("Устройство %d авария 220В ", dd.Id)
					debug.DebugChan <- debug.DebugMessage{ID: dd.Id, Time: time.Now(), FromTo: false, Info: true, Buffer: []byte(w)}

					pudge.ChanLog <- pudge.LogRecord{ID: ctrl.ID, Region: dd.Region, Type: 1, Time: time.Now(), Journal: pudge.UserDeviceStatus("Сервер", -5, 0)}
					ctrl.DK.DDK = 3
				} else {
					w = fmt.Sprintf("Устройство %d выключено ", dd.Id)
					debug.DebugChan <- debug.DebugMessage{ID: dd.Id, Time: time.Now(), FromTo: false, Info: true, Buffer: []byte(w)}
					pudge.ChanLog <- pudge.LogRecord{ID: ctrl.ID, Region: dd.Region, Type: 1, Time: time.Now(), Journal: pudge.UserDeviceStatus("Сервер", -3, 0)}
					ctrl.DK.DDK = 5
				}
				pudge.SetController(ctrl)
				// pudge.ChanLog <- pudge.LogRecord{ID: ctrl.ID, Region: dd.Region, Type: 1, Time: time.Now(), Journal: pudge.SetDeviceStatus(ctrl.ID)}
				logger.Error.Print(w)
				killDevice(dd.Id)
				timer.Stop()
				time.Sleep(1 * time.Second)
				return
			}
		case <-replay.C:
			if dd.WaitNum != 0 {
				dd.CountLost++
				if dd.CountLost > 5 {
					ctrl, _ = pudge.GetController(dd.Id)
					ctrl.StatusConnection = false
					ctrl.LastMyOperation = time.Now()
					pudge.SetController(ctrl)
					logger.Info.Printf("Устройство %d более 5 раз не отвечает удаляем текущее подключение", dd.Id)
					debug.DebugChan <- debug.DebugMessage{ID: dd.Id, Time: time.Now(), FromTo: false, Info: true, Buffer: []byte("Удалено текущее подключения")}
					pudge.ChanLog <- pudge.LogRecord{ID: ctrl.ID, Region: dd.Region, Type: 1, Time: time.Now(),
						Journal: pudge.UserDeviceStatus("Сервер", -3, 0)}
					killDevice(dd.Id)
					timer.Stop()
					time.Sleep(1 * time.Second)
					return
				}
			} else {
				sendForWait(dd, hout)
			}
		case <-dd.context.Done():
			transport.Stoped = true
			pudge.ChanLog <- pudge.LogRecord{ID: ctrl.ID, Region: dd.Region, Type: 1, Time: time.Now(), Journal: pudge.UserDeviceStatus("Сервер", -2, 0)}
			pudge.ChanLog <- pudge.LogRecord{ID: ctrl.ID, Region: dd.Region, Type: 0, Time: time.Now(), Journal: pudge.UserTechStatus(ctrl.ID, "Сервер", -2, 0)}
			logger.Info.Printf("Устройство %d приказано умереть", dd.Id)
			debug.DebugChan <- debug.DebugMessage{ID: dd.Id, Time: time.Now(), FromTo: false, Info: true, Buffer: []byte("Приказано умереть")}
			pudge.ChanLog <- pudge.LogRecord{ID: ctrl.ID, Region: dd.Region, Type: 1, Time: time.Now(),
				Journal: pudge.UserDeviceStatus("Сервер", -3, 0)}
			timer.Stop()
			time.Sleep(1 * time.Second)
			return
		case <-dd.ExitCommand:
			//transport.Stoped = true
			ctrl, _ = pudge.GetController(dd.Id)
			ctrl.StatusConnection = false
			ctrl.LastMyOperation = time.Now()
			pudge.SetController(ctrl)
			// pudge.ChanLog <- pudge.RecLogCtrl{ID: ctrl.ID, Type: 1, Time: time.Now(), LogString: "Новое подключение"}
			logger.Info.Printf("Устройство %d удаляем текущее подключение", dd.Id)
			debug.DebugChan <- debug.DebugMessage{ID: dd.Id, Time: time.Now(), FromTo: false, Info: true, Buffer: []byte("Удалено текущее подключения")}
			pudge.ChanLog <- pudge.LogRecord{ID: ctrl.ID, Region: dd.Region, Type: 1, Time: time.Now(),
				Journal: pudge.UserDeviceStatus("Сервер", -3, 0)}
			killDevice(dd.Id)
			timer.Stop()
			time.Sleep(1 * time.Second)
			return
		case changeProtocol := <-dd.ChangeProtocol:
			hs, err := makeChangeProtocol(dd, changeProtocol)
			if err != nil {
				logger.Error.Printf("При создании команды измения протокола для %d %s", dd.Id, err.Error())
			} else {
				dd.Messages.Push(hs)
				sendForWait(dd, hout)
			}

		case comARM := <-dd.CommandARM:
			//Пришла команда арма
			// logger.Info.Printf("Для %d команда %v", dd.Id, comARM)
			hs, err = makeCommandToDevice(dd, comARM)
			if err != nil {
				logger.Error.Printf("При создании команды от АРМ %d %s", dd.Id, err.Error())
			} else {
				dd.Messages.Push(hs)
				sendForWait(dd, hout)
			}

		case comArray := <-dd.CommandArray:
			ctrl, _ = pudge.GetController(dd.Id)
			//Пришла команда арма загрузки привязки
			if comArray[0].Number == 0 {
				//Команда перейти в локальный режим
				hs := makeLocalOn(dd)
				dd.Messages.Push(hs)
				sendForWait(dd, hout)
				break
			}
			if comArray[0].Number == -1 {
				//Команда выйти из локального режима

				hs := makeLocalOff(dd)
				dd.Messages.Push(hs)
				sendForWait(dd, hout)
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
			dd.Messages.Push(hs)
			sendForWait(dd, hout)

		}
	}

}
func sendForWait(dd *Device, hout chan transport.HeaderServer) {
	if dd.WaitNum != 0 {
		return
	}
	if dd.Messages.Size() != 0 {
		ctrl, _ := pudge.GetController(dd.Id)
		if ctrl == nil {
			logger.Error.Printf("id %d нет в базe", dd.Id)
			return
		}

		dd.LastMessage = dd.Messages.Pop()
		dd.WaitNum = dd.LastMessage.Number
		l := 13 + len(dd.LastMessage.Message) + 4
		ctrl.Traffic.ToDevice1Hour += uint64(l)
		ctrl.LastMyOperation = time.Now()
		// logger.Debug.Printf("В простое передали на %d %v", dd.Id, dd.LastMessage.Message)
		hout <- dd.LastMessage
		dd.WaitNum = dd.LastMessage.Number
		dd.CountLost = 0
	}
}
func killDevice(id int) {
	// logger.Info.Printf("Удаляем контроллер %d", id)
	Mutex.Lock()
	delete(Devs, id)
	Mutex.Unlock()
	time.Sleep(2 * time.Second)
	// logger.Info.Printf("Удалили контроллер %d", id)
}
func getDevice(id int) (dev *Device, ok bool) {
	// logger.Info.Printf("Читаем контроллер %d", id)
	Mutex.Lock()
	dev, ok = Devs[id]
	Mutex.Unlock()
	// logger.Info.Printf("Прочитали контроллер %d %v", id, ok)
	return
}

// Считывает полученную информацию от устройства и распаковывет ее в контроллер
func updateController(c *pudge.Controller, hDev *transport.HeaderDevice, dd *Device) (transport.HeaderServer, bool) {
	var st pudge.Statistic
	var err error
	dmess := hDev.ParseMessage()
	// logger.Info.Printf("Устройство %d прислало %v", hDev.ID, dmess)
	need := false
	changeStatus := false
	//d := devs[hDev.ID]
	c.LastMyOperation = time.Now()
	c.LastOperation = time.Now()
	c.TimeDevice = hDev.Time
	c.StatusConnection = true
	hs := transport.CreateHeaderServer(0, int(hDev.Code))
	mss := make([]transport.SubMessage, 0)
	d, ok := getDevice(hDev.ID)
	if !ok {
		logger.Error.Printf("Устройство %d удалено и должно быть отключено", hDev.ID)
		return hs, false
	}
	if hDev.Number != 0 {
		var ms transport.SubMessage
		ms.Set0x01Server(int(hDev.Number))
		mss = append(mss, ms)
		_ = hs.UpackMessages(mss)
		hs.Number = 0
		need = true
	}
	for _, mes := range dmess {
		switch mes.Type {
		case 0x00:
			//Пустое сообщение ничего не делаем
		case 0x01:
			num, _, _, mas, elem := mes.Get0x01Device()
			if mas != 0 || elem != 0 {
				logger.Error.Printf("Ошибка массива привязки %d элемент %x", mas, elem)
			}
			if num != 0 {
				d.WaitNum = 0
			}
		case 0x04:
		// 	c.Base = false
		case 0x07:
			c.Base = true
			c.Arrays = pudge.MakeArrays(*binding.NewArrays())
		case 0x09:
			//Устройство закончило сбор статистики проверим если есть такая то обновим ее заголовок
			//если нет то добавим новый заголовок в массив статистики
			if !dd.StopStatistics {
				st, err = mes.Get0x09Device()
				if err != nil {
					logger.Error.Printf("Разбор x09 от устройства %d %s", hDev.ID, err.Error())
					continue
				}
			}
			//logger.Info.Printf("Пришла статистика от %d %02d:%02d",hDev.ID,st.Hour,st.Min)
		case 0x0a:
			//Собственно статистика пришла
			if dd.StopStatistics {
				continue
			}
			err = mes.Get0x0ADevice(&st)
			if err != nil {
				logger.Error.Printf("При разборе команды 0x0a id %d %s", hDev.ID, err.Error())
				continue
			}
			if st.TLen == 0 {
				for i := range c.Input.S {
					c.Input.S[i] = false
				}
			} else {
				for _, v := range st.Datas {
					if v.Chanel < len(c.Input.S) {
						c.Input.S[v.Chanel-1] = v.Status != 0
					}
				}
				pudge.StatisticChan <- pudge.RecordStat{Region: dd.Region, Stat: st}
			}
		case 0x0B:
			//Прием сохраненного журнала от устройства
			var lg pudge.LogLine
			err = mes.Get0x0BDevice(&lg)
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
			changeStatus = true

			if c.StatusCommandDU.IsReqSFDK1 || c.StatusCommandDU.IsReqSFDK2 {
				sendPhases <- DevPhases{ID: c.ID, DK: c.DK}
			}
		case 0x10:
			need = true
			err = mes.Get0x10Device(c)
			//logger.Info.Printf("Пришла команда 0x10 id %d ", hDev.ID)
			if err != nil {
				logger.Error.Printf("При разборе команды 0x10 id %d %s", hDev.ID, err.Error())
			}
		case 0x11:
			//Состояние оборудования v2
			err = mes.Get0x11Device(c)
			if err != nil {
				logger.Error.Printf("При разборе команды 0x11 id %d %s", hDev.ID, err.Error())
			} // logger.Debug.Printf("dev %d %v", hDev.ID, c.Error)
		case 0x12:
			//Состояние ДК v3
			err = mes.Get0x12Device(c)
			//logger.Debug.Printf("Команда 0x12 от %d Переход %d %b",hDev.ID,c.DK.EDK,c.DK.PDK)
			if err != nil {
				logger.Error.Printf("При разборе команды 0x12 id %d %s", hDev.ID, err.Error())
			}
			changeStatus = true
			if c.StatusCommandDU.IsReqSFDK1 || c.StatusCommandDU.IsReqSFDK2 {
				sendPhases <- DevPhases{ID: c.ID, DK: c.DK}
			}
		case 0x13:
			//Массив привязки
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
			if defaultEthernet(*hDev, c) {
				err := mes.Get0x1DDevice(c)
				if err != nil {
					logger.Error.Printf("При разборе команды 0x1D id %d %s", hDev.ID, err.Error())
				}
			}
		case 0x1B:
			need = true
			if defaultEthernet(*hDev, c) {
				err := mes.Get0x1BDevice(c)
				if err != nil {
					logger.Error.Printf("При разборе команды 0x1B id %d %s", hDev.ID, err.Error())
				}
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
	if changeStatus {
		pudge.ChanLog <- pudge.LogRecord{ID: c.ID, Region: dd.Region, Type: 1, Time: time.Now(), Journal: pudge.SetDeviceStatus(c.ID)}
		pudge.ChanLog <- pudge.LogRecord{ID: c.ID, Region: dd.Region, Type: 0, Time: time.Now(), Journal: pudge.SetTechStatus(c.ID)}
	}
	if need && hs.Code == 0x7f {
		hs = transport.CreateHeaderServer(0, 0)
		mss = make([]transport.SubMessage, 0)
		need = true
		_ = hs.UpackMessages(mss)
	}
	return hs, need
}
func getController(id int) (*pudge.Controller, pudge.Region, error) {
	//Вначале проверим на pudge
	key := pudge.GetRegion(id)

	if key.Region == 0 {
		return nil, pudge.Region{}, fmt.Errorf("id %d не зарегистрирован", id)
	}
	// logger.Info.Printf("Check reg for %d", id)
	c, is := pudge.GetController(id)
	if !is {
		ctrl := new(pudge.Controller)
		pudge.SetDefault(ctrl, key)
		pudge.SetController(ctrl)
		return ctrl, key, nil
	}

	return c, key, nil
}
func makeChangeProtocol(dd *Device, protocol ChangeProtocol) (transport.HeaderServer, error) {
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
	// logger.Debug.Printf("Header %v", hs)
	return hs, nil
}
func makeAlive(dd *Device) (transport.HeaderServer, error) {
	dd.addNumber()
	hs := transport.CreateHeaderServer(int(dd.NumServ), int(dd.hDev.Code))
	mss := make([]transport.SubMessage, 0)
	var ms transport.SubMessage
	ms.Set0x03Server()
	mss = append(mss, ms)
	_ = hs.UpackMessages(mss)
	return hs, nil

}
func makeCommandToDevice(dd *Device, comARM pudge.CommandARM) (transport.HeaderServer, error) {
	dd.addNumber()
	hs := transport.CreateHeaderServer(int(dd.NumServ), int(dd.hDev.Code))
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
	case 0x0E:
		//Передача МГР на устройство
		ms.Set0x0EServer(comARM.Params)
	case 0x0F:
		//Есть данные/нет данных по МГР
		dd.StopStatistics = comARM.Params == 1
		ms.Set0x0FServer(comARM.Params)

	default:
		return hs, fmt.Errorf("неверная команда для %d  %d ", dd.Id, comARM.Command)
	}
	mss = append(mss, ms)
	_ = hs.UpackMessages(mss)
	//mutex.Lock()
	////dd.Messages[int(dd.NumServ)]=hs
	//mutex.Unlock()
	return hs, nil
}
func makeLocalOn(dd *Device) transport.HeaderServer {
	dd.addNumber()
	var ms transport.SubMessage
	//Сообщение об отключении управления
	hs := transport.CreateHeaderServer(int(dd.NumServ), int(dd.hDev.Code))
	mss := make([]transport.SubMessage, 0)
	ms.Set0x04Server(false, false)
	mss = append(mss, ms)
	ms.Set0x02Server(false)
	mss = append(mss, ms)
	_ = hs.UpackMessages(mss)
	// hss = append(hss, hs)
	//mutex.Lock()
	//dd.Messages[int(dd.NumServ)]=hs
	//mutex.Unlock()
	return hs

}
func makeLocalOff(dd *Device) transport.HeaderServer {
	dd.addNumber()
	var ms transport.SubMessage
	//Сообщение об отключении управления
	hs := transport.CreateHeaderServer(int(dd.NumServ), int(dd.hDev.Code))
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
func makeArrayToDevice(dd *Device, comArrays []pudge.ArrayPriv) transport.HeaderServer {
	dd.addNumber()
	hs := transport.CreateHeaderServer(int(dd.NumServ), int(dd.hDev.Code))
	mss := make([]transport.SubMessage, 0)
	for _, arp := range comArrays {
		ms := new(transport.SubMessage)
		ms.SetArray(arp.Number, arp.NElem, arp.Array)
		mss = append(mss, *ms)
		// logger.Info.Printf("Передали на устройство %d привязку %v", dd.Id, arp.Array)
	}
	_ = hs.UpackMessages(mss)
	//mutex.Lock()
	////dd.Messages[int(dd.NumServ)]=hs
	//mutex.Unlock()
	return hs
}
func defaultEthernet(hDev transport.HeaderDevice, ctrl *pudge.Controller) bool {
	switch hDev.TypeDevice {
	case 0:
		ctrl.Status.Ethernet = false
		return false
	case 10:
		ctrl.Status.Ethernet = true
		return false
	case 20:
		ctrl.Status.Ethernet = false
		return false
	case 21:
		return true
	case 30:
		ctrl.Status.Ethernet = true
		return false
	}
	return true
}
