package techComm

import (
	"fmt"
	"github.com/ruraomsk/TLServer/logger"
	"github.com/ruraomsk/ag-server/extcon"
	"github.com/ruraomsk/ag-server/memDB"
	"github.com/ruraomsk/ag-server/pudge"
	"github.com/ruraomsk/ag-server/setup"
	"github.com/ruraomsk/ag-server/transport"
	"time"
)

func (d *Device) Worker(hDev transport.HeaderDevice) {
	memDB.TableDevices.Lock()
	ctrl, err := memDB.GetController(d.ID)
	if err != nil {
		logger.Error.Printf("Пропало устройство %d", d.ID)
		memDB.TableDevices.Unlock()
		return
	}
	readTout := time.Duration((setup.Set.CommServer.TimeOutRead + 60) * int64(time.Second))
	controlTout := time.Duration(setup.Set.CommServer.TimeOutRead * int64(time.Second))
	writeTout := time.Duration(setup.Set.CommServer.TimeOutWrite * int64(time.Second))

	if ctrl.Status.TObmen != 0 {
		readTout = time.Duration((int64(ctrl.Status.TObmen*60) + 60) * int64(time.Second))
		controlTout = time.Duration(int64(ctrl.Status.TObmen*60) * int64(time.Second))
		ctrl.TimeOut = int64(ctrl.Status.TObmen * 60)
	} else {
		ctrl.TimeOut = setup.Set.CommServer.TimeOutRead
	}
	d.tOut = writeTout
	d.tIn = readTout
	d.LastToDevice = time.Now()
	d.Work = true
	d.updateController(&ctrl, &hDev)

	ctrl.StatusConnection = true
	ctrl.LastOperation = time.Now()
	ctrl.IPHost = d.Socket.RemoteAddr().String()

	go d.SendMessagesToDevice()
	go d.GetMessagesFromDevice()
	d.hOut <- accept()
	logger.Info.Printf("Подключено устройство: id %d ", ctrl.ID)
	memDB.SetController(ctrl)
	memDB.TableDevices.Unlock()
	tickerTobmen := time.NewTicker(controlTout)
	tick1h := time.NewTicker(1 * time.Hour)
	tick15min := time.NewTicker(15 * time.Minute)
	timer := extcon.SetTimerClock(1 * time.Second)
	defer func() {
		tickerTobmen.Stop()
		tick1h.Stop()
		tick15min.Stop()
		timer.Stop()
		d.Socket.Close()
		d.Work = false
		close(d.hIn)
		close(d.hOut)
		close(d.ChangeProtocol)
		close(d.ErrorTCP)
		close(d.CommandARM)
		close(d.CommandArray)
	}()
	for {
		select {
		case <-tick1h.C:
			d.Traffic.LastFromDevice1Hour = d.Traffic.FromDevice1Hour
			d.Traffic.LastToDevice1Hour = d.Traffic.ToDevice1Hour
			d.Traffic.FromDevice1Hour = 0
			d.Traffic.ToDevice1Hour = 0
		case <-tick15min.C:
			d.Traffic.LastFromDevice15Min = d.Traffic.FromDevice15Min
			d.Traffic.LastToDevice15Min = d.Traffic.ToDevice15Min
			d.Traffic.FromDevice15Min = 0
			d.Traffic.ToDevice15Min = 0
		case hDev = <-d.hIn:
			memDB.TableDevices.Lock()
			ctrl, _ = memDB.GetController(d.ID)
			lastBase := ctrl.Base
			hs, need := d.updateController(&ctrl, &hDev)
			if ctrl.Base && !lastBase {
				ctrl.Arrays = make([]pudge.ArrayPriv, 0)
			}
			d.CountLost = 0
			if len(hs.Message) != 0 || need {
				d.hOut <- hs
			} else {
				if d.WaitNum != 0 {
					d.CountLost = 0
					d.hOut <- d.LastMessage
					//logger.Debug.Printf("Повторная передача на %d %v", dd.id, dd.LastMessage.Message)

				} else {
					if d.Messages.Size() != 0 {
						d.LastMessage = d.Messages.Pop()
						d.WaitNum = d.LastMessage.Number
						d.CountLost = 0
						d.hOut <- d.LastMessage
						//logger.Debug.Printf("Передача на ответ устройства на %d %v", dd.id, dd.LastMessage.Message)
					} else {
						//logger.Debug.Printf("Нечего передавать на ответ устройства на %d", dd.id)
						d.CountLost = 0
					}
				}
			}
			memDB.SetController(ctrl)
			memDB.TableDevices.Unlock()
			d.lastOperation()
		case fl := <-d.ErrorTCP:
			txt := " при вводе c устройства"
			if fl == 0 {
				txt = " при выводе на устройство"
			}
			w := fmt.Sprintf("Контроллер %d отключается ошибки  %s", d.ID, txt)
			pudge.ChanLog <- pudge.RecLogCtrl{ID: ctrl.ID, Type: -1, Time: time.Now(), LogString: w}
			logger.Error.Print(w)
			extcon.StopForName(getNameContext(d.ID))
		case <-tickerTobmen.C:
			if d.Messages.Size() == 0 {
				//logger.Debug.Printf("keepAlive %d", dd.id)
				d.Messages.Push(d.makeAlive())
			}
		case <-timer.C:
			if time.Now().Sub(d.LastToDevice) > readTout {
				//Уже долго нет связи с устройством
				//Прощаемся с ним %-)
				w := fmt.Sprintf("Устройство %d более %f не выходит на связь ", d.ID, readTout.Seconds())
				pudge.ChanLog <- pudge.RecLogCtrl{ID: ctrl.ID, Type: -1, Time: time.Now(), LogString: w}
				logger.Error.Print(w)
				extcon.StopForName(getNameContext(d.ID))
				return
			}
			memDB.TableDevices.Lock()
			ctrl, _ := memDB.GetController(d.ID)
			if ctrl.Status.StatusV220 != 0 {
				w := ""
				ctrl.DK.EDK = 11
				if ctrl.Status.StatusV220 == 25 {
					w = fmt.Sprintf("Устройство %d авария 220В ", d.ID)
					ctrl.DK.DDK = 3
				} else {
					w = fmt.Sprintf("Устройство %d выключено ", d.ID)
					ctrl.DK.DDK = 5
				}
				pudge.ChanLog <- pudge.RecLogCtrl{ID: ctrl.ID, Type: -1, Time: time.Now(), LogString: w}
				logger.Error.Print(w)
			}
			ctrl.Traffic = d.Traffic
			memDB.SetController(ctrl)
			memDB.TableDevices.Unlock()
			if d.WaitNum == 0 && d.Messages.Size() != 0 {
				d.LastMessage = d.Messages.Pop()
				d.WaitNum = d.LastMessage.Number
				d.hOut <- d.LastMessage
				d.CountLost = 0
				d.lastOperation()
			} else {
				if d.WaitNum != 0 && d.Messages.Size() != 0 {
					d.CountLost++
					if d.CountLost > 10 {
						d.hOut <- d.LastMessage
						//logger.Debug.Printf("Повторная передача после 10 попыток на %d %v", dd.id, dd.LastMessage.Message)
						d.CountLost = 0
						d.lastOperation()
					}
				} else {
					d.CountLost = 0
				}
			}
		case <-d.Context.Done():
			memDB.TableDevices.Lock()
			ctrl, _ = memDB.GetController(d.ID)
			ctrl.StatusConnection = false
			memDB.SetController(ctrl)
			memDB.TableDevices.Unlock()
			d.Work = false
			deleteDevice(d.ID)
			pudge.ChanLog <- pudge.RecLogCtrl{ID: ctrl.ID, Type: -1, Time: time.Now(), LogString: "Остановлено устройство"}
			logger.Info.Printf("Устройство %d приказано умереть", d.ID)
			return
		case changeProtocol := <-d.ChangeProtocol:
			d.Messages.Push(d.makeChangeProtocol(changeProtocol))
		case comARM := <-d.CommandARM:
			//Пришла команда арма
			hs, err := d.makeCommandToDevice(comARM)
			if err != nil {
				logger.Error.Printf("При создании команды от АРМ %d %s", d.ID, err.Error())
				continue
			}
			d.Messages.Push(hs)

		case comArray := <-d.CommandArray:
			//Пришла команда арма загрузки привязки
			if comArray[0].Number == 0 {
				//Команда перейти в локальный режим
				d.Messages.Push(d.makeLocalOn())
				break
			}
			if comArray[0].Number == -1 {
				//Команда выйти из локального режима
				d.Messages.Push(d.makeLocalOff())
				break
			}
			memDB.TableDevices.Lock()
			ctrl, _ := memDB.GetController(d.ID)
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
			memDB.SetController(ctrl)
			memDB.TableDevices.Unlock()
			d.Messages.Push(d.makeArrayToDevice(comArray))
		}
	}

}
func accept() transport.HeaderServer {
	var hs transport.HeaderServer
	hs = transport.CreateHeaderServer(0, 0)
	mss := make([]transport.SubMessage, 0)
	_ = hs.UpackMessages(mss)
	return hs
}

