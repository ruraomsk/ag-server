package device

import "github.com/ruraomsk/ag-server/logger"

import "time"

func (dev *Device) oneStep() {
	//Выполняем работу устройства
	if dev.Controller.Base {
		//Устройство в базовой настройке проверим можно ли переходить на массивы?
		for _, ar := range dev.Controller.Arrays {
			err := dev.Arrays.SetArray(ar.Number, ar.NElem, ar.Array)
			if err != nil {
				logger.Error.Printf("массив %v %s", ar, err.Error())
				return
			}
		}
		err := dev.Arrays.IsCorrect()
		if err != nil {
			logger.Error.Println(err.Error())
			return
		}
		dev.Controller.Base = false
	}
	month := time.Now().Month()
	day := time.Now().Day()
	w := time.Now().Weekday()
	n := dev.Arrays.MonthSets.MonthSets[month-1].Days[day-1]
	nn := dev.Arrays.NedelSets.NedelSets[n-1].Days[w-1]
	for _, sk := range dev.Arrays.DaySets.DaySets {
		if sk.Number == nn {
			dev.CK = sk.Number
			dev.NK = n
			hour := time.Now().Hour()
			min := time.Now().Minute()
			hst := 0
			mst := 0
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

}
