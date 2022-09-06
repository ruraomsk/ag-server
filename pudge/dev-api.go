package pudge

import (
	"fmt"

	"github.com/ruraomsk/ag-server/logger"
)

//Make создает байт для передачи статуса команд ДУ
func (s *StatusCommandDU) Make() (command uint8) {
	command = 0
	if s.IsPK {
		command |= 1
	}
	if s.IsCK {
		command |= 2
	}
	if s.IsNK {
		command |= 4
	}
	if s.IsDUDK1 {
		command |= 8
	}
	if s.IsDUDK2 {
		command |= 16
	}
	if s.IsReqSFDK1 {
		command |= 32
	}
	if s.IsReqSFDK2 {
		command |= 64
	}
	return
}

//Set переносит из байта статуса команд ДУ в поле состояния
func (s *StatusCommandDU) Set(command uint8) {
	s.IsPK = false
	s.IsCK = false
	s.IsNK = false
	s.IsDUDK1 = false
	s.IsDUDK2 = false
	s.IsReqSFDK1 = false
	s.IsReqSFDK2 = false
	if command&1 != 0 {
		s.IsPK = true
	}
	if command&2 != 0 {
		s.IsCK = true
	}
	if command&4 != 0 {
		s.IsNK = true
	}
	if command&8 != 0 {
		s.IsDUDK1 = true
	}
	if command&16 != 0 {
		s.IsDUDK2 = true
	}
	if command&32 != 0 {
		s.IsReqSFDK1 = true
	}
	if command&64 != 0 {
		s.IsReqSFDK2 = true
	}
}

//Set устанавлиявает поля
func (d *DK) Set(buffer []byte, pos int) {
	d.RDK = int(buffer[pos] & 0xf)
	d.FDK = int((buffer[pos] & 0xf0) >> 4)
	pos++
	d.DDK = int(buffer[pos] & 0xf)
	d.EDK = int((buffer[pos] & 0xf0) >> 4)
	pos++
	if (buffer[pos] & 0x80) != 0 {
		d.PDK = true
	} else {
		d.PDK = false
	}
	d.EEDK = int(buffer[pos] & 0xf)
	if d.EEDK == 0 {
		d.EEDK = d.EDK
	}
	pos++
	if (buffer[pos] & 0x80) != 0 {
		d.ODK = true
	} else {
		d.ODK = false
	}
	d.LDK = int(buffer[pos] & 0xf)
	if d.LDK == 0 && d.EEDK == 6 {
		d.LDK = d.FDK
	}
	pos++
	d.FTUDK = int(buffer[pos])
	pos++
	d.TDK = int(buffer[pos])
	pos++
	d.FTSDK = int(buffer[pos])
	pos++
	d.TTCDK = int(buffer[pos])
}

//Make устанавлиявает поля
func (d *DK) Make(buffer []byte, pos int) {
	buffer[pos] |= uint8(d.RDK & 0xf)
	buffer[pos] |= uint8((d.FDK & 0xf) << 4)
	pos++
	buffer[pos] |= uint8(d.DDK & 0xf)
	buffer[pos] |= uint8((d.EDK & 0xf) << 4)
	pos++
	if d.PDK {
		buffer[pos] |= 0x80
	}
	buffer[pos] |= uint8(d.EEDK & 0xf)
	pos++
	if d.ODK {
		buffer[pos] |= 0x80
	}
	buffer[pos] |= uint8(d.LDK & 0xf)
	pos++
	buffer[pos] = uint8(d.FTUDK)
	pos++
	buffer[pos] = uint8(d.TDK)
	pos++
	buffer[pos] = uint8(d.FTSDK)
	pos++
	buffer[pos] = uint8(d.TTCDK)
}

