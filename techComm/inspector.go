package techComm

import (
	"github.com/ruraomsk/TLServer/logger"
	"github.com/ruraomsk/ag-server/extcon"
	"github.com/ruraomsk/ag-server/memDB"
	"github.com/ruraomsk/ag-server/pudge"
	"github.com/ruraomsk/ag-server/setup"
	"time"
)

//Тут производится анализ правильности заполнения файлов привязки в контроллере
//и если не совпадает с перекрестком то записывается в контроллер
func getActiveCrosses() []string {
	memDB.TableDevices.Lock()
	defer memDB.TableDevices.Unlock()
	devs := memDB.GetListControllers()
	crosses := make([]string, 0)
	for _, id := range devs {
		ctrl, err := memDB.GetController(id)
		if err != nil {
			continue
		}
		if !ctrl.StatusConnection {
			continue
		}
		cross, err := memDB.GetCrossFromDevice(id)
		if err != nil {
			continue
		}
		reg := pudge.Region{Region: cross.Region, Area: cross.Area, ID: cross.ID}
		crosses = append(crosses, reg.ToKey())
	}
	return crosses
}

//Start главный модуль инспектора
func Start() {
	context, _ := extcon.NewContext("inspector")
	go WriterLog()
	crosses := getActiveCrosses()
	for _, key := range crosses {
		go oneCross(key)
	}
	inspectorTicker := time.NewTicker(5 * time.Second)
	statusTicker := time.NewTicker(time.Duration(setup.Set.StepPudge) * time.Second)
	select {
	case <-statusTicker.C:
		setStatusCross()
	case <-inspectorTicker.C:
		newcrosses := getActiveCrosses()
		for _, key := range newcrosses {
			found := false
			for _, oldkey := range crosses {
				if key == oldkey {
					found = true
					break
				}
			}
			if !found {
				go oneCross(key)
			}
		}
		crosses = newcrosses
	case <-context.Done():
		return
	}
}
func oneCross(key string) {
	logger.Info.Printf("запустили инспектора для %v", key)
	for {
		time.Sleep(3 * time.Second)
		cr, err := memDB.GetCross(key)
		if err != nil {
			//Перекресток удалили
			//logger.Info.Printf("удалили перекресток %s", key)
			return
		}

		dev, err := memDB.GetController(cr.IDevice)
		if err != nil {
			//Контроллер не выходил на связь
			return

		}
		if !dev.StatusConnection {
			//Контроллер не на связи
			return
		}
		if !isCorrectVersion(cr, dev) {
			//Не совпало ПО
			logger.Info.Printf("Не совпали версии ПО id %d %d.%d  %d.%d", dev.ID, dev.Model.VPCPDL, dev.Model.VPCPDR, cr.Model.VPCPDL, cr.Model.VPCPDR)
			time.Sleep(1 * time.Minute)
			continue
		}
		_, err = GetChanArray(dev.ID)
		if err != nil {
			time.Sleep(10 * time.Second)
			continue
		}

		//Построим массивы как надо для перекрестка
		crossArrays := makeArrays(cr)
		sending := make([]pudge.ArrayPriv, 0)
		for _, ac := range crossArrays {
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
			sendLocalOn(dev.ID)
			acc := make([]pudge.ArrayPriv, 0)
			l := 0
			for _, ac := range sending {
				if l < 65 {
					acc = append(acc, ac)
					l += len(ac.Array)
					continue
				}
				sendArray(dev.ID, acc)
				acc = make([]pudge.ArrayPriv, 0)
				acc = append(acc, ac)
				l = len(ac.Array)
			}
			if len(acc) > 0 {
				sendArray(dev.ID, acc)
			}
			pudge.ChanLog <- pudge.RecLogCtrl{ID: dev.ID, Type: -1, Time: time.Now(), LogString: "Обновлены привязки на устройстве"}
			sendLocalOff(dev.ID)
			logger.Info.Printf("массивы передали %d", dev.ID)

		}
		//Все переслали все совпало можно и поспать
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

func sendLocalOn(id int) {
	ch, err := GetChanArray(id)
	if err != nil {
		return
	}
	cmd := make([]pudge.ArrayPriv, 0)
	locOn := new(pudge.ArrayPriv)
	locOn.Number = 0
	cmd = append(cmd, *locOn)
	ch <- cmd
	pudge.ChanLog <- pudge.RecLogCtrl{ID: id, Type: -1, Time: time.Now(), LogString: "Начата передача массивов"}
}
func sendLocalOff(id int) {
	ch, err := GetChanArray(id)
	if err != nil {
		return
	}
	cmd := make([]pudge.ArrayPriv, 0)
	locOff := new(pudge.ArrayPriv)
	locOff.Number = -1
	cmd = append(cmd, *locOff)
	ch <- cmd
	pudge.ChanLog <- pudge.RecLogCtrl{ID: id, Type: -1, Time: time.Now(), LogString: "Окончена передача массивов"}
}
func sendArray(id int, array []pudge.ArrayPriv) {
	//Спросить у коммуникационного сервера канал для отправки сообщения
	ch, err := GetChanArray(id)
	if err != nil {
		return
	}
	ch <- array
}
func isCorrectVersion(cr pudge.Cross, dev pudge.Controller) bool {
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
