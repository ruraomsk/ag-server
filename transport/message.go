package transport

import (
	"fmt"
	"reflect"
	"strconv"
)

// SubMessage структура для хранения сообщений
type SubMessage struct {
	Type    uint8
	Message []uint8
}

// ToString вывод в строку
func (s *SubMessage) ToString() string {
	res := fmt.Sprintf("Type %d [", s.Type)
	for _, code := range s.Message {
		res += strconv.FormatInt(int64(code), 16) + " "
	}
	res += "]"
	return res
}

// Compare сравнение сообщений
func (s *SubMessage) Compare(ss *SubMessage) bool {
	return reflect.DeepEqual(s, ss)
}

// ParseMessage разбор буфера сообщений от сервера
func (s *HeaderServer) ParseMessage() []SubMessage {
	var sb SubMessage
	sub := make([]SubMessage, 0)
	pos := 0
	for pos < len(s.Message) {
		sb.Type = s.Message[pos]
		pos++
		sb.Message = make([]uint8, int(s.Message[pos]))
		pos++
		for i := 0; i < len(sb.Message); i++ {
			sb.Message[i] = s.Message[pos]
			pos++
		}
		sub = append(sub, sb)
	}
	return sub
}

// UpackMessages записывает сообщения сервера для передачи
func (s *HeaderServer) UpackMessages(subs []SubMessage) error {
	//Считаем общую длину буфера
	s.SubMessages = subs
	l := 0
	for _, sb := range subs {
		if s.Code != 0x7f {
			l += len(sb.Message) + 2
		} else {
			l += len(sb.Message)
		}
	}
	if l > 255 {
		return fmt.Errorf("общая длина сообщения больше 255")
	}
	s.Message = make([]uint8, l)
	pos := 0
	for _, sb := range subs {
		if s.Code != 0x7f {
			s.Message[pos] = sb.Type
			pos++
			s.Message[pos] = uint8(len(sb.Message))
			pos++
		}
		for i := 0; i < len(sb.Message); i++ {
			s.Message[pos] = sb.Message[i]
			pos++
		}
	}
	return nil
}

// codeParse разбор ответов от устройства
func (d *HeaderDevice) codeParse(pos int) (SubMessage, int) {
	var sb SubMessage
	sb.Type = d.Message[pos]
	l := 0
	switch sb.Type {
	case 0x00:
		pos++
		sb.Message = make([]uint8, 0)
		return sb, pos
	case 0x01:
		l = 6
	case 0x04:
		l = 5
	case 0x07:
		l = 5
	case 0x08:
		l = 14
	case 0x09:
		l = 6
	case 0x0a:
		l = int(d.Message[pos+2]) + 3
		if l == 0 {
			l = 3
		}
	case 0x0b:
		l = 16
	case 0x0f:
		l = 22
	case 0x10:
		l = 6
	case 0x11:
		l = 4
	case 0x12:
		l = 23
	case 0x13:
		l = int(d.Message[pos+2]) + 3
	case 0x1b:
		l = 9
	case 0x1d:
		l = 13
	case 0x1c:
		l = 6
	default:
		fmt.Println("!!!!!!")
		pos++
		sb.Message = make([]uint8, 0)
		return sb, pos
	}
	// pos++
	sb.Message = make([]uint8, l)
	for i := 0; i < len(sb.Message); i++ {
		sb.Message[i] = d.Message[pos]
		pos++
	}
	return sb, pos
}

// ParseMessage разбор буфера сообщений от устройства
func (d *HeaderDevice) ParseMessage() []SubMessage {
	var sb SubMessage
	// logger.Debug.Printf("hdev %v", d)
	sub := make([]SubMessage, 0)
	pos := 0
	for pos < len(d.Message) {
		sb, pos = d.codeParse(pos)
		// logger.Debug.Printf("mess %v %d", sb, pos)
		sub = append(sub, sb)
	}
	// logger.Debug.Printf("all mess %v ", sub)

	return sub
}

// UpackMessages записывает сообщения устройства для передачи
func (d *HeaderDevice) UpackMessages(subs []SubMessage) error {
	//Считаем общую длину буфера
	l := 0
	for _, sb := range subs {
		l += len(sb.Message)
	}
	if l > 255 {
		return fmt.Errorf("общая длина сообщения больше 255")
	}
	d.Message = make([]uint8, l)
	pos := 0
	for _, sb := range subs {
		for i := 0; i < len(sb.Message); i++ {
			d.Message[pos] = sb.Message[i]
			pos++
		}
	}
	return nil
}
