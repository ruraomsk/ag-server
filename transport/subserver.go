package transport

import (
	"fmt"
	"strconv"
	"strings"
)

// GetCodeCommandServer возвращает номер команды  или ноль если это массив привязки
func (s *SubMessage) GetCodeCommandServer() int {
	if s.Type != 0 {
		//Это не команда это массив привязки
		return 0
	}
	return int(s.Message[0])
}

// Get0x01Server получение параметров команды -1 ошибки
func (s *SubMessage) Get0x01Server() int {
	if s.Type != 0 {
		//Это не команда это массив привязки
		return -1
	}
	if s.GetCodeCommandServer() != 0x01 {
		return -1
	}
	return int(s.Message[1])
}

// Get0x02Server получение параметров команды -1 ошибки
func (s *SubMessage) Get0x02Server() bool {
	if s.Type != 0 {
		//Это не команда это массив привязки
		return false
	}
	if s.GetCodeCommandServer() != 0x02 {
		return false
	}

	return s.Message[1] == 2
}

// Get0x04Server получение параметров команды -1 ошибки
func (s *SubMessage) Get0x04Server() [2]bool {
	res := [2]bool{false, false}
	if s.Type != 0 {
		//Это не команда это массив привязки
		return res
	}
	if s.GetCodeCommandServer() != 0x04 {
		return res
	}
	res[0] = (s.Message[1] & 1) != 0
	res[1] = (s.Message[1] & 2) != 0
	return res
}

// Get0x05Server получение параметров команды -1 ошибки
func (s *SubMessage) Get0x05Server() int {
	if s.Type != 0 {
		//Это не команда это массив привязки
		return -1
	}
	if s.GetCodeCommandServer() != 0x05 {
		return -1
	}
	return int(s.Message[1])
}

// Get0x06Server получение параметров команды -1 ошибки
func (s *SubMessage) Get0x06Server() int {
	if s.Type != 0 {
		//Это не команда это массив привязки
		return -1
	}
	if s.GetCodeCommandServer() != 0x06 {
		return -1
	}
	return int(s.Message[1])
}

// Get0x07Server получение параметров команды -1 ошибки
func (s *SubMessage) Get0x07Server() int {
	if s.Type != 0 {
		//Это не команда это массив привязки
		return -1
	}
	if s.GetCodeCommandServer() != 0x07 {
		return -1
	}
	return int(s.Message[1])
}

// Get0x09Server получение параметров команды -1 ошибки
func (s *SubMessage) Get0x09Server() int {
	if s.Type != 0 {
		//Это не команда это массив привязки
		return -1
	}
	if s.GetCodeCommandServer() != 0x09 {
		return -1
	}
	return int(s.Message[1])
}

// Get0x0AServer получение параметров команды -1 ошибки
func (s *SubMessage) Get0x0AServer() int {
	if s.Type != 0 {
		//Это не команда это массив привязки
		return -1
	}
	if s.GetCodeCommandServer() != 0x0a {
		return -1
	}
	return int(s.Message[1])
}

// Get0x0BServer получение параметров команды -1 ошибки
func (s *SubMessage) Get0x0BServer() [2]int {
	res := [2]int{-1, -1}
	if s.Type != 0 {
		//Это не команда это массив привязки
		return res
	}
	if s.GetCodeCommandServer() != 0x0B {
		return res
	}
	res[0] = int(s.Message[1])
	res[1] = int(s.Message[2])
	return res
}

// Get0x32Server получение параметров команды -1 ошибки
func (s *SubMessage) Get0x32Server() (ip string, port int, err error) {
	if s.Type != 0 {
		//Это не команда это массив привязки
		return "", 0, fmt.Errorf("массив привязки")
	}
	if s.GetCodeCommandServer() != 0x32 {
		return "", 0, fmt.Errorf("не команда 0x32")
	}
	var bb []byte
	for i := 1; i < len(s.Message)-5; i++ {
		bb = append(bb, s.Message[i])
	}
	ip = string(bb)
	bb = make([]byte, 0)
	for i := 17; i < len(s.Message); i++ {
		bb = append(bb, s.Message[i])
	}
	port, _ = strconv.Atoi(string(bb))
	return
}

