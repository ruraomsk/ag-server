package binding

import (
	"fmt"

	"github.com/ruraomsk/ag-server/logger"
)

// Arrays масиссивы привязок
type Arrays struct {
	TypeDevice int `json:"type"` //Тип устройства 1 C12УСДК 2 УСДК 4 ДКА 8 ДТ СК
	SetupDK    SetupDK
	SetDK      SetDK
	MonthSets  MonthSets
	WeekSets   WeekSets
	DaySets    DaySets
	SetCtrl    SetCtrl
	SetTimeUse SetTimeUse
	TimeDivice TimeDevice `json:"timedev"`   //Настройки времени
	StatDefine StatDefine `json:"defstatis"` // Описание настройки сбора статистики
	PointSet   PointSet   `json:"pointset"`  //Точки сбора статистики
	UseInput   UseInput   `json:"useinput"`  //Назначение входов для сбора статистики
	MGRs       []MGR      `json:"mgrs"`      //Массив разрешенных МГР
}

// NewArrays создание нового
func NewArrays() *Arrays {
	r := new(Arrays)
	r.TypeDevice = 2
	r.StatDefine = *NewStatDefine()
	r.PointSet = *NewPointSet()
	r.UseInput = *NewUseInput()
	r.SetupDK = *NewSetupDK()
	r.SetDK = *NewSetDK()
	r.MonthSets = *NewYearSets()
	r.WeekSets = *NewWeekSets()
	r.DaySets = *NewDaySet()
	r.SetCtrl = *NewSetCtrl()
	r.SetTimeUse = *NewSetTimeUse()
	r.MGRs = make([]MGR, 0)
	r.DaySets.DaySets[0].Lines[0].PKNom = 1
	r.WeekSets.WeekSets[0].Days = []int{1, 1, 1, 1, 1, 1, 1}

	return r
}

// IsCorrect проверяет правильность массивов
func (ar *Arrays) IsCorrect() error {
	if ar.SetupDK.IsEmpty() {
		logger.Error.Printf("нет настроек ДК ")
		// return fmt.Errorf("нет настроек ДК1")
	}
	for _, ms := range ar.MonthSets.MonthSets {
		// if ms.IsEmpty() {
		// 	continue
		// }
		for _, d := range ms.Days {
			if ar.WeekSets.WeekSets[d-1].IsEmpty() {
				return fmt.Errorf("в месяце %d нет такой недели %d", ms.Number, d)
			}
			for _, on := range ar.WeekSets.WeekSets[d-1].Days {
				find := false
				for _, od := range ar.DaySets.DaySets {
					if od.Number == on {
						find = true
						break
					}
				}
				if !find {
					return fmt.Errorf("в месяце %d и неделе %d нет дня %d", ms.Number, d, on)
				}

			}
		}
	}
	find := false
	// find2 := false
	for i := 1; i < 13; i++ {
		if !ar.SetDK.IsEmpty(1, i) {
			find = true
		}
		// if ar.SetupDK2.ExtNum != 0 {
		// 	if !ar.SetDK.IsEmpty(2, i) {
		// 		find2 = true
		// 	}

		// }
	}
	if !find {
		return fmt.Errorf("нет планов координации для ДК ")
	}
	// if ar.SetupDK2.ExtNum != 0 && !find2 {
	// 	return fmt.Errorf("нет планов координации для ДК2 ")
	// }
	return nil
}

// SetArray принимает массивы привязки на устройтсве
func (ar *Arrays) SetArray(nom, nelem int, array []int) error {
	buffer := make([]int, len(array)+3)
	buffer[2] = nom
	// buffer[3] = len(array)
	// buffer[4] = nelem
	for i := 0; i < len(array); i++ {
		buffer[3+i] = array[i]
	}
	switch nom {
	case 14:
		buffer[0] = 14
		return ar.StatDefine.FromBuffer(buffer)
	case 15:
		buffer[0] = 15
		return ar.PointSet.FromBuffer(buffer)
	case 16:
		buffer[0] = 16
		return ar.UseInput.FromBuffer(buffer)
	case 21:
		buffer[0] = 21
		return ar.TimeDivice.FromBuffer(buffer)
	case 7:
		if nelem == 1 {
			buffer[0] = 40
			return ar.SetupDK.FromBuffer(buffer)
		}
	case 8:
		buffer[0] = 44 + buffer[4]
		return ar.WeekSets.FromBuffer(buffer)
	case 137:
		buffer[0] = 64 + buffer[4]
		return ar.DaySets.FromBuffer(buffer)
	case 22:
		buffer[0] = 84 + buffer[4]
		return ar.MonthSets.FromBuffer(buffer)
	case 133:
		buffer[0] = 99 + nelem
		return ar.SetDK.FromBuffer(buffer)
	case 23:
		buffer[0] = 148
		return ar.SetTimeUse.FromBuffer(buffer)
	case 20:
		buffer[0] = 157
		return ar.SetTimeUse.FromBuffer(buffer)
	case 24:
		buffer[0] = 149
		return ar.SetCtrl.FromBuffer(buffer)
	}
	return fmt.Errorf("нет такого массива %d", nom)
}
