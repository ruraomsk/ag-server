package device

import (
	"time"

	"github.com/ruraomsk/TLServer/logger"
)

func (dev *Device) oneStep() {
	//Выполняем работу устройства
	if dev.ExtCtrl {
		return
	}
	if len(dev.Controller.Arrays) == 0 {
		return
	}
	if true {
		return
	}
	for _, ar := range dev.Controller.Arrays {
		err := dev.Arrays.SetArray(ar.Number, ar.NElem, ar.Array)
		if err != nil {
			logger.Error.Printf("массив %v %s", ar, err.Error())
			return
		}
	}
	err := dev.Arrays.IsCorrect()
	if err != nil {
		// logger.Error.Printf("id %d %s", dev.ID, err.Error())
		return
	}
	if dev.Controller.Base {
		dev.Controller.Base = false
		logger.Info.Printf("id %d перешел в режим управления", dev.ID)
	}

	month := time.Now().Month()
	day := time.Now().Day()
	w := time.Now().Weekday()
	n := dev.Arrays.MonthSets.MonthSets[month-1].Days[day-1]
	if w == 0 || n == 0 {
		// logger.Error.Printf("n=%d w=%d", n, w)
		return
	}
	nn := dev.Arrays.WeekSets.WeekSets[n-1].Days[w-1]
	find := false
	for _, sk := range dev.Arrays.DaySets.DaySets {
		if sk.Number == nn {
			find = true
			dev.CK = sk.Number
			dev.NK = n
			hour := time.Now().Hour()
			min := time.Now().Minute()
			hst := 0
			mst := 0
			dev.PK = 1
			for _, pk := range sk.Lines {
				if (hour >= hst && min >= mst) && (hour <= pk.Hour && min <= pk.Min) {
					dev.PK = pk.PKNom
					break
				}
				hst = pk.Hour
				mst = pk.Min
			}
		}
	}
	if !find {
		logger.Error.Printf("id %d не найдена суточная карта n=%d nn=%d w=%d %v", dev.ID, n, nn, w, dev.Arrays.WeekSets)
	}
	dev.Controller.PK = dev.PK
	dev.Controller.CK = dev.CK
	dev.Controller.NK = dev.NK
	dev.Controller.StatusCommandDU.IsDUDK1 = true
	dev.Controller.StatusCommandDU.IsNK = true
	dev.Controller.StatusCommandDU.IsPK = true
	dev.Controller.StatusCommandDU.IsCK = true
	dev.Controller.TechMode = 8
	dev.Controller.DK.DDK = 11
	dev.Controller.DK.FDK = 1
	dev.Controller.DK.RDK = 11
}