// var lrezim int
// var lfaza int
// var lerr int
// var ldev int
// var llamp int
// var ldoor int
func (cc *Controller) getSource() string {
	//	1 - ДК
	//	2 - ВПУ
	//	3 - инженерный пульт (ИП УСДК)
	//	4 - УСДК/ДКА
	//	5 - инженерная панель (ИП ДКА)
	//	6 - система (ЭВМ)
	//	7 - система (ЭВМ)
	// 	8 - система (ЭВМ)
	// 	9 - система (ЭВМ)

	switch cc.DK.DDK {
	case 1:
		return "ДК "
	case 2:
		return "ВПУ "
	case 3:
		if cc.Model.C12 {
			return "ИП C12 "
		}
		if cc.Model.DKA {
			return "ИП ДКА "
		}
		if cc.Model.DTA {
			return "ИП ДТА "
		}
		return "ИП УСДК "
	case 4:
		if cc.Model.C12 {
			return "C12 "
		}
		if cc.Model.DKA {
			return "ДКА "
		}
		if cc.Model.DTA {
			return "ДТА "
		}
		return "УСДК "
	case 5:
		return "ИП ДКА "
	}
	if cc.DK.DDK >= 6 && cc.DK.DDK <= 9 {
		return "ЭВМ "
	}
	return fmt.Sprintf("Источник %d", cc.DK.DDK)
}

