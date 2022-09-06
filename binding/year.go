package binding

import (
	"fmt"
	"reflect"
)

//MonthSets все месячные планы
type MonthSets struct {
	MonthSets []*OneMonth `json:"monthset"`
}

//Compare сравнивание истина если равны
func (ms *MonthSets) Compare(mm *MonthSets) bool {
	return reflect.DeepEqual(ms, mm)
}

//NewYearSets создает новый набор месячных карт
func NewYearSets() *MonthSets {
	r := new(MonthSets)
	r.MonthSets = make([]*OneMonth, 12)
	for index := 0; index < len(r.MonthSets); index++ {
		r.MonthSets[index] = newOneMonth(index + 1)
	}
	return r
}

//FromBuffer переводит из массива кодов в недельный массив
func (ms *MonthSets) FromBuffer(buffer []int) error {
	om, err := monthFromBuffer(buffer)
	if err != nil {
		return err
	}
	ms.MonthSets[om.Number-1] = om
	return nil
}

//OneMonth Одна строка недельных планов
type OneMonth struct {
	Number int   `json:"num"`
	Days   []int `json:"days"`
}

func newOneMonth(number int) *OneMonth {
	r := new(OneMonth)
	r.Number = number
	r.Days = make([]int, 31)
	for index := 0; index < len(r.Days); index++ {
		r.Days[index] = 1
	}
	return r
}

//IsEmpty возвращает истину если данный недельный массив пустой
func (om *OneMonth) IsEmpty() bool {
	// for _, d := range om.Days {
	// 	if d != 1 {
	// 		return false
	// 	}
	// }
	return false
}

//ToBuffer переводит из недельного массива в буфер кодов
func (om *OneMonth) ToBuffer() []int {
	r := make([]int, 36)
	r[0] = om.Number + 84
	r[1] = 0
	r[2] = 22
	r[3] = 32
	r[4] = om.Number
	for index := 0; index < len(om.Days); index++ {
		r[5+index] = om.Days[index]
	}
	return r
}

//monthFromBuffer переводит из массива кодов в недельный массив
func monthFromBuffer(buffer []int) (*OneMonth, error) {
	if len(buffer) != 36 {
		return nil, fmt.Errorf("неверная длина массива")
	}
	if buffer[2] != 22 {
		return nil, fmt.Errorf("несовпал номер массива на сервере и номер массива")
	}
	if buffer[0]-84 <= 0 || buffer[0]-84 > 12 {
		return nil, fmt.Errorf("неверный номер массива")
	}
	om := newOneMonth(buffer[0] - 84)
	for index := 0; index < len(om.Days); index++ {
		om.Days[index] = buffer[5+index]
	}
	return om, nil
}
