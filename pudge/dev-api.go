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
