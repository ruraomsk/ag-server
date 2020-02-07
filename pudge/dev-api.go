package pudge

//Make создает байт для передачи статуса команд ДУ
func (s *StatusCommandDU) Make() (command uint8) {
	command = 0
	if s.IsPK {
		command |= 1
	}
	if s.IsPKS {
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
	s.IsPKS = false
	s.IsNK = false
	s.IsDUDK1 = false
	s.IsDUDK2 = false
	s.IsReqSFDK1 = false
	s.IsReqSFDK2 = false
	if command&1 != 0 {
		s.IsPK = true
	}
	if command&2 != 0 {
		s.IsPKS = true
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
	d.RDK = int(buffer[pos] & 0x7)
	d.FDK = int(buffer[pos]&0x70) >> 4
	pos++
	d.DDK = int(buffer[pos] & 0x7)
	d.EDK = int(buffer[pos]&0x70) >> 4
	pos++
	if (buffer[pos] & 0x80) != 0 {
		d.PDK = true
	} else {
		d.PDK = false
	}
	d.EEDK = int(buffer[pos] & 0x7)
	pos++
	if (buffer[pos] & 0x80) != 0 {
		d.ODK = true
	} else {
		d.ODK = false
	}
	d.LDK = int(buffer[pos] & 0x7)
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
	buffer[pos] |= uint8(d.RDK & 0x7)
	buffer[pos] |= uint8((d.FDK & 0x7) << 4)
	pos++
	buffer[pos] |= uint8(d.DDK & 0x7)
	buffer[pos] |= uint8((d.EDK & 0x7) << 4)
	pos++
	if d.PDK {
		buffer[pos] |= 0x80
	}
	buffer[pos] |= uint8(d.EEDK & 0x7)
	pos++
	if d.ODK {
		buffer[pos] |= 0x80
	}
	buffer[pos] |= uint8(d.LDK & 0x7)
	pos++
	buffer[pos] = uint8(d.FTUDK)
	pos++
	buffer[pos] = uint8(d.TDK)
	pos++
	buffer[pos] = uint8(d.FTSDK)
	pos++
	buffer[pos] = uint8(d.TTCDK)
}
func (cc *Controller) calcStatus() int {
	rezim := cc.TexRezim
	faza := cc.DK.FDK
	err := cc.coderr()
	dev := cc.codeDevice()
	lamp := cc.lamps()
	door := cc.doors()
	// if cc.LastOperation
	if !cc.IsConnected() {
		return 18
	}
	if cc.Base {
		return 23
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
	if rezim == 3 && (err == 0 || err == 1) && (faza >= 1 && faza <= 9) && lamp == 0 && door == 0 {
		//Зеленая улица 4
		return 4
	}
	if (rezim == 5 || rezim == 6) && (err == 0 || err == 1) && (faza >= 1 && faza <= 9) && door == 0 {
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
	if (rezim == 8 || rezim == 9) && (err == 0 || err == 1) && faza == 12 && door == 0 {
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
	if (rezim == 1 || rezim == 2) && (err == 0 || err == 1) && faza == 11 {
		//Отключение светофора заданное на перекрестке 13
		return 13
	}
	if (rezim == 5 || rezim == 6) && (err == 0 || err == 1) && faza == 11 && door == 0 {
		//Отключение светофора по расписанию ДК 14
		return 14
	}
	if door != 0 {
		//Двери открыты 15
		return 15
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
		//Базовая привязка 23
		return 23
	}

	return 1
}
func (cc *Controller) coderr() int {
	if cc.DK.EDK != 0 {
		return cc.DK.EDK
	}
	if cc.DK.PDK {
		return 1
	}
	return 0
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
