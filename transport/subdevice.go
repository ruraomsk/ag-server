package transport

import (
	"fmt"
	"time"

	"github.com/ruraomsk/ag-server/logger"
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
	c.TechMode = int(s.Message[1] & 0x7f)
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
	s.Message[1] = uint8(c.TechMode)
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
	s.Message[1] = uint8(c.TechMode)
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
	// logger.Debug.Printf("id %d %v", c.ID, s.Message)
	c.TechMode = int(s.Message[1] & 0x7f)
	c.Base = false
	if (s.Message[1] & 0x80) != 0 {
		c.Base = true
	}
	c.PK = int(s.Message[2])
	c.CK = int(s.Message[3])
	c.NK = int(s.Message[4])
	c.StatusCommandDU.Set(s.Message[5])
	c.DK.Set(s.Message, 6)
	c.TMax = int64(s.Message[22])
	return nil
}

//Get0x10Device изменяет состояние контроллера по команду
func (s *SubMessage) Get0x10Device(c *pudge.Controller) error {
	if s.Message[0] != 0x10 {
		return fmt.Errorf("неверный номер команды %x", s.Message[0])
	}
	c.Model.VPCPDL = int(s.Message[1])
	c.Model.VPCPDR = int(s.Message[2])
	c.Model.VPBSL = int(s.Message[3])
	c.Model.VPBSR = int(s.Message[4])
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
	s.Message[1] = uint8(c.Model.VPCPDL & 0xff)
	s.Message[2] = uint8(c.Model.VPCPDR & 0xff)
	s.Message[3] = uint8(c.Model.VPBSL & 0xff)
	s.Message[4] = uint8(c.Model.VPBSR & 0xff)
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
	if s.Message[2] > 0x0A {
		s.Message[2] = 1
	}
	c.GPS.Ok = s.Message[2] == 0
	c.GPS.E01 = s.Message[2] == 1
	c.GPS.E02 = s.Message[2] == 2
	c.GPS.E03 = s.Message[2] == 3
	c.GPS.E04 = s.Message[2] == 4
	c.GPS.Seek = s.Message[2] == 0x0A
	//fmt.Println(s.Message[2])
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
	lg.Time = takeDateDevice(s.Message)
	s.Message[7] = 1
	s.Message[8] = 2

}

