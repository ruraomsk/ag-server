package transport

//GetCodeCommandServer возвращает номер команды  или ноль если это массив привязки
func (s *SubMessage) GetCodeCommandServer() int {
	if s.Type != 0 {
		//Это не команда это массив привязки
		return 0
	}
	return int(s.Message[0])
}

//Get0x01Server получение параметров команды -1 ошибки
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

//Get0x02Server получение параметров команды -1 ошибки
func (s *SubMessage) Get0x02Server() int {
	if s.Type != 0 {
		//Это не команда это массив привязки
		return -1
	}
	if s.GetCodeCommandServer() != 0x02 {
		return -1
	}
	return int(s.Message[1])
}

//Get0x04Server получение параметров команды -1 ошибки
func (s *SubMessage) Get0x04Server() [2]int {
	res := [2]int{-1, -1}
	if s.Type != 0 {
		//Это не команда это массив привязки
		return res
	}
	if s.GetCodeCommandServer() != 0x04 {
		return res
	}
	res[0] = int(s.Message[1] & 1)
	res[1] = int(s.Message[1] & 2)
	return res
}

//Get0x05Server получение параметров команды -1 ошибки
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

//Get0x06Server получение параметров команды -1 ошибки
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

//Get0x07Server получение параметров команды -1 ошибки
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

//Get0x09Server получение параметров команды -1 ошибки
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

//Get0x0AServer получение параметров команды -1 ошибки
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

//Get0x0BServer получение параметров команды -1 ошибки
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

//GetArray возвращает номер и массив от сервера
func (s *SubMessage) GetArray() (int, []int) {
	res := []int{-1}
	if s.Type == 0 {
		//Это не команда это массив привязки
		return -1, res
	}
	res = make([]int, len(s.Message))
	for i := 0; i < len(s.Message); i++ {
		res[i] = int(s.Message[i])
	}
	return int(s.Type), res
}

//Set0x01Server записывает субсообщение для команды с номером в имени
func (s *SubMessage) Set0x01Server(num int) {
	//num номер сообщения
	s.Type = 1
	s.Message = make([]uint8, 2)
	s.Message[0] = 0x01
	s.Message[1] = uint8(num)
}

//Set0x02Server записывает субсообщение для команды с номером в имени
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

//Set0x03Server записывает субсообщение для команды с номером в имени
func (s *SubMessage) Set0x03Server() {
	//num номер сообщения
	s.Type = 0
	s.Message = make([]uint8, 1)
	s.Message[0] = 0x03
}

//Set0x04Server записывает субсообщение для команды с номером в имени
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

//Set0x05Server записывает субсообщение для команды с номером в имени
func (s *SubMessage) Set0x05Server(num int) {
	//num номер плана координации
	s.Type = 0
	s.Message = make([]uint8, 2)
	s.Message[0] = 0x05
	s.Message[1] = uint8(num)
}

//Set0x06Server записывает субсообщение для команды с номером в имени
func (s *SubMessage) Set0x06Server(num int) {
	//num номер карты по времени суток
	s.Type = 0
	s.Message = make([]uint8, 2)
	s.Message[0] = 0x06
	s.Message[1] = uint8(num)
}

//Set0x07Server записывает субсообщение для команды с номером в имени
func (s *SubMessage) Set0x07Server(num int) {
	//num номер карты недельной
	s.Type = 0
	s.Message = make([]uint8, 2)
	s.Message[0] = 0x07
	s.Message[1] = uint8(num)
}

//Set0x09Server записывает субсообщение для команды с номером в имени
func (s *SubMessage) Set0x09Server(num int) {
	//num номер режима по ДК1
	s.Type = 0
	s.Message = make([]uint8, 2)
	s.Message[0] = 0x09
	s.Message[1] = uint8(num)
}

//Set0x0AServer записывает субсообщение для команды с номером в имени
func (s *SubMessage) Set0x0AServer(num int) {
	//num номер режима по ДК2
	s.Type = 0
	s.Message = make([]uint8, 2)
	s.Message[0] = 0x0A
	s.Message[1] = uint8(num)
}

//Set0x0BServer записывает субсообщение для команды с номером в имени
func (s *SubMessage) Set0x0BServer(m, n int) {
	s.Type = 0
	s.Message = make([]uint8, 3)
	s.Message[0] = 0x0B
	s.Message[1] = uint8(m)
	s.Message[2] = uint8(n)

}

//SetArray возвращает номер и массив от сервера
func (s *SubMessage) SetArray(num int, array []int) {
	s.Type = uint8(num)
	s.Message = make([]uint8, len(array))
	for i := 0; i < len(s.Message); i++ {
		s.Message[i] = uint8(array[i])
	}
}
