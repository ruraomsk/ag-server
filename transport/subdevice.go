package transport

import (
	"fmt"
	"time"

	"github.com/ruraomsk/ag-server/pudge"
)

//Set0x00Device записывает субсообщение для команды с номером в имени
func (s *SubMessage) Set0x00Device() {
	//num номер сообщения
	s.Type = 0
	s.Message = make([]uint8, 1)
	s.Message[0] = 0x00
}

//Set0x01Device записывает субсообщение для команды с номером в имени
func (s *SubMessage) Set0x01Device(num, min, sec, emas, elem int) {
	//num номер сообщения
	s.Type = 1
	s.Message = make([]uint8, 6)
	s.Message[0] = 0x01
	s.Message[1] = uint8(num)
	s.Message[2] = uint8(min)
	s.Message[3] = uint8(sec)
	s.Message[4] = uint8(emas)
	s.Message[5] = uint8(elem)
}

//Get0x01Device читает субсообщение
func (s *SubMessage) Get0x01Device() (num, min, sec, emas, elem int) {
	//num номер сообщения
	if s.Message[0] != 0x01 {
		return
	}
	num = int(s.Message[1])
	min = int(s.Message[2])
	sec = int(s.Message[3])
	emas = int(s.Message[4])
	elem = int(s.Message[5])
	return
}

//Set0x04Device записывает субсообщение для команды с номером в имени
func (s *SubMessage) Set0x04Device(hour, min, day, month int) {
	//num номер сообщения
	s.Type = 4
	s.Message = make([]uint8, 5)
	s.Message[0] = 0x04
	s.Message[1] = uint8(hour)
	s.Message[2] = uint8(min)
	s.Message[3] = uint8(day)
	s.Message[4] = uint8(month)
}

//Get0x04Device записывает субсообщение для команды с номером в имени
func (s *SubMessage) Get0x04Device() (hour, min, day, month int) {
	if s.Message[0] != 0x04 {
		return
	}
	hour = int(s.Message[1])
	min = int(s.Message[2])
	day = int(s.Message[3])
	month = int(s.Message[4])
	return
}

//Set0x07Device записывает субсообщение для команды с номером в имени
func (s *SubMessage) Set0x07Device(hour, min, day, month int) {
	//num номер сообщения
	s.Type = 0x07
	s.Message = make([]uint8, 5)
	s.Message[0] = 0x07
	s.Message[1] = uint8(hour)
	s.Message[2] = uint8(min)
	s.Message[3] = uint8(day)
	s.Message[4] = uint8(month)
}

//Get0x07Device записывает субсообщение для команды с номером в имени
func (s *SubMessage) Get0x07Device() (hour, min, day, month int) {
	if s.Message[0] != 0x07 {
		return
	}
	hour = int(s.Message[1])
	min = int(s.Message[2])
	day = int(s.Message[3])
	month = int(s.Message[4])
	return
}

//Get0x0FDevice изменяет состояние контроллера по команду
func (s *SubMessage) Get0x0FDevice(c *pudge.Controller) error {
	if s.Message[0] != 0x0f {
		return fmt.Errorf("неверный номер команды %x", s.Message[0])
	}
	c.TexRezim = int(s.Message[1] & 0x7f)
	c.Base = false
	if (s.Message[1] & 0x80) != 0 {
		c.Base = true
	}
	c.PK = int(s.Message[2])
	c.CK = int(s.Message[3])
	c.NK = int(s.Message[4])
	c.StatusCommandDU.Set(s.Message[5])
	c.DK.Set(s.Message, 6)
	return nil
}

//Set0x0FDevice изменяет состояние сообщения по состоянию контроллера
func (s *SubMessage) Set0x0FDevice(c *pudge.Controller) error {
	s.Type = 0x0f
	s.Message = make([]uint8, 22)
	s.Message[0] = 0x0f
	s.Message[1] = uint8(c.TexRezim)
	if c.Base {
		s.Message[1] |= 0x80
	}
	s.Message[2] = uint8(c.PK)
	s.Message[3] = uint8(c.CK)
	s.Message[4] = uint8(c.NK)
	s.Message[5] = c.StatusCommandDU.Make()
	c.DK.Make(s.Message, 6)
	return nil
}

