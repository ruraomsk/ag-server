package binding

import (
	"fmt"
	"reflect"
)

//TimeDevice Описание времени устройства
type TimeDevice struct {
	TimeZone int  `json:"tz"`      //Смещение от Гринвича
	Summer   bool `json:"summer"`  // Есть ди переход на летнее время
	Journal  bool `json:"journal"` // true если получать журнал
	NoGprs   bool `json:"nogprs"`  // true если запрещено передвать журнал при GPRS
}

//Compare сравнивание истина если равны
func (td *TimeDevice) Compare(tt *TimeDevice) bool {
	return reflect.DeepEqual(td, tt)
}

//NewTimeDevice создает новое описание времени устройства
func NewTimeDevice() *TimeDevice {
	r := new(TimeDevice)
	r.TimeZone = 6
	return r
}

//FromBuffer переводит из массива кодов в структуру
func (td *TimeDevice) FromBuffer(buffer []int) error {
	if len(buffer) != 9 {
		return fmt.Errorf("неверная длина массива")
	}
	if buffer[0] != buffer[2] {
		return fmt.Errorf("несовпал номер массива на сервере и номер массива")
	}
	if buffer[2] != 21 {
		return fmt.Errorf("неверный номер массива")
	}
	td.TimeZone = buffer[5]
	td.Summer = buffer[6] != 0
	td.Journal = buffer[7]&2 != 0
	td.NoGprs = buffer[7]&4 != 0
	return nil
}

//IsEmpty возвращает истину если данный массив пустой
func (td *TimeDevice) IsEmpty() bool {
	return false
}

//ToBuffer переводит из структуры в буфер кодов
func (td *TimeDevice) ToBuffer() []int {
	r := make([]int, 9)
	r[0] = 21
	r[1] = 0
	r[2] = 21
	r[3] = 5
	r[4] = 1
	r[5] = td.TimeZone
	if td.Summer {
		r[6] = 1
	}
	if td.Journal {
		r[7] += 2
	}
	if td.NoGprs {
		r[7] += 4
	}
	return r
}