//Считывает полученную информацию от устройства и распаковывет ее в контроллер
func (d *Device) updateController(c *pudge.Controller, hDev *transport.HeaderDevice) (transport.HeaderServer, bool) {
	dmess := hDev.ParseMessage()
	need := false
	c.LastOperation = time.Now()
	c.TimeDevice = hDev.Time
	c.StatusConnection = true
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
	return hs, need
}
func (d *Device) trafficIn(len int) {
	d.Traffic.FromDevice1Hour += uint64(len)
	d.Traffic.FromDevice15Min += uint64(len)
}
func (d *Device) trafficOut(len int) {
	d.Traffic.ToDevice15Min += uint64(len)
	d.Traffic.ToDevice1Hour += uint64(len)
}
func (d *Device) makeAlive() transport.HeaderServer {
	d.addNumber()
	hs := transport.CreateHeaderServer(0, 1)
	mss := make([]transport.SubMessage, 0)
	var ms transport.SubMessage
	ms.Set0x03Server()
	mss = append(mss, ms)
	_ = hs.UpackMessages(mss)
	return hs

}
func (d *Device) lastOperation() {
	d.LastToDevice = time.Now()
	memDB.TableDevices.Lock()
	ctrl, _ := memDB.GetController(d.ID)
	ctrl.LastOperation = time.Now()
	memDB.SetController(ctrl)
	memDB.TableDevices.Unlock()
}
func (d *Device) makeCommandToDevice(comARM CommandARM) (transport.HeaderServer, error) {
	d.addNumber()
	hs := transport.CreateHeaderServer(int(d.NumServ), 1)
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
		return hs, fmt.Errorf("Неверная команда от АРМ для %d  %x ", d.ID, comARM.Command)
	}
	mss = append(mss, ms)
	_ = hs.UpackMessages(mss)
	return hs, nil
}
func (d *Device) makeChangeProtocol(protocol ChangeProtocol) transport.HeaderServer {
	hs := transport.CreateHeaderServer(int(d.NumServ), 0x7f)
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
	return hs
}
func (d *Device) makeLocalOn() transport.HeaderServer {
	d.addNumber()
	var ms transport.SubMessage
	//Сообщение об отключении управления
	hs := transport.CreateHeaderServer(int(d.NumServ), 0)
	mss := make([]transport.SubMessage, 0)
	ms.Set0x02Server(false)
	mss = append(mss, ms)
	_ = hs.UpackMessages(mss)
	return hs

}
func (d *Device) makeLocalOff() transport.HeaderServer {
	d.addNumber()
	var ms transport.SubMessage
	//Сообщение об отключении управления
	hs := transport.CreateHeaderServer(int(d.NumServ), 0)
	mss := make([]transport.SubMessage, 0)
	ms.Set0x02Server(true)
	mss = append(mss, ms)
	_ = hs.UpackMessages(mss)
	return hs

}
func (d *Device) makeArrayToDevice(comArrays []pudge.ArrayPriv) transport.HeaderServer {
	d.addNumber()
	hs := transport.CreateHeaderServer(int(d.NumServ), 0)
	mss := make([]transport.SubMessage, 0)
	for _, arp := range comArrays {
		ms := new(transport.SubMessage)
		ms.SetArray(arp.Number, arp.NElem, arp.Array)
		mss = append(mss, *ms)
		logger.Info.Printf("Передали на устройство %d привязку %v", d.ID, arp.Array)
	}
	_ = hs.UpackMessages(mss)
	return hs
}