//Set0x12Device изменяет состояние сообщения по состоянию контроллера
func (s *SubMessage) Set0x12Device(c *pudge.Controller) error {
	s.Type = 0x12
	s.Message = make([]uint8, 23)
	s.Message[0] = 0x12
	s.Message[1] = uint8(c.TexRezim)
	if c.Base {
		s.Message[1] |= 0x80
	}
	s.Message[2] = uint8(c.PK)
	s.Message[3] = uint8(c.CK)
	s.Message[4] = uint8(c.NK)
	s.Message[5] = c.StatusCommandDU.Make()
	c.DK.Make(s.Message, 6)
	s.Message[22] = uint8(c.TMax)
	return nil
}

//Get0x12Device изменяет состояние контроллера по команду
func (s *SubMessage) Get0x12Device(c *pudge.Controller) error {
	if s.Message[0] != 0x12 {
		return fmt.Errorf("неверный номер команды %x", s.Message[0])
	}
	c.TexRezim = int(s.Message[1] & 0x12)
	c.Base = false
	if (s.Message[1] & 0x80) != 0 {
		c.Base = true
	}
	c.PK = int(s.Message[2])
	c.CK = int(s.Message[3])
	c.NK = int(s.Message[4])
	c.StatusCommandDU.Set(s.Message[5])
	c.DK.Set(s.Message, 6)
	c.TMax = int(s.Message[22])
	return nil
}

//Get0x10Device изменяет состояние контроллера по команду
func (s *SubMessage) Get0x10Device(c *pudge.Controller) error {
	if s.Message[0] != 0x10 {
		return fmt.Errorf("неверный номер команды %x", s.Message[0])
	}
	c.Model.VPCPD = int(s.Message[1])<<8 | int(s.Message[2])
	c.Model.VPBS = int(s.Message[3])<<8 | int(s.Message[4])
	c.Model.C12 = (s.Message[5] & 1) != 0
	c.Model.STP = (s.Message[5] & 2) != 0
	c.Model.DKA = (s.Message[5] & 4) != 0
	c.Model.DTA = (s.Message[5] & 8) != 0
	return nil
}

//Set0x10Device изменяет состояние контроллера по команду
func (s *SubMessage) Set0x10Device(c *pudge.Controller) {
	s.Type = 0x10
	s.Message = make([]uint8, 6)
	s.Message[0] = 0x10
	s.Message[1] = uint8((c.Model.VPCPD >> 8) & 0xff)
	s.Message[2] = uint8(c.Model.VPCPD & 0xff)
	s.Message[3] = uint8((c.Model.VPBS >> 8) & 0xff)
	s.Message[4] = uint8(c.Model.VPBS & 0xff)
	s.Message[5] = 0
	if c.Model.C12 {
		s.Message[5] |= 1
	}
	if c.Model.STP {
		s.Message[5] |= 2
	}
	if c.Model.DKA {
		s.Message[5] |= 4
	}
	if c.Model.DTA {
		s.Message[5] |= 8
	}
}

//Get0x11Device изменяет состояние контроллера по команду
func (s *SubMessage) Get0x11Device(c *pudge.Controller) error {
	if s.Message[0] != 0x11 {
		return fmt.Errorf("неверный номер команды %x", s.Message[0])
	}
	c.Error.V220DK1 = (s.Message[1] & 1) != 0
	c.Error.V220DK2 = (s.Message[1] & 32) != 0
	c.Error.RTC = (s.Message[1] & 2) != 0
	c.Error.TVP1 = (s.Message[1] & 4) != 0
	c.Error.TVP2 = (s.Message[1] & 8) != 0
	c.Error.FRAM = (s.Message[1] & 16) != 0
	c.GPS.Ok = s.Message[2] == 0
	c.GPS.E01 = s.Message[2] == 1
	c.GPS.E02 = s.Message[2] == 2
	c.GPS.E03 = s.Message[2] == 3
	c.GPS.E04 = s.Message[2] == 4
	c.GPS.Seek = s.Message[2] == 0x0A
	c.Input.V1 = (s.Message[3] & 1) != 0
	c.Input.V2 = (s.Message[3] & 2) != 0
	c.Input.V3 = (s.Message[3] & 4) != 0
	c.Input.V4 = (s.Message[3] & 8) != 0
	c.Input.V5 = (s.Message[3] & 16) != 0
	c.Input.V6 = (s.Message[3] & 32) != 0
	c.Input.V7 = (s.Message[3] & 64) != 0
	c.Input.V8 = (s.Message[3] & 128) != 0
	return nil
}

