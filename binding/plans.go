package binding

import "reflect"

//Планы координации

//SetDK наборы планов координации для обеих ДК перекрестка
type SetDK struct {
	DK1 []SetPk `json:"dk1"` // Наборы для ДК1
	DK2 []SetPk `json:"dk2"` // Наборы для ДК2
}

//Compare сравнивание истина если равны
func (sd *SetDK) Compare(ss *SetDK) bool {
	return reflect.DeepEqual(sd, ss)
}

//SetPk набор планов координации перекрестка
type SetPk struct {
	DK     int     `json:"dk"`    //Номер ДК
	Pk     int     `json:"pk"`    //Номер программы от 1 до 12
	TypePU int     `json:"tpu"`   //Тип программы управления управления 0-ЛПУ (локальная) 1-ПК(координации)
	Tc     int     `json:"tc"`    //Время цикла программы
	Sdvig  int     `json:"sdvig"` //Время цикла
	Stages []Stage `json:"sts"`   //Фазы переключения
}

//Stage описание одной фазы плана координации
type Stage struct {
	Nline  int `json:"l"`     //Номер строки
	Start  int `json:"start"` //Время начала фазы
	Number int `json:"num"`   //Номер фазы
	Tf     int `json:"tf"`    //Тип фазы 0 -простая
	// 1 - МГР
	// 2 - 1ТВП
	// 3 - 2ТВП
	// 4 - 1,2ТВП
	// 5 - Зам 1 ТВП
	// 6 - Зам 2 ТВП
	// 7 - Зам
	Len int `json:"len"` //Длительность секунд
}

//NewSetDK создание нового набора планов координации
func NewSetDK() *SetDK {
	r := new(SetDK)
	r.DK1 = make([]SetPk, 12)
	r.DK2 = make([]SetPk, 12)
	for n := range r.DK1 {
		r.DK1[n] = newSetPk(1, n+1)
		r.DK2[n] = newSetPk(2, n+1)
	}
	return r
}
func newSetPk(dk int, pk int) SetPk {
	r := new(SetPk)
	r.DK = dk
	r.Pk = pk
	r.Stages = make([]Stage, 12)
	return *r
}

//IsEmpty если план координации для данного ДК нулеыой то истина
func (sd *SetDK) IsEmpty(dk, pk int) bool {
	if dk == 1 {
		return isEmpty(sd.DK1, pk)
	}
	if dk == 2 {
		return isEmpty(sd.DK2, pk)
	}
	return true
}
func isEmpty(set []SetPk, p int) bool {
	pk := set[p-1]
	if pk.Tc != 0 || pk.Sdvig != 0 || pk.TypePU != 0 {
		return false
	}
	for _, st := range pk.Stages {
		if st.Nline != 0 || st.Start != 0 || st.Number != 0 || st.Tf != 0 || st.Len != 0 {
			return false
		}
	}
	return true
}