//Get0x0BDevice сообщение статистики
func (s *SubMessage) Get0x0BDevice(lg *pudge.LogLine) error {
	if s.Message[0] != 0x0b {
		return fmt.Errorf("неверный номер команды %x", s.Message[0])
	}
	lg.Time = takeDateDevice(s.Message)
	lg.Record, lg.Info = getTypeRecod(int(s.Message[7]), s.Message)
	// logger.Debug.Printf("Длина сообщения %d", len(s.Message))
	return nil
}
func getTypeRecod(t int, buffer []uint8) (tp string, info string) {
	s0 := ""
	s1 := ""
	s2 := ""
	s3 := ""
	if t == 2 {
		logger.Info.Printf("есть ДК2")
	}
	switch t {
	case 1:
		logger.Debug.Printf("%d-%d %d-%d %d-%d %d-%d ", buffer[8]>>4, buffer[8]&15, buffer[9]>>4, buffer[9]&15, buffer[10]>>4, buffer[10]&15, buffer[11]>>4, buffer[11]&15)
		// switch buffer[8] & 15 {
		switch buffer[8] >> 4 {
		case 1:
			s0 = "РУ"
		case 2:
			s0 = "РП"
		case 3:
			s0 = "ЗУ"
		case 4:
			s0 = "ДУ"
		case 5:
			s0 = "ЛУ"
		case 6:
			s0 = "ЛРП"
		case 7:
			s0 = "МГР"
		case 8:
			s0 = "КУ"
		case 9:
			s0 = "РКУ"
		default:
			s0 = fmt.Sprintf("%d", buffer[8]>>4)
		}
		switch buffer[8] & 15 {
		case 9:
			s1 = "ПР"
		case 10:
			s1 = "ЖМ"
		case 11:
			s1 = "ОС"
		case 12:
			s1 = "КК"
		default:
			s1 = fmt.Sprintf("%d", buffer[8]&15)
		}
		switch buffer[11] & 15 {
		case 0:
			s2 = "УСТОЙЧИВОЕ"
		case 1:
			s2 = "ПЕРЕХОД"
		case 2:
			s2 = "ОБРЫВ ЛС"
		case 3:
			s2 = "НГ ПО ПАРИТЕТУ"
		case 4:
			s2 = "НЕСУЩЕСТВУЮЩИЙ КОД"
		case 5:
			s2 = "КОНФЛИКТ"
		case 6:
			s2 = "ПЕРЕГОРАНИЕ ЛАМП"
		case 7:
			s2 = "НЕВКЛЮЧЕНИЕ В КООРДИНАЦИЮ"
		case 8:
			s2 = "НЕПОДЧИНЕНИЕ"
		case 9:
			s2 = "ДЛИННЫЙ ПРОМТАКТ"
		case 10:
			s2 = "НЕСУЩЕСТВУЮЩИЙ КОД"
		default:
			s2 = fmt.Sprintf("%d", buffer[10]&15)
		}
		switch buffer[10] & 15 {
		case 1:
			s3 = "ВПУ"
		case 2:
			s3 = "СКА"
		case 3:
			s3 = "ИП КЗЦ"
		case 4:
			s3 = "КЗЦ"
		case 5:
			s3 = "ИП КЗЦ1"
		case 6:
			s3 = "ПОУ(ДПОУ)"
		case 7:
			s3 = "ЭВМ-НУ"
		case 8:
			s3 = "КЗЦ1 (ПКУ"
		case 9:
			s3 = "ЭВМ ВУ"
		default:
			s3 = fmt.Sprintf("%d", buffer[11]&15)
		}

		return "ДК1", fmt.Sprintf("Режим %s фаза %s состояние %s устройство %s ", s0, s1, s2, s3)
	case 3:
		if buffer[8]&128 != 0 {
			s0 = "БАЗОВАЯ ПРИВЯЗКА "
			buffer[8] = buffer[8] & 127
		}
		switch buffer[8] & 15 {
		case 1:
			s0 += "ПК-СК"
		case 2:
			s0 += "ПК-НК"
		case 3:
			s0 += "ПК-СК ДУ"
		case 4:
			s0 += "ПК-РК ДУ"
		case 5:
			s0 += "ПК ДУ"
		case 6:
			s0 += "РП"
		case 7:
			s0 += "Коррекция привязки ИП"
		case 8:
			s0 += "Коррекция привязки сервер"
		case 9:
			s0 += "ПК ГК"
		case 10:
			s0 += "ПК ХТ"
		case 11:
			s0 += "ПК КАРТОГРАММА"
		case 12:
			s0 += "ПЗУ"
		default:
			s0 = fmt.Sprintf("%d", buffer[8])

		}

		return "Технология", fmt.Sprintf("Режим %s ПУ %d СК %d НК %d", s0, buffer[9], buffer[10], buffer[11])
	case 4:
		switch buffer[8] {
		case 0:
			return "Начало работы", "ОК"
		case 1:
			return "Начало работы", "Ошибка контрольной суммы привязки"
		case 2:
			return "Начало работы", "Базовая привязка"
		}
		return "Начало работы", fmt.Sprintf("Код %d", buffer[8])
	case 5:
		if buffer[8]&1 != 0 {
			s1 += "Пропадание 220В "
		}
		if buffer[8]&2 != 0 {
			s1 += "Неисправность часов реального времени "
		}
		if buffer[8]&4 != 0 {
			s1 += "Неисправность часов ТВП1 "
		}
		if buffer[8]&8 != 0 {
			s1 += "Неисправность часов ТВП2 "
		}
		if len(s1) == 0 {
			s2 = "Неисправностей нет "
		}
		if buffer[9]&1 != 0 {
			s2 += "1 "
		}
		if buffer[9]&2 != 0 {
			s2 += "2 "
		}
		if buffer[9]&4 != 0 {
			s2 += "3 "
		}
		if buffer[9]&8 != 0 {
			s2 += "4 "
		}
		if buffer[9]&16 != 0 {
			s2 += "5 "
		}
		if buffer[9]&32 != 0 {
			s2 += "6 "
		}
		if buffer[9]&64 != 0 {
			s2 += "7 "
		}
		if buffer[9]&128 != 0 {
			s2 += "8 "
		}
		if len(s2) == 0 {
			s2 = "Входы исправны "
		}
		s3 = fmt.Sprintf("%d", buffer[10])
		if buffer[10] == 0 {
			s3 = "Исправно"
		}
		if buffer[10] == 1 {
			s3 = "Нет связи с приемником"
		}
		if buffer[10] == 2 {
			s3 = "Ошибка CRC"
		}
		if buffer[10] == 3 {
			s3 = "Нет валидного времени"
		}
		if buffer[10] == 4 {
			s3 = "Мало спутников"
		}
		if buffer[10] == 0x0A {
			s3 = "Поиск спутников"
		}
		return "ПСПД", fmt.Sprintf("%s %s GPS:%s", s1, s2, s3)
	case 6:
		switch buffer[8] {
		case 0:
			s0 = "НОРМАЛЬНЫЙ ОБМЕН"
		case 1:
			s0 = "НЕИСПРАВЕН МОДЕМ"
		case 2:
			s0 = "НЕТ СЕТИ"
		case 3:
			s0 = "НЕТ ИНТЕРНЕТА"
		case 4:
			s0 = "НЕТ СЕРВЕРА"
		case 5:
			s0 = "НЕТ SIM-КАРТЫ"
		case 6:
			s0 = "НЕ ПРОПИСАН"
		case 7:
			s0 = "НЕТ ОТВЕТА СЕРВЕРА"
		case 8:
			s0 = "РАЗРЫВ СВЯЗИ"
		case 10:
			s0 = "СОЕДИНЕНИЕ"
		case 11:
			s0 = "ОШ КС СООБЩЕНИЯ СЕРВЕРА"
		case 12:
			s0 = "ПСПД НЕТ ОТВЕТА"
		case 30:
			s0 = "ПСПД НЕТ ПБС"
		case 31:
			s0 = "ПСПД ОШ КС ПБС"
		default:
			s0 = fmt.Sprintf("%d", buffer[8])
		}
		return "Сервер", fmt.Sprintf("Состояние  %s", s0)
	case 7:
		return "GPS", fmt.Sprintf("Состояние  %d", buffer[8])
	}
	return fmt.Sprintf("Тип %d", t), fmt.Sprintf("Состояние  %d", buffer[8])
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
	c.Status.StatusV220 = int(s.Message[1])
	c.Status.StatusGPS = int(s.Message[2])
	c.Status.StatusServer = int(s.Message[3])
	c.Status.StatusPSPD = int(s.Message[4])
	c.Status.ErrorLastConn = int(s.Message[5])
	c.Status.Ethernet = s.Message[6]&8 != 0
	c.Status.TObmen = int(s.Message[7])
	c.Status.LevelGSMNow = int(s.Message[8])
	c.Status.LevelGSMLast = int(s.Message[9])
	c.Status.Motiv = int(s.Message[10])
	//logger.Debug.Printf("1D %v",s.Message)
	return nil
}