//Set0x11Device изменяет состояние контроллера по команду
func (s *SubMessage) Set0x11Device(c *pudge.Controller) {
	s.Type = 0x10
	s.Message = make([]uint8, 4)
	s.Message[0] = 0x11
	s.Message[1] = 0
	if c.Error.V220DK1 {
		s.Message[1] |= 1
	}
	if c.Error.V220DK2 {
		s.Message[1] |= 32
	}
	if c.Error.RTC {
		s.Message[1] |= 2
	}
	if c.Error.TVP1 {
		s.Message[1] |= 4
	}
	if c.Error.TVP2 {
		s.Message[1] |= 8
	}
	if c.Error.FRAM {
		s.Message[1] |= 16
	}
	if c.GPS.Ok {
		s.Message[2] = 0
	}
	if c.GPS.E01 {
		s.Message[2] = 1
	}
	if c.GPS.E02 {
		s.Message[2] = 2
	}
	if c.GPS.E03 {
		s.Message[2] = 3
	}
	if c.GPS.E04 {
		s.Message[2] = 4
	}
	if c.GPS.Seek {
		s.Message[2] = 0x0A
	}
	s.Message[3] = 0

	if c.Input.V1 {
		s.Message[3] |= 1
	}
	if c.Input.V2 {
		s.Message[3] |= 2
	}
	if c.Input.V3 {
		s.Message[3] |= 4
	}
	if c.Input.V4 {
		s.Message[3] |= 8
	}
	if c.Input.V5 {
		s.Message[3] |= 16
	}
	if c.Input.V6 {
		s.Message[3] |= 32
	}
	if c.Input.V7 {
		s.Message[3] |= 64
	}
	if c.Input.V8 {
		s.Message[3] |= 128
	}
}

//Set0x09Device записывает субсообщение для команды с номером в имени
func (s *SubMessage) Set0x09Device(st *pudge.Statistic) {
	//num номер сообщения
	s.Type = 0x09
	s.Message = make([]uint8, 6)
	s.Message[0] = 0x09
	s.Message[1] = uint8(st.Type)
	s.Message[2] = uint8(st.Period)
	s.Message[3] = uint8(st.TLen)
	s.Message[4] = uint8(st.Hour)
	s.Message[5] = uint8(st.Min)
}

//Get0x09Device записывает субсообщение для команды с номером в имени
func (s *SubMessage) Get0x09Device() (st pudge.Statistic, err error) {
	if s.Message[0] != 0x09 {
		return st, fmt.Errorf("неверный номер команды %x", s.Message[0])
	}
	st.Datas = make([]pudge.DataStat, 0)
	st.Type = int(s.Message[1])
	st.Period = int(s.Message[2])
	st.TLen = int(s.Message[3])
	st.Hour = int(s.Message[4])
	st.Min = int(s.Message[5])
	return
}

//Set0x0ADevice сообщение статистики
func (s *SubMessage) Set0x0ADevice(st *pudge.Statistic) error {
	if len(st.Datas) > 16 {
		return fmt.Errorf("слишком много элементов статистики %d", len(st.Datas))
	}
	s.Type = 0x0a
	s.Message = make([]uint8, 3+(len(st.Datas)*3))
	s.Message[0] = 0x0a
	s.Message[1] = uint8(st.Period)
	s.Message[2] = uint8(len(st.Datas) * 3)
	pos := 3
	for _, el := range st.Datas {
		s.Message[pos] = uint8(el.Chanel&0x3f) | uint8(el.Status<<6)
		pos++
		s.Message[pos] = uint8(el.Intensiv & 0xff)
		pos++
		s.Message[pos] = uint8((el.Intensiv >> 8) & 0xff)
		pos++
	}
	return nil
}

