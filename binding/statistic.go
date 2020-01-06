package binding

import (
	"fmt"
	"reflect"
)

//StatDefine настройка сбора статистики
type StatDefine struct {
	Number int     `json:"num"` //Номер элемента массива 0 всегда
	Levels []Level `json:"lvs"` //Два уровня сбора статистики
}

//Level один из двух уровней сбора
type Level struct {
	TypeSt int `json:"typst"`  //Признак статистики
	Period int `json:"period"` //Период усреднения в минутах
	Ninput int `json:"ninput"` //Колличество входов сбора статистики
	Count  int `json:"count"`  //Число направлений
}

//PointSet Точки сбора статистики
type PointSet struct {
	Number int     `json:"num"` //Номер элемента массива 0 всегда
	Points []Point `json:"pts"` //Описание точек сбора статистики

}

//Point Описание одной точки
type Point struct {
	NumPoint int `json:"num"`   //Номер точки
	TypeSt   int `json:"typst"` //Признак статистики
}

//UseInput назначение входов для сбора статистики
type UseInput struct {
	Number int    `json:"num"`  //Номер элемента массива 0 всегда
	Used   []bool `json:"used"` //Признаки использования входа для сбора статистики
}

//Compare сравнивание истина если равны
func (st *StatDefine) Compare(ss *StatDefine) bool {
	return reflect.DeepEqual(st, ss)
}

//Compare сравнивание истина если равны
func (ps *PointSet) Compare(ss *PointSet) bool {
	return reflect.DeepEqual(ps, ss)
}

//Compare сравнивание истина если равны
func (us *UseInput) Compare(ss *UseInput) bool {
	return reflect.DeepEqual(us, ss)
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
	r.Used = make([]bool, 8)
	return r
}

//ToBuffer сохранить в буфер
func (st *StatDefine) ToBuffer() []int {
	buffer := make([]int, 13)
	buffer[0] = 14
	buffer[2] = 14
	buffer[3] = 9
	buffer[4] = st.Number
	pos := 5
	for _, l := range st.Levels {
		buffer[pos] = l.TypeSt
		pos++
		buffer[pos] = l.Period
		pos++
		buffer[pos] = l.Count
	}
	return buffer
}

//ToBuffer сохранить в буфер
func (ps *PointSet) ToBuffer() []int {
	buffer := make([]int, (len(ps.Points)*2)+5)
	buffer[0] = 15
	buffer[2] = 15
	buffer[3] = (len(ps.Points) * 2) + 1
	buffer[4] = ps.Number
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
func (us *UseInput) ToBuffer() []int {
	buffer := make([]int, 13)
	buffer[0] = 16
	buffer[2] = 16
	buffer[3] = 9
	buffer[4] = us.Number
	pos := 5
	for _, l := range us.Used {
		buffer[pos] = 0
		if l {
			buffer[pos] = 1

		}
		pos++
	}
	return buffer
}

//FromBuffer заполнить из буфера
func (st *StatDefine) FromBuffer(buffer []int) error {
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
	for n := range st.Levels {
		st.Levels[n].TypeSt = buffer[pos]
		pos++
		st.Levels[n].Period = buffer[pos]
		pos++
		st.Levels[n].Count = buffer[pos]
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
func (us *UseInput) FromBuffer(buffer []int) error {
	if len(buffer) != 13 {
		return fmt.Errorf("неверная длина массива")
	}
	if buffer[0] != buffer[2] {
		return fmt.Errorf("не совпал номер массива на сервере и номер массива")
	}
	if buffer[2] != 16 {
		return fmt.Errorf("неверный номер массива")
	}
	pos := 5
	for n := range us.Used {
		us.Used[n] = false
		if buffer[pos] == 1 {
			us.Used[n] = true
		}
		pos++
	}
	return nil
}
