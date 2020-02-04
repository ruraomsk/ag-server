package inspect

import (
	"time"

	"github.com/ruraomsk/ag-server/comm"
	"github.com/ruraomsk/ag-server/extcon"
	"github.com/ruraomsk/ag-server/logger"
	"github.com/ruraomsk/ag-server/pudge"
)

var croses map[string]pudge.Region

//Start главный модуль инспектора
//Тут производится анализ правильности заполнения файлов привязки в контроллере
//и если не совпадает с перекрестком то записывается в контроллер
func Start(context *extcon.ExtContext, stop chan int) {
	// time.Sleep(1 * time.Minute)
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
	// logger.Info.Printf("запустили инспектора %v", reg)
main:
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
			logger.Info.Printf("контроллер не на связи %v %d", reg, cr.IDevice)
			time.Sleep(time.Duration(10 * time.Second))
			continue
		}
		if !dev.IsConnected() {
			//Контроллер не на связи проверим через минуту
			logger.Info.Printf("контроллер в ауте %v %d", reg, cr.IDevice)
			time.Sleep(time.Duration(10 * time.Second))
			continue
		}
		_, is = comm.GetChanArray(dev.ID)
		if !is {
			logger.Info.Printf("Нет канала слать массив на %d", dev.ID)
			time.Sleep(time.Duration(10 * time.Second))
			continue
		}

		//Построим массивы как надо для перекрестка
		crossarrays := makeArrays(cr)
		// logger.Info.Printf("массивы создали %v", reg)

		for _, ac := range crossarrays {
			found := false
			for _, d := range dev.Arrays {
				if d.Number == ac.Number && d.NElem == ac.NElem {
					if !d.Compare(&ac) {
						sendArray(dev, ac)
						continue main
					}
					found = true
				}
			}
			if !found {
				sendArray(dev, ac)
				continue main
			}
		}
		//Все переслали все совпало можно и поспать
		// logger.Info.Printf("все совпало %v", reg)
		// pudge.SetController(dev)
		time.Sleep(time.Duration(10 * time.Second))
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
			buffer := cr.Arrays.SetDK.DK1[i-1].ToBuffer() //
			r = appBuffer(r, buffer)
		}
		if !cr.Arrays.SetDK.IsEmpty(2, i) {
			buffer := cr.Arrays.SetDK.DK2[i-1].ToBuffer() //
			r = appBuffer(r, buffer)
		}
	}
	for _, ns := range cr.Arrays.NedelSets.NedelSets { //
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
func notZerro(buffer []int) bool {
	for i := 5; i < len(buffer); i++ {
		if buffer[i] != 0 {
			return true
		}
	}
	return false
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
	// if cmd.ID == 222222 {
	// 	logger.Debug.Printf("send %v", cmd)
	// }
	// logger.Info.Printf("послали массив на %v", cmd)
	ch <- *cmd
	// pudge.SetController(dev)
}
