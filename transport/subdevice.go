package transport

import "rura/ag-server/pudge"

import "fmt"

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

//Set0x0FDevice изменяет состояние контроллера по команду
func (s *SubMessage) Set0x0FDevice(c *pudge.Controller) error {
	if s.Type != 0x0f && s.Type != 0x12 {
		return fmt.Errorf("неверный номер команды %x", s.Type)
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
	c.DK1.Set(s.Message, 6)
	c.DK2.Set(s.Message, 14)
	if s.Type == 0x12 {
		c.TMax = int(s.Message[22])
	}
	return nil
}

//Make0x0FDevice изменяет состояние сообщения по состоянию контроллера
func (s *SubMessage) Make0x0FDevice(c *pudge.Controller) error {
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
	c.DK1.Make(s.Message, 6)
	c.DK2.Make(s.Message, 14)
	return nil
}

//Make0x12Device изменяет состояние сообщения по состоянию контроллера
func (s *SubMessage) Make0x12Device(c *pudge.Controller) error {
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
	c.DK1.Make(s.Message, 6)
	c.DK2.Make(s.Message, 14)
	s.Message[22] = uint8(c.TMax)
	return nil
}
