package binding

import (
	"fmt"
	"reflect"
)

//StatDefine настройка сбора статистики
type StatDefine struct {
	Levels []Level `json:"lvs"` //Два уровня сбора статистики
}

//Level один из двух уровней сбора
type Level struct {
	TypeSt int `json:"typst"`  //Признак статистики
	Period int `json:"period"` //Период усреднения в минутах
	Ninput int `json:"ninput"` //Колличество входов сбора статистики
	Count  int `json:"count"`  //Число направлений
}

//IsEmpty вернет истину если массив пустой
func (sd *StatDefine) IsEmpty() bool {
	for _, s := range sd.Levels {
		if s.TypeSt != 0 || s.Period != 0 || s.Ninput != 0 || s.Count != 0 {
			return false
		}
	}
	return false
}

//PointSet Точки сбора статистики
type PointSet struct {
	Points []Point `json:"pts"` //Описание точек сбора статистики

}

//Point Описание одной точки
type Point struct {
	NumPoint int `json:"num"`   //Номер точки
	TypeSt   int `json:"typst"` //Признак статистики
}

//IsEmpty вернет истину если массив пустой
func (ps *PointSet) IsEmpty() bool {
	for _, s := range ps.Points {
		if s.TypeSt != 0 || s.NumPoint != 0 {
			return false
		}
	}
	return false
}

//UseInput назначение входов для сбора статистики
type UseInput struct {
	Used []bool `json:"used"` //Признаки использования входа для сбора статистики
}

//IsEmpty вернет истину если массив пустой
func (ui *UseInput) IsEmpty() bool {
	// for _, s := range ui.Used {
	// 	if s {
	// 		return false
	// 	}
	// }
	return false
}

//Compare сравнивание истина если равны
func (sd *StatDefine) Compare(ss *StatDefine) bool {
	return reflect.DeepEqual(sd, ss)
}

//Compare сравнивание истина если равны
func (ps *PointSet) Compare(ss *PointSet) bool {
	return reflect.DeepEqual(ps, ss)
}

//Compare сравнивание истина если равны
func (ui *UseInput) Compare(ss *UseInput) bool {
	return reflect.DeepEqual(ui, ss)
}

//NewStatDefine создание новой
func NewStatDefine() *StatDefine {
	r := new(StatDefine)
	r.Levels = make([]Level, 2)
	return r
}

//NewPointSet создание новой
func NewPointSet() *PointSet {
	r := new(PointSet)
	r.Points = make([]Point, 0)
	return r
}

//NewUseInput создание новой
func NewUseInput() *UseInput {
	r := new(UseInput)
	r.Used = make([]bool, 0)
	return r
}

//ToBuffer сохранить в буфер
func (sd *StatDefine) ToBuffer() []int {
	buffer := make([]int, 13)
	buffer[0] = 14
	buffer[2] = 14
	buffer[3] = 9
	buffer[4] = 0
	pos := 5
	for _, l := range sd.Levels {
		buffer[pos] = l.TypeSt
		pos++
		buffer[pos] = l.Period
		pos++
		buffer[pos] = l.Ninput
		pos++
		buffer[pos] = l.Count
		pos++
	}
	return buffer
}

//ToBuffer сохранить в буфер
func (ps *PointSet) ToBuffer() []int {
	buffer := make([]int, (len(ps.Points)*2)+5)
	buffer[0] = 15
	buffer[2] = 15
	buffer[3] = (len(ps.Points) * 2) + 1
	buffer[4] = 0
	pos := 5
	for _, l := range ps.Points {
		buffer[pos] = l.NumPoint
		pos++
		buffer[pos] = l.TypeSt
		pos++
	}
	return buffer
}

//ToBuffer сохранить в буфер
func (ui *UseInput) ToBuffer() []int {
	buffer := make([]int, len(ui.Used)+5)
	buffer[0] = 16
	buffer[2] = 16
	buffer[3] = len(ui.Used) + 1
	buffer[4] = 0
	pos := 5
	for _, l := range ui.Used {
		buffer[pos] = 0
		if l {
			buffer[pos] = 1

		}
		pos++
	}
	return buffer
}

//FromBuffer заполнить из буфера
func (sd *StatDefine) FromBuffer(buffer []int) error {
	if len(buffer) != 13 {
		return fmt.Errorf("неверная длина массива")
	}
	if buffer[0] != buffer[2] {
		return fmt.Errorf("не совпал номер массива на сервере и номер массива")
	}
	if buffer[2] != 14 {
		return fmt.Errorf("неверный номер массива")
	}
	pos := 5
	for n := range sd.Levels {
		sd.Levels[n].TypeSt = buffer[pos]
		pos++
		sd.Levels[n].Period = buffer[pos]
		pos++
		sd.Levels[n].Ninput = buffer[pos]
		pos++
		sd.Levels[n].Count = buffer[pos]
		pos++
	}
	return nil
}

//FromBuffer заполнить из буфера
func (ps *PointSet) FromBuffer(buffer []int) error {
	// if len(buffer) != 17 {
	// 	return fmt.Errorf("неверная длина массива")
	// }
	if buffer[0] != buffer[2] {
		return fmt.Errorf("не совпал номер массива на сервере и номер массива")
	}
	if buffer[2] != 15 {
		return fmt.Errorf("неверный номер массива")
	}
	pos := 5
	ps.Points = make([]Point, (buffer[3]-1)/2)
	for n := range ps.Points {
		ps.Points[n].NumPoint = buffer[pos]
		pos++
		ps.Points[n].TypeSt = buffer[pos]
		pos++
	}
	return nil
}

//FromBuffer заполнить из буфера
func (ui *UseInput) FromBuffer(buffer []int) error {
	// if len(buffer) != 13 {
	// 	return fmt.Errorf("неверная длина массива")
	// }
	if buffer[0] != buffer[2] {
		return fmt.Errorf("не совпал номер массива на сервере и номер массива")
	}
	if buffer[2] != 16 {
		return fmt.Errorf("неверный номер массива")
	}
	pos := 5
	if buffer[3] <= 1 {
		ui.Used = make([]bool, 8)
		return nil
	}
	ui.Used = make([]bool, buffer[3]-1)
	for n := range ui.Used {
		ui.Used[n] = false
		if buffer[pos] == 1 {
			ui.Used[n] = true
		}
		pos++
	}
	return nil
}