// Get0x33Server получение параметров команды -1 ошибки
func (s *SubMessage) Get0x33Server() int {
	if s.Type != 0 {
		//Это не команда это массив привязки
		return -1
	}
	if s.GetCodeCommandServer() != 0x33 {
		return -1
	}
	return int(s.Message[1]) | (int(s.Message[2]) << 8)
}

// Get0x34Server получение параметров команды -1 ошибки
func (s *SubMessage) Get0x34Server() bool {
	if s.Type != 0 {
		//Это не команда это массив привязки
		return false
	}
	if s.GetCodeCommandServer() != 0x34 {
		return false
	}

	return s.Message[1] == 1
}

// Get0x35Server получение параметров команды -1 ошибки
func (s *SubMessage) Get0x35Server() (int, bool) {
	if s.Type != 0 {
		//Это не команда это массив привязки
		return -1, false
	}
	if s.GetCodeCommandServer() != 0x35 {
		return -1, false
	}

	return int(s.Message[1] & 0x7f), (s.Message[1] & 0x80) != 0
}

// GetArray возвращает номер и массив от сервера
func (s *SubMessage) GetArray() (int, int, []int) {
	res := []int{-1}
	if s.Type == 0 {
		//Это не команда это массив привязки
		return -1, 0, res
	}
	res = make([]int, len(s.Message)-1)
	for i := 1; i < len(s.Message); i++ {
		res[i-1] = int(s.Message[i])
	}
	return int(s.Type), int(s.Message[0]), res
}

// Set0x01Server записывает субсообщение для команды с номером в имени
func (s *SubMessage) Set0x01Server(num int) {
	//num номер сообщения
	s.Type = 0
	s.Message = make([]uint8, 2)
	s.Message[0] = 0x01
	s.Message[1] = uint8(num)
}

// Set0x02Server записывает субсообщение для команды с номером в имени
func (s *SubMessage) Set0x02Server(flag bool) {
	// flag false - отключить true включить
	s.Type = 0
	s.Message = make([]uint8, 2)
	s.Message[0] = 0x02
	s.Message[1] = 1
	if flag {
		s.Message[1] = 2
	}
}

// Set0x03Server записывает субсообщение для команды с номером в имени
func (s *SubMessage) Set0x03Server() {
	//num номер сообщения
	s.Type = 0
	s.Message = make([]uint8, 1)
	s.Message[0] = 0x03
}

// Set0x04Server записывает субсообщение для команды с номером в имени
func (s *SubMessage) Set0x04Server(dk1, dk2 bool) {
	//dk1 ДК1
	//dk2 ДК2
	s.Type = 0
	s.Message = make([]uint8, 2)
	s.Message[0] = 0x04
	set := 0
	if dk1 {
		set = set | 1
	}
	if dk2 {
		set = set | 2
	}
	s.Message[1] = uint8(set)
}

// Set0x05Server записывает субсообщение для команды с номером в имени
func (s *SubMessage) Set0x05Server(num int) {
	//num номер плана координации
	s.Type = 0
	s.Message = make([]uint8, 2)
	s.Message[0] = 0x05
	s.Message[1] = uint8(num)
}

// Set0x06Server записывает субсообщение для команды с номером в имени
func (s *SubMessage) Set0x06Server(num int) {
	//num номер карты по времени суток
	s.Type = 0
	s.Message = make([]uint8, 2)
	s.Message[0] = 0x06
	s.Message[1] = uint8(num)
}

// Set0x07Server записывает субсообщение для команды с номером в имени
func (s *SubMessage) Set0x07Server(num int) {
	//num номер карты недельной
	s.Type = 0
	s.Message = make([]uint8, 2)
	s.Message[0] = 0x07
	s.Message[1] = uint8(num)
}

