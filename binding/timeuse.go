package binding

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

//SetTimeUse хранение настроек внешних входов
type SetTimeUse struct {
	Uses       []Use `json:"uses"`
	IntervalTE int   `json:"ite"` //Интервал между ТЕ
	//Tuin       int   `json:"tuin"` //Т уср ИН
	MGRNotWork []int `json:"notwork"`
}

//Use один вход
type Use struct {
	Name  string  `json:"name"`
	Type  int     `json:"type"`
	Tvps  int     `json:"tvps"`
	Dk    int     `json:"dk"`
	Fazes string  `json:"fazes"`
	Long  float32 `json:"long"`
}

//Compare сравнивание истина если равны
func (s *SetTimeUse) Compare(ss *SetTimeUse) bool {
	return reflect.DeepEqual(s, ss)
}

//NewSetTimeUse создание нового описания
func NewSetTimeUse() *SetTimeUse {
	r := new(SetTimeUse)
	r.MGRNotWork = make([]int, 8)
	r.Uses = make([]Use, 18)
	for i := 0; i < len(r.Uses); i++ {
		if i == 0 {
			r.Uses[i].Name = "1 ТВП"
		}
		if i == 1 {
			r.Uses[i].Name = "2 ТВП"
		}
		if i > 1 {
			r.Uses[i].Name = fmt.Sprintf("%d вх", i-1)
		}
	}
	return r
}

// func (u *Use) isEmpty() bool {
// 	if u.Type != 0 || u.Tvps != 0 || int(u.Long*10) != 0 || len(u.Fazes) != 0 {
// 		return false
// 	}
// 	return false
// }

//IsEmpty вернет истину если весь набор пустой
func (s *SetTimeUse) IsEmpty() bool {
	// if s.IntervalTE != 0 {
	// 	return false
	// }
	// for _, ss := range s.Uses {
	// 	if !ss.isEmpty() {
	// 		return false
	// 	}
	// }
	// return true
	return false
}

//FromBuffer загружает из буфера
func (s *SetTimeUse) FromBuffer(buffer []int) error {
	if buffer[2] == 23 {
		//Считываем массив 148
		if buffer[0] != 148 {
			return fmt.Errorf("несовпал номер массива на сервере и номер массива")
		}
		if len(buffer) < 32 {
			return fmt.Errorf("слишком маленький массив")
		}
		if len(buffer) == 32 {
			s.Uses = s.Uses[0:8]
		}
		if len(buffer) == 38 {
			s.Uses = s.Uses[0:18]
		}
		pos := 5
		for i := 0; i < len(s.Uses); i++ {
			tp := 0
			tvps := 0
			dk := 1
			if pos >= len(buffer) {
				return nil
			}
			if buffer[pos]&1 != 0 {
				tp = 1
				tvps = 3
			} else {
				if buffer[pos]&128 != 0 {
					tp = 0
					tvps = 1
				} else {
					if buffer[pos]&2 != 0 {
						tp = 0
						tvps = 2
					}
				}
			}
			if tp == 1 {
				if buffer[pos]&8 != 0 {
					dk = 1
				}
				if buffer[pos]&2 != 0 {
					dk = 2
				}
			} else {
				if tvps == 1 || tvps == 2 {
					if buffer[pos]&8 != 0 {
						dk = 1
					} else {
						dk = 2
					}
				}
			}
			if dk == 2 {
				dk = 1
			}
			fazes := ""
			mask := 1
			for j := 1; j < 9; j++ {
				if buffer[pos+1]&mask != 0 {
					fazes += strconv.Itoa(j) + ","
				}
				mask = mask << 1
			}
			fazes = strings.TrimSuffix(fazes, ",")
			j := 0
			if len(s.Uses) > 8 {
				j = i
			} else {
				if i < len(s.Uses)-2 {
					j = i + 2
				} else {
					j = i - 6
					if j < 0 {
						j = 0
					}
				}
			}
			s.Uses[j].Type = tp
			s.Uses[j].Tvps = tvps
			s.Uses[j].Dk = dk
			s.Uses[j].Fazes = fazes
			if tvps == 3 {
				s.Uses[j].Long = float32(buffer[pos+2]) / 10.0
			} else {
				s.Uses[j].Long = float32(buffer[pos+2])
			}
			pos += 3
		}
		s.IntervalTE = buffer[pos]
		return nil
	}
	if buffer[2] == 20 {
		//Считываем массив 157
		if buffer[0] != 157 {
			return fmt.Errorf("несовпал номер массива на сервере и номер массива")
		}
		pos := 5
		for i := 0; i < len(s.MGRNotWork); i++ {
			s.MGRNotWork[i] = buffer[pos]
			pos++
		}
		return nil
	}
	return fmt.Errorf("неверный номер массива")
}

//ToBuffer выгружает в буффер num номер массива на сервере
func (s *SetTimeUse) ToBuffer(num int) []int {
	if num == 157 {
		r := make([]int, 13)
		r[0] = 157
		r[2] = 20
		r[3] = 9
		pos := 5
		for i := 0; i < len(s.MGRNotWork); i++ {
			r[pos] = s.MGRNotWork[i]
			pos++
		}
		return r
	}
	r := make([]int, len(s.Uses)*3+8)
	r[0] = 148
	r[2] = 23
	r[3] = len(s.Uses)*3 + 4
	pos := 5
	for i := 0; i < len(s.Uses); i++ {
		r = makeElem(r, pos, s.Uses[i])
		pos += 3
	}
	r[pos] = s.IntervalTE
	return r
}
func makeElem(r []int, pos int, ss Use) []int {
	r[pos] = 0
	r[pos+1] = 0
	r[pos+2] = 0

	if ss.Dk == 1 {
		r[pos] |= 8
	}
	if ss.Dk == 2 {
		r[pos] |= 2
	}
	if ss.Tvps == 1 {
		r[pos] |= 130
	}
	if ss.Tvps == 2 {
		r[pos] |= 2
	}
	if ss.Tvps == 3 {
		r[pos] |= 1
	}
	if ss.Tvps == 4 {
		r[pos] |= 16
	}
	if ss.Tvps == 5 {
		r[pos] |= 64
	}
	if ss.Tvps == 6 {
		r[pos] |= 128
	}
	fs := strings.Split(ss.Fazes, ",")
	for _, ff := range fs {
		if len(ff) == 0 {
			continue
		}
		f, _ := strconv.Atoi(ff)
		if f == 0 {
			continue
		}
		if ss.Tvps == 4 {
			r[pos+1] = f
		} else {
			f--
			r[pos+1] |= 1 << f
		}
	}
	if ss.Tvps == 3 {
		r[pos+2] = int(ss.Long * 10)
	} else {
		r[pos+2] = int(ss.Long)
	}
	return r
}
