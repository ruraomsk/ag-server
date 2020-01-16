package binding

import (
	"fmt"
	"reflect"
)

//SetCtrl массив контроля входов
type SetCtrl struct {
	Stage []CtrlStage
}

//StageTime время в массиве контроля
type StageTime struct {
	Hour   int `json:"hour"`
	Minute int `json:"min"`
}

//CtrlStage один интервал контроля
type CtrlStage struct {
	Nline  int       `json:"line"`   //Номер строки
	Start  StageTime `json:"start"`  //Время начала контроля
	End    StageTime `json:"end"`    //Время конца контроля
	TVPLen int       `json:"lenTVP"` //Длительность секунд контроля ТВП
	MGRLen int       `json:"lenMGR"` //Длительность секунд контроля МГР
}

//Compare сравнивание истина если равны
func (s *CtrlStage) Compare(ss *CtrlStage) bool {
	return reflect.DeepEqual(s, ss)
}

//NewSetCtrl создает новый набор контроля входов
func NewSetCtrl() *SetCtrl {
	r := new(SetCtrl)
	r.Stage = make([]CtrlStage, 8)
	for i := 0; i < len(r.Stage); i++ {
		r.Stage[i].Nline = i + 1
	}
	return r
}
func (s *CtrlStage) isEmpty() bool {
	if s.Start.Hour != 0 || s.Start.Minute != 0 {
		return false
	}
	if s.End.Hour != 0 || s.End.Minute != 0 {
		return false
	}
	if s.TVPLen != 0 || s.MGRLen != 0 {
		return false
	}
	return true
}

//IsEmpty вернет истину если весь набор пустой
func (sc *SetCtrl) IsEmpty() bool {
	for _, s := range sc.Stage {
		if !s.isEmpty() {
			return false
		}
	}
	return true
}

//FromBuffer заполнить из буфера
func (sc *SetCtrl) FromBuffer(buffer []int) error {
	if len(buffer) != 39 {
		return fmt.Errorf("неверная длина массива")
	}
	if buffer[2] != 24 || buffer[0] != 149 {
		return fmt.Errorf("несовпал номер массива на сервере и номер массива")
	}
	start := StageTime{Hour: 0, Minute: 0}
	pos := 5
	for i := 0; i < len(sc.Stage); i++ {
		sc.Stage[i].Start = start
		sc.Stage[i].TVPLen = buffer[pos]
		sc.Stage[i].MGRLen = buffer[pos+1]
		end := StageTime{Hour: buffer[pos+2], Minute: buffer[pos+3]}
		sc.Stage[i].End = end
		start = end
		if buffer[pos] == 0 && buffer[pos+1] == 0 && buffer[pos+2] == 0 && buffer[pos+3] == 0 {
			break
		}
		pos += 4
	}

	return nil
}

//ToBuffer выгружает набор в буфер
func (sc *SetCtrl) ToBuffer() []int {
	r := make([]int, 39)
	r[0] = 149
	r[2] = 24
	r[3] = 35
	r[4] = 1
	pos := 5
	for i := 0; i < len(sc.Stage); i++ {
		r[pos] = sc.Stage[i].TVPLen
		r[pos+1] = sc.Stage[i].MGRLen
		r[pos+2] = sc.Stage[i].End.Hour
		r[pos+3] = sc.Stage[i].End.Minute
		pos += 4
	}
	r[37] = 255
	r[38] = 255
	return r
}