//Get0x0ADevice сообщение статистики
func (s *SubMessage) Get0x0ADevice(st *pudge.Statistic) error {
	if s.Message[0] != 0x0a {
		return fmt.Errorf("неверный номер команды %x", s.Message[0])
	}
	if st.Period != int(s.Message[1]) {
		return fmt.Errorf("неверный номер периода %d", s.Message[1])
	}
	st.Datas = make([]pudge.DataStat, 0)
	count := int(s.Message[2]) / 3
	pos := 3
	var el pudge.DataStat
	for count > 0 {
		el.Chanel = int(s.Message[pos] & 0x3f)
		el.Status = int(s.Message[pos] >> 6)
		pos++
		el.Intensiv = int(s.Message[pos])
		pos++
		el.Intensiv |= int(s.Message[pos]) << 8
		pos++
		st.Datas = append(st.Datas, el)
		count--
	}
	return nil
}

//Set0x0BDevice сообщение статистики
func (s *SubMessage) Set0x0BDevice(lg *pudge.LogLine) {
	s.Type = 0x0b
	s.Message = make([]uint8, 9)
	s.Message[0] = 0x0b
	putDateDevice(lg.Time, s.Message, 1)
	lg.Time = takeDateDevice(s.Message, 1)
	s.Message[7] = uint8(lg.Record)
	s.Message[8] = uint8(lg.Info)

}

//Get0x0BDevice сообщение статистики
func (s *SubMessage) Get0x0BDevice(lg *pudge.LogLine) error {
	if s.Message[0] != 0x0b {
		return fmt.Errorf("неверный номер команды %x", s.Message[0])
	}
	lg.Time = takeDateDevice(s.Message, 1)
	lg.Record = int(s.Message[7])
	lg.Info = int(s.Message[8])

	return nil
}

//Set0x13Device сообщение массива привязки
func (s *SubMessage) Set0x13Device(ar *pudge.ArrayPriv) {
	s.Type = 0x13
	s.Message = make([]uint8, len(ar.Array)+3)
	s.Message[0] = 0x13
	s.Message[1] = uint8(ar.Number)
	s.Message[2] = uint8(len(ar.Array))
	pos := 3
	for _, al := range ar.Array {
		s.Message[pos] = uint8(al)
		pos++
	}
}

//Get0x13Device сообщение массива привязки
func (s *SubMessage) Get0x13Device(ar *pudge.ArrayPriv) error {
	if s.Message[0] != 0x13 {
		return fmt.Errorf("неверный номер команды %x", s.Message[0])
	}
	ar.Number = int(s.Message[1])
	ar.Array = make([]int, 0)
	count := int(s.Message[2])
	pos := 3
	for count > 0 {
		ar.Array = append(ar.Array, int(s.Message[pos]))
		pos++
		count--
	}
	return nil
}

//Set0x1CDevice записывает субсообщение для команды с номером в имени
func (s *SubMessage) Set0x1CDevice(num int) {
	//num номер сообщения
	s.Type = 0x1c
	s.Message = make([]uint8, 6)
	s.Message[0] = 0x1C
	s.Message[1] = uint8(num)
}

//Get0x1CDevice читает субсообщение
func (s *SubMessage) Get0x1CDevice() (num int) {
	//num номер сообщения
	if s.Message[0] != 0x1C {
		return -1
	}
	num = int(s.Message[1])
	return
}

