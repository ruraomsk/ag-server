package binding

import "reflect"

//Планы координации

//SetDK наборы планов координации для обеих ДК перекрестка
type SetDK struct {
	DK1 SetPk `json:"dk1"` // Наборы для ДК1
	DK2 SetPk `json:"dk2"` // Наборы для ДК2
}

//Compare сравнивание истина если равны
func (sd *SetDK) Compare(ss *SetDK) bool {
	return reflect.DeepEqual(sd, ss)
}

//SetPk набор планов координации перекрестка
type SetPk struct {
	DK     int     `json:"dk"`  //Номер ДК
	Pk     int     `json:"pk"`  //Номер программы
	TypePU int     `json:"tpu"` //Тип программы управления управления 0-ЛПУ (локальная) 1-ПК(координации)
	Tc     int     `json:"tc"`  //Время цикла программы
	Stages []Stage `json:"sts"` //Фазы переключения
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
