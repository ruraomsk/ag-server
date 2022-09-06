package binding

import (
	"fmt"
	"reflect"
	"strconv"
)

//DaySets все суточные планы
type DaySets struct {
	DaySets []*OneDay `json:"daysets"`
}

func (ds *DaySets) GetPKs(nd int) []int {
	result := make([]int, 0)
	pks := make(map[int]int)
	for _, od := range ds.DaySets {
		if od.Number != nd {
			continue
		}
		for _, l := range od.Lines {
			if l.PKNom != 0 {
				pks[l.PKNom] = l.PKNom
			}
		}
		break
	}
	for _, p := range pks {
		result = append(result, p)
	}
	return result
}

func (ds *DaySets) String() string {
	r := "DaySets\n"
	for _, d := range ds.DaySets {
		r += d.String() + "\n"
	}
	return r
}

//Compare сравнивание истина если равны
func (ds *DaySets) Compare(dd *DaySets) bool {
	return reflect.DeepEqual(ds, dd)
}

//FromBuffer загружает суточные планы из буфера
func (ds *DaySets) FromBuffer(buffer []int) error {
	d, err := dayFromBuffer(buffer)
	if err != nil {
		return err
	}
	ds.DaySets[d.Number-1] = d
	return nil
}

//NewDaySet создание нового набора суточных планов
func NewDaySet() *DaySets {
	r := new(DaySets)
	r.DaySets = make([]*OneDay, 12)
	for index := 0; index < len(r.DaySets); index++ {
		r.DaySets[index] = NewOneDay(index + 1)
	}
	return r
}

//OneDay Один день плана
type OneDay struct {
	Number int     `json:"num"`
	Count  int     `json:"count"` //Счетчик переключений
	Lines  []*Line `json:"lines"`
}

func (od *OneDay) String() string {
	r := strconv.Itoa(od.Number) + ":" + strconv.Itoa(od.Count) + " "
	for _, l := range od.Lines {
		r += fmt.Sprintf("[%d %d %d]", l.PKNom, l.Hour, l.Min)
	}
	return r
}

//Line структура одного периода времени
type Line struct {
	PKNom int `json:"npk"`
	Hour  int `json:"hour"`
	Min   int `json:"min"`
}

func (l *Line) isEmpty() bool {
	// if l.PKNom != 0 || l.Hour != 0 || l.Min != 0 {
	// 	return false
	// }
	return false
}

func NewOneDay(number int) *OneDay {
	r := new(OneDay)
	r.Number = number
	r.Lines = make([]*Line, 12)
	for index := 0; index < len(r.Lines); index++ {
		l := new(Line)
		l.PKNom = 0
		l.Hour = 0
		l.Min = 0
		r.Lines[index] = l
	}
	return r
}

//IsEmpty возвращает истину если данный недельный массив пустой
func (od *OneDay) IsEmpty() bool {
	for _, l := range od.Lines {
		if !l.isEmpty() {
			return false
		}
	}
	return true
}

//ToBuffer переводит из суточного массива в буфер кодов
func (od *OneDay) ToBuffer() []int {
	r := make([]int, 43)
	r[0] = od.Number + 64
	r[1] = 0
	r[2] = 137
	r[3] = 39
	r[4] = od.Number
	r[5] = od.Count
	pos := 6
	for _, l := range od.Lines {
		r[pos] = l.PKNom
		pos++
		r[pos] = l.Hour
		pos++
		r[pos] = l.Min
		pos++
	}
	r[42] = 255
	return r
}

//dayFromBuffer переводит из массива кодов в недельный массив
func dayFromBuffer(buffer []int) (*OneDay, error) {
	if len(buffer) != 43 {
		return nil, fmt.Errorf("неверная длина массива")
	}
	if buffer[2] != 137 {
		return nil, fmt.Errorf("несовпал номер массива на сервере и номер массива")
	}
	if buffer[0]-64 <= 0 || buffer[0]-64 > 12 {
		return nil, fmt.Errorf("неверный номер массива")
	}
	d := NewOneDay(buffer[0] - 64)
	d.Count = buffer[5]
	for index := 0; index < len(d.Lines); index++ {
		l := new(Line)
		l.PKNom = buffer[6+(index*3)]
		l.Hour = buffer[6+(index*3)+1]
		l.Min = buffer[6+(index*3)+2]
		d.Lines[index] = l
	}
	return d, nil
}