//Get0x1DDevice изменяет состояние контроллера по команду
func (s *SubMessage) Get0x1DDevice(c *pudge.Controller) error {
	if s.Message[0] != 0x1D {
		return fmt.Errorf("неверный номер команды %x", s.Message[0])
	}
	c.Error.V220DK1 = (s.Message[1] & 1) != 0
	c.Error.V220DK2 = (s.Message[1] & 32) != 0
	c.Error.RTC = (s.Message[1] & 2) != 0
	c.Error.TVP1 = (s.Message[1] & 4) != 0
	c.Error.TVP2 = (s.Message[1] & 8) != 0
	c.Error.FRAM = (s.Message[1] & 16) != 0
	c.GPS.Ok = s.Message[2] == 0
	c.GPS.E01 = s.Message[2] == 1
	c.GPS.E02 = s.Message[2] == 2
	c.GPS.E03 = s.Message[2] == 3
	c.GPS.E04 = s.Message[2] == 4
	c.GPS.Seek = s.Message[2] == 0x0A
	// c.Input.V1 = (s.Message[3] & 1) != 0
	// c.Input.V2 = (s.Message[3] & 2) != 0
	// c.Input.V3 = (s.Message[3] & 4) != 0
	// c.Input.V4 = (s.Message[3] & 8) != 0
	// c.Input.V5 = (s.Message[3] & 16) != 0
	// c.Input.V6 = (s.Message[3] & 32) != 0
	// c.Input.V7 = (s.Message[3] & 64) != 0
	// c.Input.V8 = (s.Message[3] & 128) != 0
	return nil
}

//Set0x1DDevice изменяет состояние контроллера по команду
func (s *SubMessage) Set0x1DDevice(c *pudge.Controller) {
	s.Type = 0x1D
	s.Message = make([]uint8, 13)
	s.Message[0] = 0x1D
	s.Message[1] = 0
	if c.Error.V220DK1 {
		s.Message[1] |= 1
	}
	if c.Error.V220DK2 {
		s.Message[1] |= 32
	}
	if c.Error.RTC {
		s.Message[1] |= 2
	}
	if c.Error.TVP1 {
		s.Message[1] |= 4
	}
	if c.Error.TVP2 {
		s.Message[1] |= 8
	}
	if c.Error.FRAM {
		s.Message[1] |= 16
	}
	if c.GPS.Ok {
		s.Message[2] = 0
	}
	if c.GPS.E01 {
		s.Message[2] = 1
	}
	if c.GPS.E02 {
		s.Message[2] = 2
	}
	if c.GPS.E03 {
		s.Message[2] = 3
	}
	if c.GPS.E04 {
		s.Message[2] = 4
	}
	if c.GPS.Seek {
		s.Message[2] = 0x0A
	}
	s.Message[3] = 0

	// if c.Input.V1 {
	// 	s.Message[3] |= 1
	// }
	// if c.Input.V2 {
	// 	s.Message[3] |= 2
	// }
	// if c.Input.V3 {
	// 	s.Message[3] |= 4
	// }
	// if c.Input.V4 {
	// 	s.Message[3] |= 8
	// }
	// if c.Input.V5 {
	// 	s.Message[3] |= 16
	// }
	// if c.Input.V6 {
	// 	s.Message[3] |= 32
	// }
	// if c.Input.V7 {
	// 	s.Message[3] |= 64
	// }
	// if c.Input.V8 {
	// 	s.Message[3] |= 128
	// }
}

func takeDateDevice(buffer []byte, pos int) time.Time {
	year := int(buffer[pos+5]) + 2000
	month := time.Month(int(buffer[pos+4]))
	day := int(buffer[pos] + 3)
	hour := int(buffer[pos])
	minut := int(buffer[pos+1])
	sec := int(buffer[pos+2])
	location, _ := time.LoadLocation("Local")
	return time.Date(year, month, day, hour, minut, sec, 0, location)

}
func putDateDevice(t time.Time, buffer []byte, pos int) {
	year, month, day := t.Date()
	hour := t.Hour()
	min := t.Minute()
	sec := t.Second()
	buffer[pos] = uint8(hour)
	buffer[pos+1] = uint8(min)
	buffer[pos+2] = uint8(sec)
	buffer[pos+3] = uint8(day)
	buffer[pos+4] = uint8(month)
	buffer[pos+5] = uint8(year % 100)
}
