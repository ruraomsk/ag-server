package binding

import "fmt"

//Arrays масиссивы привязок
type Arrays struct {
	SetupDK1   SetupDK
	SetupDK2   SetupDK
	SetDK      SetDK
	MonthSets  MonthSets
	NedelSets  NedelSets
	DaySets    DaySets
	TimeDivice TimeDevice `json:"timedev"`   //Настройки времени
	StatDefine StatDefine `json:"defstatis"` // Описание настройки сбора статистики
	PointSet   PointSet   `json:"pointset"`  //Точки сбора статистики
	UseInput   UseInput   `json:"useinput"`  //Назначение входов для сбора статистики
}

//NewArrays создание нового
func NewArrays() *Arrays {
	r := new(Arrays)
	r.StatDefine = *NewStatDefine()
	r.PointSet = *NewPointSet()
	r.UseInput = *NewUseInput()
	r.SetupDK1 = *NewSetupDK()
	r.SetupDK2 = *NewSetupDK()
	r.SetDK = *NewSetDK()
	r.MonthSets = *NewYearSets()
	r.NedelSets = *NewNedelSets()
	r.DaySets = *NewDaySet()
	return r
}
func (ar *Arrays) IsCorrect() bool {
	if ar.SetupDK1.IsEmpty() {
		return false
	}
	// if SetDK.
	return true
}

//SetArray принимает массивы привязки на устройтсве
func (ar *Arrays) SetArray(nom int, array []int) error {
	buffer := make([]int, array[0]+4)
	buffer[2] = nom
	for i := 0; i < len(array); i++ {
		buffer[3] = array[i]
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
	case 40:
		buffer[0] = 40
		return ar.SetupDK1.FromBuffer(buffer)
	case 41:
		buffer[0] = 41
		return ar.SetupDK2.FromBuffer(buffer)
	case 8:
		buffer[0] = 44 + buffer[4]
		return ar.NedelSets.FromBuffer(buffer)
	case 137:
		buffer[0] = 64 + buffer[4]
		return ar.DaySets.FromBuffer(buffer)
	case 22:
		buffer[0] = 84 + buffer[4]
		return ar.MonthSets.FromBuffer(buffer)
	case 133:
		if buffer[4] <= 12 {
			buffer[0] = 99 + buffer[4]
		} else {
			buffer[0] = buffer[4] - 9
		}
		return ar.SetDK.FromBuffer(buffer)

	}
	return fmt.Errorf("нет такого массива %d", nom)
}
