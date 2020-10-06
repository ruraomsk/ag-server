package inspect

//Тут производится анализ правильности заполнения файлов привязки в контроллере
//и если не совпадает с перекрестком то записывается в контроллер

import (
	"time"

	"github.com/JanFant/TLServer/logger"
	"github.com/ruraomsk/ag-server/comm"
	"github.com/ruraomsk/ag-server/extcon"
	"github.com/ruraomsk/ag-server/pudge"
)

var croses map[string]pudge.Region

//Start главный модуль инспектора
func Start(context *extcon.ExtContext, stop chan int) {
	// time.Sleep(1 * time.Minute)
	for !pudge.Works {
		time.Sleep(1 * time.Second)
	}
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
		time.Sleep(time.Duration(1 * time.Second))
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
		if dev.Local {
			//Беда у нас в прошлый обмен связь порвалась в опасном месте
			//Необходимо перепослать все массивы привязки
			dev.Arrays = make([]pudge.ArrayPriv, 0)
			dev.Local = false
			pudge.SetController(dev)
		}
		_, is = comm.GetChanArray(dev.ID)
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
						// logger.Info.Printf("не совпали\n%v\n%v\n", d, ac)
						sending = append(sending, ac)
						dev.Arrays[i] = ac
					}
				}
			}
			if !found {
				// logger.Info.Printf("не найден %v", ac)
				sending = append(sending, ac)
				dev.Arrays = append(dev.Arrays, ac)
			}
		}
		if len(sending) != 0 {
			// logger.Info.Printf("массивы создали %v", reg)
			dev.Local = true
			pudge.SetController(dev)
			sendLocalOn(dev)

			for _, ac := range sending {
				// logger.Info.Printf("id %d массив -> %v", dev.ID, ac)

				sendArray(dev, ac)
				time.Sleep(500 * time.Millisecond)
			}
			sendLocalOff(dev)
			dev.Local = false
			pudge.SetController(dev)
			logger.Info.Printf("массивы передали %d", dev.ID)
			pudge.ChanLog <- pudge.RecLogCtrl{ID: dev.ID, Type: -1, Time: time.Now(), LogString: "Обновлены привязки на устройстве"}

		}
		//Все переслали все совпало можно и поспать
		// logger.Info.Printf("все совпало %v", reg)
		// pudge.SetController(dev)

		// Посмотрим на статистику
		//if !reflect.DeepEqual(&cr.Statistics, &dev.Statistics){
		//	cr.Statistics=dev.Statistics
		//}
		time.Sleep(time.Duration(10 * time.Second))
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
		if !cr.Arrays.SetDK.IsEmpty(1, i) {
			buffer := cr.Arrays.SetDK.DK[i-1].ToBuffer() //
			r = appBuffer(r, buffer)
		}
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

//func notZerro(buffer []int) bool {
//	for i := 5; i < len(buffer); i++ {
//		if buffer[i] != 0 {
//			return true
//		}
//	}
//	return false
//}
func sendLocalOn(dev *pudge.Controller) {
	ch, is := comm.GetChanArray(dev.ID)
	if !is {
		logger.Info.Printf("Нет канала слать массив на %d", dev.ID)
		return
	}
	cmd := new(comm.CommandArray)
	cmd.ID = 0
	cmd.Number = 0
	ch <- *cmd
	pudge.ChanLog <- pudge.RecLogCtrl{ID: dev.ID, Type: -1, Time: time.Now(), LogString: "Начата передача массивов"}
}
func sendLocalOff(dev *pudge.Controller) {
	ch, is := comm.GetChanArray(dev.ID)
	if !is {
		logger.Info.Printf("Нет канала слать массив на %d", dev.ID)
		return
	}
	cmd := new(comm.CommandArray)
	cmd.ID = 0
	cmd.Number = 1
	ch <- *cmd
	pudge.ChanLog <- pudge.RecLogCtrl{ID: dev.ID, Type: -1, Time: time.Now(), LogString: "Окончена передача массивов"}
}
func sendArray(dev *pudge.Controller, array pudge.ArrayPriv) {
	//Спросить у коммуникационного сервера канал для отправки сообщения
	ch, is := comm.GetChanArray(dev.ID)
	if !is {
		logger.Info.Printf("Нет канала слать массив на %d", dev.ID)
		return
	}
	cmd := new(comm.CommandArray)
	cmd.ID = dev.ID
	cmd.Number = array.Number
	cmd.NElem = array.NElem
	cmd.Elems = array.Array
	ch <- *cmd
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