func (cc *Controller) CalcStatus() int {
	rezim := cc.DK.RDK
	faza := cc.DK.FDK
	err := cc.DK.EDK
	dev := cc.codeDevice()
	lamp := cc.lamps()
	door := cc.doors()
	// if lrezim != rezim || lfaza != faza || lerr != err || ldev != dev || llamp != lamp || ldoor != door {
	// 	logger.Info.Printf("rezim=%d faza=%d err=%d dev=%d lamp=%d door=%d", rezim, faza, err, dev, lamp, door)
	// 	lrezim = rezim
	// 	lfaza = faza
	// 	ldev = dev
	// 	llamp = lamp
	// 	ldoor = door

	//	dev := cc.codeDevice()
	// 1 - ВПУ
	// 2 - ДК
	// 3 - ИП УСДК
	// 4 - УСДК
	// 5 - ИП ДК
	//
	// 7 - ДУ ЭВМ

	//rezim := cc.DK.RDK
	// 1 - РУ
	// 2 - РП
	// 3 - ЗУ
	// 4 - ДУ
	// 6 - ЛР
	// 7 - ЛРП
	// 8 - МГР
	// 9 - КУ
	// 10 - РКУ

	// }
	if !cc.IsConnected() {
		if err == 11 && dev == 3 {
			//Авария 220 16
			return 16
		}
		if err == 11 && dev == 5 { //Спросить как узнать ошибки контроллера УСДК
			//Выключен УСДК/ДК 17
			return 17
		}
		return 18
	}
	if cc.Base {
		return 22
	}
	if (rezim == 8 || rezim == 9) && (err == 0 || err == 1) && (faza >= 1 && faza <= 9) && lamp == 0 && door == 0 {
		//Координированное управление 1
		return 1
	}
	if rezim == 4 && (err == 0 || err == 1) && (faza >= 1 && faza <= 9) {
		//Диспетчерское управление 2
		return 2
	}
	if (rezim == 1 || rezim == 2) && (err == 0 || err == 1) && (faza >= 1 && faza <= 9) {
		//Ручное управление 3
		return 3
	}
	if (rezim == 1) && (err == 0 || err == 1) && (faza == 12) {
		//Ручное управление 3
		return 3
	}
	if rezim == 3 && (err == 0 || err == 1) && (faza >= 1 && faza <= 9) && lamp == 0 && door == 0 {
		//Зеленая улица 4
		return 4
	}
	if (rezim == 5 || rezim == 6) && (err == 0 || err == 1) && (faza >= 1 && faza <= 9) && door == 0 {
		//Локальное управление 5
		return 5
	}
	if (rezim == 9 || rezim == 4 || rezim == 6) && (err == 0 || err == 1) && (faza == 0) && door == 0 {
		//Локальное управление 5
		return 5
	}

	if (rezim == 8 || rezim == 9) && (err == 0 || err == 1) && (faza == 10) && door == 0 {
		//Желтое мигание по расписанию 6
		return 6
	}
	if rezim == 4 && (err == 0 || err == 1) && (faza == 10) {
		//Желтое мигание из центра 7
		return 7
	}
	if (rezim == 1 || rezim == 2) && (err == 0 || err == 1) && faza == 10 {
		//Желтое мигание заданное на перекрестке 8
		return 8
	}
	if (rezim == 5 || rezim == 6) && (err == 0 || err == 1) && faza == 10 && door == 0 {
		//Желтое мигание по расписанию 9
		return 9
	}
	if (rezim == 8 || rezim == 9 || rezim == 4) && (err == 0 || err == 1) && faza == 12 && door == 0 {
		//Кругом красный 10
		return 10
	}
	if (rezim == 8 || rezim == 9) && (err == 0 || err == 1) && faza == 11 && door == 0 {
		//Отключение светофора по расписанию 11
		return 11
	}
	if rezim == 4 && (err == 0 || err == 1) && faza == 11 {
		//Желтое мигание заданное из центра 12 1
		return 12
	}
	if (rezim == 1 || rezim == 2) && (err == 0 || err == 1) && (faza == 11 || faza == 0) {
		//Отключение светофора заданное на перекрестке 13
		return 13
	}
	if (rezim == 5 || rezim == 6) && (err == 0 || err == 1) && faza == 11 && door == 0 {
		//Отключение светофора по расписанию ДК 14
		return 14
	}
	if err == 11 && dev == 3 {
		//Авария 220 16
		return 16
	}
	if err == 11 && dev == 5 { //Спросить как узнать ошибки контроллера УСДК
		//Выключен УСДК/ДК 17
		return 17
	}
	if err == 11 && dev == 4 { //Спросить как узнать ошибки GPRS
		//Нет связи с УСДК 18
		return 18
	}
	if err == 11 && dev == 8 { //Спросить как узнать ошибки ПБС УСДК
		//Нет связи с ПСПД 19
		return 19
	}
	if err == 11 {
		//Обрыв ЛС КЗЦ 20
		return 20
	}
	if err == 4 && dev == 8 { //Спросить как узнать ошибки ПБС УСДК
		//Превышение трафика
		return 21
	}
	if (rezim == 5 || rezim == 6) && err == 4 && (dev == 0 || dev == 1) {
		//Базовая привязка 22
		return 22
	}
	if (rezim == 1 || rezim == 2) && err == 4 && dev == 4 {
		//Неисправность часов или GPS 22
		return 23
	}
	if (rezim == 5 || rezim == 6) && err == 4 && (dev == 4 || dev == 5) {
		//Коррекция привязки 24
		return 24
	}
	if err == 10 {
		//Несуществующая фаза
		return 25
	}
	if err == 4 {
		//Несуществующий код
		return 26
	}
	if (rezim == 8 || rezim == 9) && (faza > 0 && faza < 10) && lamp == 1 {
		//Координированное управление и перегоревшая лампа
		return 27
	}
	if err == 2 {
		//Обрыв линий связи
		return 28
	}
	if err == 3 {
		//Негоден по паритету
		return 29
	}
	if err == 5 && faza == 11 {
		//Отключен из-за конфликта направлений
		return 30
	}
	if err == 5 {
		//Конфликт направлений
		return 31
	}
	if err == 6 && faza == 10 {
		//Желтое мигание из-за перегорания
		return 32
	}
	if err == 6 {
		//Не годен по перегоранию ламп
		return 33
	}
	if err == 7 {
		//Не включается в координацию
		return 34
	}
	if err == 8 {
		//Дорожный контроллер не подчиняется командам
		return 35
	}
	if err == 9 {
		//Длинный промежуточный такт
		return 36
	}
	if err == 12 {
		//Обрыв линий связи ЭВМ с перекрестками
		return 37
	}
	if err == 3 {
		//Нет информации о работе перекрестка
		return 38
	}
	if door != 0 {
		//Двери открыты 15
		return 15
	}
	logger.Debug.Printf("Режим=%v Фаза=%v Ошибка=%v Устройство=%v Лампа=%v Дверь=%v ID %v", rezim, faza, err, dev, lamp, door, cc.ID)
	return 39
}

func (cc *Controller) lamps() int {
	if cc.DK.LDK == 0 {
		return 0
	}
	return 1
}
func (cc *Controller) doors() int {
	if !cc.DK.ODK {
		return 0
	}
	return 1
}
func (cc *Controller) codeDevice() int {
	return cc.DK.DDK
}
