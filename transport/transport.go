package transport

import (
	"reflect"
	"time"

	"github.com/ruraomsk/ag-server/logger"

	"fmt"

	"strconv"
)

// HeaderDevice Сообщение от устройства
type HeaderDevice struct {
	ID         int       //ID
	TypeDevice uint8     //Тип устройства
	Time       time.Time // Время сообщения
	Number     uint8     // Номер сообщения
	Code       uint8     // Код отправителя
	Length     int       // Размер в байтах
	Message    []uint8   // Собственно сообщение без контрольной суммы
}

// CreateHeaderDevice создает заголовок устройства
func CreateHeaderDevice(id, tp, number, code int) HeaderDevice {
	var h HeaderDevice
	h.ID = id
	h.TypeDevice = uint8(tp)
	h.Time = time.Now()
	h.Number = uint8(number)
	h.Code = uint8(code)
	h.Message = make([]uint8, 0)
	return h
}

// Compare Сравнение
func (d *HeaderDevice) Compare(dd *HeaderDevice) bool {
	return reflect.DeepEqual(d, dd)
}

// HeaderServer Сообщение от сервера
type HeaderServer struct {
	IDServer int       //ID Сервера 0xa7 0x8D
	Time     time.Time // Время сообщения
	Number   uint8     // Номер сообщения
	Code     uint8     // Код отправителя
	Message  []uint8   // Собственно сообщение без контрольной суммы
}

func (h *HeaderServer) Repeat() uint8 {
	// if h.Number == 0 {
	// 	return h.Number
	// }
	// h.Number++
	// if h.Number >= 250 {
	// 	h.Number = 1
	// }
	return h.Number
}

// CreateHeaderServer создает заголовок сервера
func CreateHeaderServer(num, code int) HeaderServer {
	var h HeaderServer
	h.IDServer = 0xa78d
	h.Time = time.Now()
	h.Number = uint8(num)
	h.Code = uint8(code)
	h.Message = make([]uint8, 0)
	// var ms SubMessage
	// mss := make([]SubMessage, 0)
	// ms.Set0x03Server()
	// mss = append(mss, ms)
	// h.UpackMessages(mss)

	return h
}

// Compare Сравнение
func (s *HeaderServer) Compare(ss *HeaderServer) bool {
	return reflect.DeepEqual(s, ss)
}

// Parse разбор сообщения от устройства
func (d *HeaderDevice) Parse(buffer []byte) error {
	d.Length = len(buffer)
	err := checkCRC(buffer, 19)
	if err != nil {
		return err
	}
	t := make([]byte, 2)
	for i := 0; i < 2; i++ {
		t[i] = buffer[i] + '0'
	}
	lt, err := strconv.Atoi(string(t))
	if err != nil {
		return fmt.Errorf("при разборе типа %s", err.Error())
	}
	d.TypeDevice = uint8(lt)
	id := make([]byte, 8)
	for i := 0; i < len(id); i++ {
		id[i] = buffer[i+2] + '0'
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

// Parse разбор сообщения от сервера
func (s *HeaderServer) Parse(buffer []byte) error {
	err := checkCRC(buffer, 13)
	if err != nil {
		return err
	}
	s.IDServer = 0xa78d
	s.Time = takeDate(buffer, 4)
	s.Number = buffer[10]
	s.Code = buffer[11]
	s.Message = make([]uint8, buffer[12]-2)
	for i := 0; i < len(s.Message); i++ {
		s.Message[i] = buffer[13+i]
	}
	return nil
}

// MakeBuffer создает буфер для передачи полностью упакованный со всеми КС
func (d *HeaderDevice) MakeBuffer() []byte {
	buffer := make([]byte, 19+len(d.Message)+4)
	buffer[0] = d.TypeDevice
	str := []byte(fmt.Sprintf("%02d%08d", d.TypeDevice, d.ID))
	for i := 0; i < 10; i++ {
		buffer[i] = str[i] - '0'
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
	buffer[pos+1] = uint8((sumB >> 8) & 0xff)
	buffer[pos] = uint8(sumB & 0xff)
	buffer[pos+3] = uint8((sumP >> 8) & 0xff)
	buffer[pos+2] = uint8(sumP & 0xff)

	return buffer
}

// MakeBuffer создает буфер для передачи полностью упакованный со всеми КС
func (s *HeaderServer) MakeBuffer() []byte {
	buffer := make([]byte, 13+len(s.Message)+4)
	buffer[0] = 0xa7
	buffer[1] = 0x8d
	putDate(time.Now(), buffer, 4)
	// s.Time = takeDate(buffer, 4)
	buffer[10] = s.Number
	buffer[11] = s.Code
	buffer[12] = uint8(len(s.Message) + 2)
	for i := 0; i < len(s.Message); i++ {
		buffer[i+13] = s.Message[i]
	}
	sumB, sumP := makeCRC(buffer, 13)
	pos := len(buffer) - 4
	buffer[pos+1] = uint8((sumB >> 8) & 0xff)
	buffer[pos] = uint8(sumB & 0xff)
	buffer[pos+3] = uint8((sumP >> 8) & 0xff)
	buffer[pos+2] = uint8(sumP & 0xff)
	return buffer
}

func makeCRC(buffer []byte, lenHeader int) (sumB uint, sumP uint) {
	sumP = 0xe752
	sumB = 0x2756
	for i := 0; i < lenHeader; i++ {
		sumP += uint(buffer[i])
		sumP = sumP & 0xffff
	}
	lenght := int(buffer[lenHeader-1] - 2)
	if lenght == 0 {
		sumB = 0
	}
	for i := 0; i < lenght; i++ {
		if lenHeader+i >= len(buffer) {
			logger.Error.Printf("error makeCRC len %d %v", lenght, buffer)
			return sumB, sumP
		}
		sumP += uint(buffer[lenHeader+i])
		sumP = sumP & 0xffff
		sumB += uint(buffer[lenHeader+i])
		sumB = sumB & 0xffff
	}
	sumB = (sumB) & 0xffff
	sumP += uint(sumB>>8) + uint(sumB&0xff)
	sumP = (sumP) & 0xffff //0xe752???
	return sumB, sumP
}
func checkCRC(buffer []byte, lenHeader int) error {
	sumB, sumP := makeCRC(buffer, lenHeader)
	l := int(buffer[lenHeader-1] - 2)
	if lenHeader+l >= len(buffer) || lenHeader+l+1 >= len(buffer) || lenHeader+l+2 >= len(buffer) || lenHeader+l+3 >= len(buffer) {
		return fmt.Errorf("ошибка CRC неверная длина буфера %d %d %v", lenHeader, len(buffer), buffer)
	}

	tb := ((uint(buffer[lenHeader+l+1]) & 0xff) << 8) | uint(buffer[lenHeader+l]&0xff)
	tp := ((uint(buffer[lenHeader+l+3]) & 0xff) << 8) | uint(buffer[lenHeader+l+2]&0xff)
	if tb != sumB || tp != sumP {
		return fmt.Errorf("ошибка CRC %d %d != %d %d", sumB, sumP, tb, tp)
	}
	return nil
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
	if t.Nanosecond() > 500000000 {
		t.Add(time.Second)
	}
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