// Set0x09Server записывает субсообщение для команды с номером в имени
func (s *SubMessage) Set0x09Server(num int) {
	//num номер режима по ДК1
	s.Type = 0
	s.Message = make([]uint8, 2)
	s.Message[0] = 0x09
	s.Message[1] = uint8(num)
}

// Set0x0AServer записывает субсообщение для команды с номером в имени
func (s *SubMessage) Set0x0AServer(num int) {
	//num номер режима по ДК2
	s.Type = 0
	s.Message = make([]uint8, 2)
	s.Message[0] = 0x0A
	s.Message[1] = uint8(num)
}

// Set0x0BServer записывает субсообщение для команды с номером в имени
func (s *SubMessage) Set0x0BServer(m, n int) {
	s.Type = 0
	s.Message = make([]uint8, 3)
	s.Message[0] = 0x0B
	s.Message[1] = uint8(m)
	s.Message[2] = uint8(n)

}
func (s *SubMessage) Set0x0EServer(mgr int) {
	s.Type = 0
	s.Message = make([]uint8, 3)
	s.Message[0] = 0x0E
	s.Message[1] = uint8((mgr >> 8) & 0xff)
	s.Message[2] = uint8(mgr & 0xff)

}
func (s *SubMessage) Set0x0FServer(num int) {
	//num номер режима по ДК2
	s.Type = 0
	s.Message = make([]uint8, 2)
	s.Message[0] = 0x0F
	s.Message[1] = uint8(num)
}

// Set0x32Server записывает субсообщение для команды с номером в имени
func (s *SubMessage) Set0x32Server(ip string, port int) error {
	s.Type = 0
	s.Message = make([]uint8, 21)
	s.Message[0] = 0x32
	st := strings.Split(ip, ".")
	if len(st) != 4 {
		return fmt.Errorf("неверно задан ip %s", ip)
	}
	if port > 9999 {
		return fmt.Errorf("неверно задан port %d", port)
	}
	for i := 0; i < len(st); i++ {
		for len(st[i]) < 3 {
			st[i] = "0" + st[i]
		}
	}
	p := fmt.Sprintf("%04d", port)
	rs := ""
	for _, ss := range st {
		rs += ss
		rs += "."
	}
	rs += p
	bs := []byte(rs)
	pos := 1
	for _, b := range bs {
		s.Message[pos] = b
		pos++
	}
	// logger.Debug.Printf("0x32 %v", s.Message)
	return nil
}

// Set0x33Server записывает субсообщение для команды с номером в имени
func (s *SubMessage) Set0x33Server(min int) {
	s.Type = 0
	s.Message = make([]uint8, 3)
	s.Message[0] = 0x33
	s.Message[1] = uint8(min & 0xff)
	s.Message[2] = uint8((min >> 8) & 0xff)
}

// Set0x34Server записывает субсообщение для команды с номером в имени
func (s *SubMessage) Set0x34Server(rez bool) {
	s.Type = 0
	s.Message = make([]uint8, 2)
	s.Message[0] = 0x34
	s.Message[1] = 0
	if rez {
		s.Message[1] = 1
	}
}

// Set0x35Server записывает субсообщение для команды с номером в имени
func (s *SubMessage) Set0x35Server(interval int, ignor bool) {
	s.Type = 0
	s.Message = make([]uint8, 2)
	s.Message[0] = 0x35
	s.Message[1] = uint8(interval)
	if ignor {
		s.Message[1] |= 0x80
	}
}

// SetArray возвращает номер и массив от сервера
func (s *SubMessage) SetArray(num int, nelem int, array []int) {
	s.Type = uint8(num)
	if num == 133 {
		s.Type = 5
	}
	if num == 137 {
		s.Type = 9
	}
	s.Message = make([]uint8, len(array)-1)
	s.Message[0] = uint8(nelem)
	for i := 2; i < len(array); i++ {
		s.Message[i-1] = uint8(array[i])
	}
	//logger.Debug.Printf("massiv :%v",s)
}
