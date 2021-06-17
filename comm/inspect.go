package comm

//Тут производится анализ правильности заполнения файлов привязки в контроллере
//и если не совпадает с перекрестком то записывается в контроллер

import (
	"time"

	"github.com/ruraomsk/TLServer/logger"
	"github.com/ruraomsk/ag-server/extcon"
	"github.com/ruraomsk/ag-server/pudge"
)

var croses map[string]pudge.Region

//Start главный модуль инспектора
func Start(stop chan interface{}) {
	// time.Sleep(1 * time.Minute)
	for !pudge.Works {
		time.Sleep(1 * time.Second)
	}
	context, _ := extcon.NewContext("inspect")
	listCross := pudge.GetCrosses()
	croses = make(map[string]pudge.Region)
	for _, r := range listCross {
		_, is := pudge.GetCross(r.Region, r.Area, r.ID)
		if !is {
			logger.Error.Printf("Нет такого перекрестка %d %d ", r.Region, r.ID)
		}
		croses[r.ToKey()] = r
	}
	for _, r := range croses {
		go oneCross(r)
	}
	timer := extcon.SetTimerClock(time.Duration(5 * time.Second))
	select {
	case <-timer.C:
		listCross := pudge.GetCrosses()
		for _, r := range listCross {
			_, is := croses[r.ToKey()]

			if !is {
				croses[r.ToKey()] = r
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
	flagError := 0
	count := 0
	// logger.Info.Printf("запустили инспектора %v", reg)
	for {
		time.Sleep(time.Duration(3 * time.Second))
		cr, is := pudge.GetCross(reg.Region, reg.Area, reg.ID)
		if !is {
			//Перекресток удалили
			logger.Info.Printf("удалили перекресток %v", reg)
			return
		}
		dev, is := pudge.GetController(cr.IDevice)
		if !is {
			//Контроллер не выходил на связь проверим через минуту
			if flagError != 1 || count%100 == 0 {
				// logger.Info.Printf("контроллер не на связи %v %d", reg, cr.IDevice)
				flagError = 1
				count++
			}

			time.Sleep(time.Duration(10 * time.Second))
			continue
		}
		if !dev.IsConnected() {
			//Контроллер не на связи проверим через минуту
			if flagError != 2 || count%100 == 0 {
				// logger.Info.Printf("контроллер в ауте %v %d", reg, cr.IDevice)
				flagError = 2
				count++
			}
			time.Sleep(time.Duration(10 * time.Second))
			continue
		}
		if !isCorrectVersion(cr, dev) {
			//Не совпало ПО
			logger.Info.Printf("Не совпали версии ПО id %d %d.%d  %d.%d", dev.ID, dev.Model.VPCPDL, dev.Model.VPCPDR, cr.Model.VPCPDL, cr.Model.VPCPDR)
			time.Sleep(time.Duration(1 * time.Minute))
			continue
		}
		if cr.CK != dev.CK || cr.NK != dev.NK || cr.PK != dev.PK {
			pudge.Reload <- 0
			continue
		}
		//if dev.Local {
		//	//Беда у нас в прошлый обмен связь порвалась в опасном месте
		//	//Необходимо перепослать все массивы привязки
		//	logger.Info.Printf("Устройство %d не вышло из привязки!", dev.ID)
		//	dev.Arrays = make([]pudge.ArrayPriv, 0)
		//	dev.Local = false
		//	pudge.SetController(dev)
		//}
		_, is = GetChanArray(dev.ID)
		if !is {
			logger.Info.Printf("Нет канала слать массив на %d", dev.ID)
			time.Sleep(time.Duration(10 * time.Second))
			continue
		}

		//Построим массивы как надо для перекрестка
		// logger.Info.Printf("Проверяем %v", reg)
		crossarrays := makeArrays(cr)
		sending := make([]pudge.ArrayPriv, 0)
		for _, ac := range crossarrays {
			found := false
			for i, d := range dev.Arrays {
				if d.Number == ac.Number && d.NElem == ac.NElem {
					found = true
					if !d.Compare(&ac) {
						logger.Info.Printf("на устройстве %d не совпали\n%v\n%v\n", dev.ID, d, ac)
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
			//time.Sleep(time.Duration(1 * time.Second))
			sendLocalOn(dev)
			//time.Sleep(2 * time.Second)
			acc := make([]pudge.ArrayPriv, 0)
			l := 0
			for _, ac := range sending {
				if l < 65 {
					acc = append(acc, ac)
					l += len(ac.Array)
					continue
				}
				sendArray(dev, acc)
				//time.Sleep(2 * time.Second)
				acc = make([]pudge.ArrayPriv, 0)
				acc = append(acc, ac)
				l = len(ac.Array)
			}
			if len(acc) > 0 {
				sendArray(dev, acc)
				//time.Sleep(2 * time.Second)
			}
			pudge.ChanLog <- pudge.RecLogCtrl{ID: dev.ID, Type: -1, Time: time.Now(), LogString: "Обновлены привязки на устройстве"}
			sendLocalOff(dev)
			logger.Info.Printf("массивы передали %d", dev.ID)
			//time.Sleep(20 * time.Second)

		}
		//Все переслали все совпало можно и поспать
		flagError = 0
		count = 0
	}
}
func makeArrays(cr pudge.Cross) []pudge.ArrayPriv {
	r := make([]pudge.ArrayPriv, 0)
	if !cr.Arrays.StatDefine.IsEmpty() {
		buffer := cr.Arrays.StatDefine.ToBuffer() //
		r = appBuffer(r, buffer)
	}
	if !cr.Arrays.PointSet.IsEmpty() {
		buffer := cr.Arrays.PointSet.ToBuffer() //
		r = appBuffer(r, buffer)
	}
	if !cr.Arrays.UseInput.IsEmpty() {
		buffer := cr.Arrays.UseInput.ToBuffer() //
		r = appBuffer(r, buffer)

	}
	if !cr.Arrays.TimeDivice.IsEmpty() {
		buffer := cr.Arrays.TimeDivice.ToBuffer() //
		r = appBuffer(r, buffer)
	}
	if !cr.Arrays.SetupDK.IsEmpty() {
		buffer := cr.Arrays.SetupDK.ToBuffer() //
		r = appBuffer(r, buffer)
	}
	if !cr.Arrays.SetCtrl.IsEmpty() {
		buffer := cr.Arrays.SetCtrl.ToBuffer() //
		r = appBuffer(r, buffer)
	}
	if !cr.Arrays.SetTimeUse.IsEmpty() {
		buffer := cr.Arrays.SetTimeUse.ToBuffer(157) //
		r = appBuffer(r, buffer)
		buffer = cr.Arrays.SetTimeUse.ToBuffer(148) //
		r = appBuffer(r, buffer)
	}
	for i := 1; i < 13; i++ {
		r = appBuffer(r, cr.Arrays.SetDK.DK[i-1].ToBuffer())
	}
	for _, ns := range cr.Arrays.WeekSets.WeekSets { //
		if !ns.IsEmpty() {
			buffer := ns.ToBuffer()
			r = appBuffer(r, buffer)
		}
	}
	for _, ss := range cr.Arrays.DaySets.DaySets { //
		if !ss.IsEmpty() {
			buffer := ss.ToBuffer()
			r = appBuffer(r, buffer)
		}
	}
	for _, ys := range cr.Arrays.MonthSets.MonthSets { //
		if !ys.IsEmpty() {
			buffer := ys.ToBuffer()
			r = appBuffer(r, buffer)
		}
	}

	return r
}
func appBuffer(res []pudge.ArrayPriv, buffer []int) []pudge.ArrayPriv {
	return append(res, makePriv(buffer))
}
func makePriv(buffer []int) pudge.ArrayPriv {
	r := new(pudge.ArrayPriv)
	r.Array = make([]int, 0)
	r.Number = buffer[2]
	r.NElem = buffer[4]
	for i := 3; i < len(buffer); i++ {
		r.Array = append(r.Array, buffer[i])
	}
	return *r
}

func sendLocalOn(dev *pudge.Controller) {
	ch, is := GetChanArray(dev.ID)
	if !is {
		logger.Info.Printf("Нет канала слать массив на %d", dev.ID)
		return
	}
	cmd := make([]pudge.ArrayPriv, 0)
	locOn := new(pudge.ArrayPriv)
	locOn.Number = 0
	cmd = append(cmd, *locOn)
	ch <- cmd
	pudge.ChanLog <- pudge.RecLogCtrl{ID: dev.ID, Type: -1, Time: time.Now(), LogString: "Начата передача массивов"}
}
func sendLocalOff(dev *pudge.Controller) {
	ch, is := GetChanArray(dev.ID)
	if !is {
		logger.Info.Printf("Нет канала слать массив на %d", dev.ID)
		return
	}
	cmd := make([]pudge.ArrayPriv, 0)
	locOff := new(pudge.ArrayPriv)
	locOff.Number = -1
	cmd = append(cmd, *locOff)
	ch <- cmd
	pudge.ChanLog <- pudge.RecLogCtrl{ID: dev.ID, Type: -1, Time: time.Now(), LogString: "Окончена передача массивов"}
}
func sendArray(dev *pudge.Controller, array []pudge.ArrayPriv) {
	//Спросить у коммуникационного сервера канал для отправки сообщения
	ch, is := GetChanArray(dev.ID)
	if !is {
		logger.Info.Printf("Нет канала слать массив на %d", dev.ID)
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
