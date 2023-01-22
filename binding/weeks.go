package binding

import (
	"fmt"
	"reflect"
)

// WeekSets все недельные планы
type WeekSets struct {
	WeekSets []OneWeek `json:"wsets"`
}

func (w *WeekSets) GetWeek(wn int) []int {
	result := make([]int, 0)
	for _, ow := range w.WeekSets {
		if ow.Number == wn {
			return ow.Days
		}
	}
	return result
}

// Compare сравнивание истина если равны
func (ns *WeekSets) Compare(nn *WeekSets) bool {
	return reflect.DeepEqual(ns, nn)
}

// FromBuffer загружает недельный массив из буфера
func (ns *WeekSets) FromBuffer(buffer []int) error {
	n, err := weekFromBuffer(buffer)
	if err != nil {
		return err
	}
	ns.WeekSets[n.Number-1] = *n
	return nil
}

// NewWeekSets создает новый набор недельных карт
func NewWeekSets() *WeekSets {
	r := new(WeekSets)
	r.WeekSets = make([]OneWeek, 12)
	for index := 0; index < len(r.WeekSets); index++ {
		r.WeekSets[index] = *NewOneWeek(index + 1)
	}
	return r
}

// OneWeek Одна строка недельных планов
type OneWeek struct {
	Number int   `json:"num"`
	Days   []int `json:"days"`
}

func NewOneWeek(number int) *OneWeek {
	r := new(OneWeek)
	r.Number = number
	r.Days = make([]int, 7)
	for index := 0; index < len(r.Days); index++ {
		r.Days[index] = 0
	}
	return r
}

// IsEmpty возвращает истину если данный недельный массив пустой
func (nm *OneWeek) IsEmpty() bool {
	for _, d := range nm.Days {
		if d != 0 {
			return false
		}
	}
	return false
}

// ToBuffer переводит из недельного массива в буфер кодов
func (nm *OneWeek) ToBuffer() []int {
	r := make([]int, 12)
	r[0] = nm.Number + 44
	r[1] = 0
	r[2] = 8
	r[3] = 8
	r[4] = nm.Number
	for index := 0; index < len(nm.Days); index++ {
		r[5+index] = nm.Days[index]
	}
	return r
}

// nedelFromBuffer переводит из массива кодов в недельный массив
func weekFromBuffer(buffer []int) (*OneWeek, error) {
	if len(buffer) != 12 {
		return nil, fmt.Errorf("неверная длина массива")
	}
	if buffer[2] != 8 {
		return nil, fmt.Errorf("несовпал номер массива на сервере и номер массива")
	}
	if buffer[0]-44 <= 0 || buffer[0]-44 > 12 {
		return nil, fmt.Errorf("неверный номер массива")
	}
	nm := NewOneWeek(buffer[4])
	for index := 0; index < len(nm.Days); index++ {
		nm.Days[index] = buffer[5+index]
	}
	return nm, nil
}
