package comm

//Тут производится анализ правильности заполнения файлов привязки в контроллере
//и если не совпадает с перекрестком то записывается в контроллер

import (
	"strconv"
	"time"

	"github.com/ruraomsk/ag-server/setup"

	"github.com/ruraomsk/ag-server/extcon"
	"github.com/ruraomsk/ag-server/logger"
	"github.com/ruraomsk/ag-server/pudge"
)

var croses map[pudge.Region]int

// Start главный модуль инспектора
func Start(stop chan interface{}) {
	// time.Sleep(1 * time.Minute)
	for !pudge.Works {
		time.Sleep(1 * time.Second)
	}
	setRegions := make(map[int]bool)
	for _, r := range setup.Set.Statistic.Regions {
		s, _ := strconv.Atoi(r[0])
		setRegions[s] = true
	}
	context, _ := extcon.NewContext("inspect")
	listCross := pudge.GetCrosses()
	croses = make(map[pudge.Region]int)
	for _, r := range listCross {
		_, i := setRegions[r.Region]
		if !i {
			continue
		}
		cr, is := pudge.GetCross(r)
		if !is {
			logger.Error.Printf("Нет такого перекрестка %d %d ", r.Region, r.ID)
		}
		croses[r] = cr.IDevice
	}
	for r := range croses {
		go oneCross(r)
	}
	timer := extcon.SetTimerClock(time.Duration(5 * time.Second))
	select {
	case <-timer.C:
		listCross := pudge.GetCrosses()
		for _, r := range listCross {
			_, i := setRegions[r.Region]
			if !i {
				continue
			}
			_, is := croses[r]

			if !is {
				cr, _ := pudge.GetCross(r)
				logger.Debug.Printf("новый перекресток %v", r)
				croses[r] = cr.IDevice
				go oneCross(r)
			}
		}
	case <-context.Done():
		return
	case <-stop:
		return
	}
}
func oneCross(reg pudge.Region) {
	needExit := 0
	// if reg.Area == 1 && (reg.ID == 11 || reg.ID == 7) {
	// 	logger.Info.Printf("запустили инспектора %v", reg)
	// }
	for {
		time.Sleep(time.Duration(1 * time.Second))

		cr, is := pudge.GetCross(reg)
		if !is {
			//Перекресток удалили
			logger.Info.Printf("удалили перекресток %v", reg)
			return
		}
		dev, is := pudge.GetController(cr.IDevice)
		if !is {
			//Контроллер не выходил на связь
			// if flagError != 1 || count%100 == 0 {
			// 	// logger.Info.Printf("контроллер не на связи %v %d", reg, cr.IDevice)
			// 	flagError = 1
			// 	count++
			// }

			continue
		}
		if !dev.IsConnected() {
			//Контроллер не на связи
			// if flagError != 2 || count%100 == 0 {
			// 	// logger.Info.Printf("контроллер в ауте %v %d", reg, cr.IDevice)
			// 	flagError = 2
			// 	count++
			// }
			continue
		}
		if !isCorrectVersion(cr, dev) {
			//Не совпало ПО
			logger.Info.Printf("Не совпали версии ПО id %d %d.%d  %d.%d", dev.ID, dev.Model.VPCPDL, dev.Model.VPCPDR, cr.Model.VPCPDL, cr.Model.VPCPDR)
			time.Sleep(time.Duration(1 * time.Minute))
			continue
		}
		_, is = GetChanArray(dev.ID)
		if !is {
			logger.Error.Printf("Нет канала слать массив на %d", dev.ID)
			time.Sleep(time.Duration(1 * time.Minute))
			continue
		}

		//Построим массивы как надо для перекрестка
		crossarrays := pudge.MakeArrays(cr.Arrays)
		sending := make([]pudge.ArrayPriv, 0)
		for _, ac := range crossarrays {
			found := false
			for i, d := range dev.Arrays {
				if d.Number == ac.Number && d.NElem == ac.NElem {
					found = true
					if !d.Compare(&ac) {
						logger.Info.Printf("на устройстве %v %d не совпали\n%v\n%v\n", reg, dev.ID, d, ac)
						sending = append(sending, ac)
						dev.Arrays[i] = ac
					}
				}
			}
			if !found {
				sending = append(sending, ac)
				dev.Arrays = append(dev.Arrays, ac)
			}
		}
		if len(sending) != 0 {
			sendLocalOn(dev)
			acc := make([]pudge.ArrayPriv, 0)
			l := 0
			for _, ac := range sending {
				if l < 65 {
					acc = append(acc, ac)
					l += len(ac.Array)
					continue
				}
				sendArray(dev, acc)
				acc = make([]pudge.ArrayPriv, 0)
				acc = append(acc, ac)
				l = len(ac.Array)
			}
			if len(acc) > 0 {
				sendArray(dev, acc)
			}
			sendLocalOff(dev)
			pudge.ChanLog <- pudge.LogRecord{ID: dev.ID, Region: reg, Type: 1, Time: time.Now(), Journal: pudge.UserDeviceStatus("Сервер", -1, 0)}

			logger.Info.Printf("массивы передали %d", dev.ID)

		}
		//Все переслали все совпало можно и поспать
		if dev.CalcStatus() == 60 { //24
			needExit++
			if needExit > 20 {
				d, ok := getDevice(dev.ID)
				if ok {
					//Остановим текущее
					d.ExitCommand <- 1
					killDevice(dev.ID)
					needExit = 0

				}
			}
		} else {
			needExit = 0
		}
	}
}

func sendLocalOn(dev *pudge.Controller) {
	ch, is := GetChanArray(dev.ID)
	if !is {
		logger.Error.Printf("Нет канала слать массив на %d", dev.ID)
		return
	}
	cmd := make([]pudge.ArrayPriv, 0)
	locOn := new(pudge.ArrayPriv)
	locOn.Number = 0
	cmd = append(cmd, *locOn)
	ch <- cmd
}
func sendLocalOff(dev *pudge.Controller) {
	ch, is := GetChanArray(dev.ID)
	if !is {
		logger.Error.Printf("Нет канала слать массив на %d", dev.ID)
		return
	}
	cmd := make([]pudge.ArrayPriv, 0)
	locOff := new(pudge.ArrayPriv)
	locOff.Number = -1
	cmd = append(cmd, *locOff)
	ch <- cmd
}
func sendArray(dev *pudge.Controller, array []pudge.ArrayPriv) {
	//Спросить у коммуникационного сервера канал для отправки сообщения
	ch, is := GetChanArray(dev.ID)
	if !is {
		logger.Error.Printf("Нет канала слать массив на %d", dev.ID)
		return
	}
	//cmd := new(comm.CommandArray)
	//cmd.ID = dev.ID
	//cmd.Number = array.Number
	//cmd.NElem = array.NElem
	//cmd.Elems = array.Array
	//logger.Debug.Printf("На устройство %d передали %v", dev.ID, cmd)
	ch <- array
}
func isCorrectVersion(cr pudge.Cross, dev *pudge.Controller) bool {
	c := cr.Model.VPCPDL*100 + cr.Model.VPCPDR
	d := dev.Model.VPCPDL*100 + dev.Model.VPCPDR
	if c <= 1203 && d <= 1203 {
		return true
	}
	if c >= 1204 && d >= 1204 {
		return true
	}
	return false
}
