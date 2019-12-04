package transport

import (
	"reflect"
	"time"

	"fmt"

	"strconv"
)

//HeaderDevice Сообщение от устройства
type HeaderDevice struct {
	ID         int       //ID
	TypeDevice uint8     //Тип устройства
	Time       time.Time // Время сообщения
	Number     uint8     // Номер сообщения
	Code       uint8     // Код отправителя
	Message    []uint8   // Собственно сообщение без контрольной суммы
}

//Compare Сравнение
func (hd *HeaderDevice) Compare(hdd *HeaderDevice) bool {
	return reflect.DeepEqual(hd, hdd)
}

//HeaderServer Сообщение от сервера
type HeaderServer struct {
	IDServer uint8     //ID Сервера 0xa7 or 0x8D
	Time     time.Time // Время сообщения
	Number   uint8     // Номер сообщения
	Code     uint8     // Код отправителя
	Message  []uint8   // Собственно сообщение без контрольной суммы
}

//Compare Сравнение
func (hs *HeaderServer) Compare(hss *HeaderServer) bool {
	return reflect.DeepEqual(hs, hss)
}

//Parse разбор сообщения от устройства
func (d *HeaderDevice) Parse(buffer []byte) error {

	if !checkCRC(buffer, 19) {
		return fmt.Errorf("неверная контрольная сумма")
	}
	d.TypeDevice = buffer[0]
	id := make([]byte, 9)
	for i := 0; i < len(id); i++ {
		id[i] = buffer[i+1]
	}
	lid, err := strconv.Atoi(string(id))
	if err != nil {
		return fmt.Errorf("при разборе id %s", err.Error())
	}
	d.ID = lid
	d.Time = takeDate(buffer, 10)
	d.Number = buffer[16]
	d.Code = buffer[17]
	d.Message = make([]uint8, buffer[18]-2)
	for i := 0; i < len(d.Message); i++ {
		d.Message[i] = buffer[19+i]
	}
	return nil
}

//Parse разбор сообщения от сервера
func (s *HeaderServer) Parse(buffer []byte) error {
	if !checkCRC(buffer, 13) {
		return fmt.Errorf("неверная контрольная сумма")
	}
	s.IDServer = uint8(buffer[1])
	s.Time = takeDate(buffer, 4)
	s.Number = buffer[10]
	s.Code = buffer[11]
	s.Message = make([]uint8, buffer[12]-2)
	for i := 0; i < len(s.Message); i++ {
		s.Message[i] = buffer[13+i]
	}
	return nil
}

//MakeBuffer создает буфер для передачи полностью упакованный со всеми КС
func (d *HeaderDevice) MakeBuffer() []byte {
	buffer := make([]byte, 19+len(d.Message)+4)
	buffer[0] = d.TypeDevice
	str := []byte(fmt.Sprintf("%09d", d.ID))
	for i := 0; i < 9; i++ {
		buffer[i+1] = str[i]
	}
	putDate(d.Time, buffer, 10)
	d.Time = takeDate(buffer, 10)
	buffer[16] = d.Number
	buffer[17] = d.Code
	buffer[18] = uint8(len(d.Message) + 2)
	for i := 0; i < len(d.Message); i++ {
		buffer[i+19] = d.Message[i]
	}
	sumB, sumP := makeCRC(buffer, 19)
	pos := len(buffer) - 4
	buffer[pos] = uint8((sumB >> 8) & 0xff)
	buffer[pos+1] = uint8(sumB & 0xff)
	buffer[pos+2] = uint8((sumP >> 8) & 0xff)
	buffer[pos+3] = uint8(sumP & 0xff)

	return buffer
}

//MakeBuffer создает буфер для передачи полностью упакованный со всеми КС
func (s *HeaderServer) MakeBuffer() []byte {
	buffer := make([]byte, 13+len(s.Message)+4)
	buffer[1] = s.IDServer
	putDate(s.Time, buffer, 4)
	s.Time = takeDate(buffer, 4)
	buffer[10] = s.Number
	buffer[11] = s.Code
	buffer[12] = uint8(len(s.Message) + 2)
	for i := 0; i < len(s.Message); i++ {
		buffer[i+13] = s.Message[i]
	}
	sumB, sumP := makeCRC(buffer, 13)
	pos := len(buffer) - 4
	buffer[pos] = uint8((sumB >> 8) & 0xff)
	buffer[pos+1] = uint8(sumB & 0xff)
	buffer[pos+2] = uint8((sumP >> 8) & 0xff)
	buffer[pos+3] = uint8(sumP & 0xff)
	return buffer
}

func makeCRC(buffer []byte, lenHeader int) (sumB uint, sumP uint) {
	sumP = 0
	sumB = 0
	for i := 0; i < lenHeader; i++ {
		sumP += uint(buffer[i])
		sumP = sumP & 0xffff
	}
	len := int(buffer[lenHeader-3])
	for i := 0; i < len; i++ {
		sumP += uint(buffer[lenHeader+i])
		sumP = sumP & 0xffff
		sumB += uint(buffer[lenHeader+i])
		sumB = sumB & 0xffff
	}
	sumB = (sumB + 0x2756) & 0xffff
	sumP = (sumP + 0xe752) & 0xffff
	return sumB, sumP
}
func checkCRC(buffer []byte, lenHeader int) bool {
	sumB, sumP := makeCRC(buffer, lenHeader)
	len := int(buffer[lenHeader-1] - 2)
	tb := ((uint(buffer[lenHeader+len]) & 0xff) << 8) | uint(buffer[lenHeader+len+1]&0xff)
	tp := ((uint(buffer[lenHeader+len+2]) & 0xff) << 8) | uint(buffer[lenHeader+len+3]&0xff)
	if tb != sumB || tp != sumP {
		return false
	}
	return true
}

func takeDate(buffer []byte, pos int) time.Time {
	year := int(buffer[pos+2]) + 2000
	month := time.Month(int(buffer[pos+1]))
	day := int(buffer[pos])
	hour := int(buffer[pos+3])
	minut := int(buffer[pos+4])
	sec := int(buffer[pos+5])
	location, _ := time.LoadLocation("Local")
	return time.Date(year, month, day, hour, minut, sec, 0, location)

}
func putDate(t time.Time, buffer []byte, pos int) {
	year, month, day := t.Date()
	hour := t.Hour()
	min := t.Minute()
	sec := t.Second()
	buffer[pos] = uint8(day)
	buffer[pos+1] = uint8(month)
	buffer[pos+2] = uint8(year % 100)
	buffer[pos+3] = uint8(hour)
	buffer[pos+4] = uint8(min)
	buffer[pos+5] = uint8(sec)
}
