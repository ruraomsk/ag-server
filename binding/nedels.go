package binding

import (
	"fmt"
	"github.com/ruraomsk/ag-server/logger"
	"reflect"
)

//NedelSets все недельные планы
type NedelSets struct {
	NedelSets []OneNedel `json:"nsets"`
}

//Compare сравнивание истина если равны
func (ns *NedelSets) Compare(nn *NedelSets) bool {
	return reflect.DeepEqual(ns, nn)
}

//FromBuffer загружает недельный массив из буфера
func (ns *NedelSets) FromBuffer(buffer []int) error {
	n, err := nedelFromBuffer(buffer)
	if err != nil {
		return err
	}
	if n.Number < 1 {
		logger.Error.Printf("number <1 %v", buffer)
		return fmt.Errorf("number <1 %v", buffer)
	}

	ns.NedelSets[n.Number-1] = *n
	return nil
}

//NewNedelSets создает новый набор недельных карт
func NewNedelSets() *NedelSets {
	r := new(NedelSets)
	r.NedelSets = make([]OneNedel, 12)
	for index := 0; index < len(r.NedelSets); index++ {
		r.NedelSets[index] = *newOneNedel(index + 1)
	}
	return r
}

//OneNedel Одна строка недельных планов
type OneNedel struct {
	Number int   `json:"num"`
	Days   []int `json:"days"`
}

func newOneNedel(number int) *OneNedel {
	r := new(OneNedel)
	r.Number = number
	r.Days = make([]int, 7)
	for index := 0; index < len(r.Days); index++ {
		r.Days[index] = 0
	}
	return r
}

//IsEmpty возвращает истину если данный недельный массив пустой
func (nm *OneNedel) IsEmpty() bool {
	for _, d := range nm.Days {
		if d != 0 {
			return false
		}
	}
	return true
}

//ToBuffer переводит из недельного массива в буфер кодов
func (nm *OneNedel) ToBuffer() []int {
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

//nedelFromBuffer переводит из массива кодов в недельный массив
func nedelFromBuffer(buffer []int) (*OneNedel, error) {
	if len(buffer) != 12 {
		return nil, fmt.Errorf("неверная длина массива")
	}
	if buffer[2] != 8 {
		return nil, fmt.Errorf("несовпал номер массива на сервере и номер массива")
	}
	if buffer[0]-44 <= 0 || buffer[0]-44 > 12 {
		return nil, fmt.Errorf("неверный номер массива")
	}
	nm := newOneNedel(buffer[0] - 44)
	for index := 0; index < len(nm.Days); index++ {
		nm.Days[index] = buffer[5+index]
	}
	return nm, nil
}