//Set0x1DDevice изменяет состояние контроллера по команду
func (s *SubMessage) Set0x1DDevice(c *pudge.Controller) {
	s.Type = 0x1D
	s.Message = make([]uint8, 13)
	s.Message[0] = 0x1D
	s.Message[1] = uint8(c.Status.StatusV220)
	s.Message[2] = uint8(c.Status.StatusGPS)
	s.Message[3] = uint8(c.Status.StatusServer)
	s.Message[4] = uint8(c.Status.StatusPSPD)
	s.Message[5] = uint8(c.Status.ErrorLastConn)
	s.Message[6] = 0
	if c.Status.Ethernet {
		s.Message[6] = 4
	}
	s.Message[7] = uint8(c.Status.TObmen)
	s.Message[8] = uint8(c.Status.LevelGSMNow)
	s.Message[9] = uint8(c.Status.LevelGSMLast)
	s.Message[10] = uint8(c.Status.Motiv)
}

//Get0x1BDevice изменяет состояние контроллера по команду
func (s *SubMessage) Get0x1BDevice(c *pudge.Controller) error {
	if s.Message[0] != 0x1B {
		return fmt.Errorf("неверный номер команды %x", s.Message[0])
	}
	c.Status.StatusV220 = int(s.Message[1])
	c.Status.StatusGPS = int(s.Message[2])
	c.Status.StatusServer = int(s.Message[3])
	c.Status.StatusPSPD = int(s.Message[4])
	c.Status.ErrorLastConn = int(s.Message[5])
	c.Status.Ethernet = s.Message[6]&8 != 0
	c.Status.TObmen = int(s.Message[7])
	c.Status.LevelGSMNow = int(s.Message[8])
	//logger.Debug.Printf("1B %v",s.Message)
	return nil
}

//Set0x1BDevice изменяет состояние контроллера по команду
func (s *SubMessage) Set0x1BDevice(c *pudge.Controller) {
	s.Type = 0x1D
	s.Message = make([]uint8, 9)
	s.Message[0] = 0x1D
	s.Message[1] = uint8(c.Status.StatusV220)
	s.Message[2] = uint8(c.Status.StatusGPS)
	s.Message[3] = uint8(c.Status.StatusServer)
	s.Message[4] = uint8(c.Status.StatusPSPD)
	s.Message[5] = uint8(c.Status.ErrorLastConn)
	s.Message[6] = 0
	if c.Status.Ethernet {
		s.Message[6] = 4
	}
	s.Message[7] = uint8(c.Status.TObmen)
	s.Message[8] = uint8(c.Status.LevelGSMNow)
}

func takeDateDevice(buffer []byte) time.Time {
	year := int(buffer[6]) + 2000
	month := time.Month(int(buffer[5]))
	day := int(buffer[4])
	hour := int(buffer[1])
	minut := int(buffer[2])
	sec := int(buffer[3])
	location, _ := time.LoadLocation("Local")
	return time.Date(year, month, day, hour, minut, sec, 0, location)

}
func putDateDevice(t time.Time, buffer []byte, pos int) {
	year, month, day := t.Date()
	hour := t.Hour()
	min := t.Minute()
	sec := t.Second()
	buffer[pos+1] = uint8(hour)
	buffer[pos+2] = uint8(min)
	buffer[pos+3] = uint8(sec)
	buffer[pos+4] = uint8(day)
	buffer[pos+5] = uint8(month)
	buffer[pos+6] = uint8(year % 100)
}
