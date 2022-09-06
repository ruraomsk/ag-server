package binding

import (
	"fmt"
	"reflect"
	"sort"

	"github.com/ruraomsk/ag-server/logger"
)

//Планы координации

//SetDK наборы планов координации для обеих ДК перекрестка
type SetDK struct {
	DK []SetPk `json:"dk"` // Наборы для ДК1
}

//Compare сравнивание истина если равны
func (sd *SetDK) Compare(ss *SetDK) bool {
	return reflect.DeepEqual(sd, ss)
}
func (sd *SetDK) CorrectPk() {
	//for i := 0; i < 12; i++ {
	//	sd.DK[i].Pk=i+1
	//}
}

//GetPhases возращает список доступных фаз
func (sd *SetDK) GetPhases(pk int) []int {
	res := make([]int, 0)
	m := make(map[int]int)
	for _, sp := range sd.DK {
		if sp.Pk != pk {
			continue
		}
		for _, st := range sp.Stages {
			if st.Number != 0 {
				if st.Number < 9 {
					_, is := m[st.Number]
					if !is {
						m[st.Number] = st.Number
					}
				}
			}
		}
		break
	}
	var i int
	for i = range m {
		res = append(res, i)
	}
	return res
}

//SetPk набор планов координации перекрестка
type SetPk struct {
	DK          int    `json:"dk"`     //Номер ДК
	Pk          int    `json:"pk"`     //Номер программы от 1 до 12
	Description string `json:"desc"`   //Описание плана координации
	TypePU      int    `json:"tpu"`    //Тип программы управления управления 0-ЛПУ (локальная) 1-ПК(координации)
	RazLen      bool   `json:"razlen"` //Признак наличия разнодлительных фаз
	Tc          int    `json:"tc"`     //Время цикла программы
	Shift       int    `json:"shift"`  //Сдвиг начала цикла
	//LastType    int     `json:"lasttype"`   //Тип переходной фазы при сдвиге
	//LastNumber  int     `json:"lastnumber"` //Номер переходной фазы при сдвиге
	TwoT   bool    `json:"twot"` //Признак 2Т
	Stages []Stage `json:"sts"`  //Фазы переключения
}

//Stage описание одной фазы плана координации
type Stage struct {
	Nline  int `json:"line"`  //Номер строки
	Start  int `json:"start"` //Время начала фазы
	Number int `json:"num"`   //Номер фазы
	Tf     int `json:"tf"`    //Тип фазы 0 -простая
	// 1 - МГР
	// 2 - 1ТВП
	// 3 - 2ТВП
	// 4 - 1,2ТВП
	// 5 - Зам 1 ТВП
	// 6 - Зам 2 ТВП
	// 7 - Зам
	// 8 - МДК
	// 9 - ВДК
	Stop int  `json:"stop"` //Завершение фазы
	Plus bool `json:"plus"` //Признак переноса времени на следующую фазу
	Trs  bool `json:"trs"`  //Признак перехода через цикл
	Dt   int  `json:"dt"`   //Сколтко времени работать в начале цикла
}

//NewSetDK создание нового набора планов координации
func NewSetDK() *SetDK {
	r := new(SetDK)
	r.DK = make([]SetPk, 12)
	for n := range r.DK {
		r.DK[n] = NewSetPk(n + 1)
	}
	return r
}

//NewSetPk новый план
func NewSetPk(pk int) SetPk {
	r := new(SetPk)
	r.DK = 1
	r.Pk = pk
	r.Description = fmt.Sprintf("План координации %d", pk)
	r.Stages = make([]Stage, 12)
	for n := range r.Stages {
		r.Stages[n].Nline = n + 1
	}
	return *r
}

//IsEmpty если план координации для данного ДК нулеыой то истина
func (sd *SetDK) IsEmpty(dk, pk int) bool {
	return false
}

//ToBuffer выгружает в буфер
func (st *SetPk) ToBuffer() []int {
	sts := make([]Stage, 0)
	for _, s := range st.Stages {
		if s.Number == 0 && s.Tf == 0 && s.Stop == 0 {
			continue
		}
		sts = append(sts, s)
	}
	sort.Slice(sts, func(i int, j int) bool {
		if sts[i].Start != sts[j].Start {
			return sts[i].Start < sts[j].Start
		} else {
			return sts[i].Tf < sts[j].Tf
		}
	})
	r := make([]int, 34)
	r[0] = st.Pk + 99
	if st.DK == 2 {
		r[0] = st.Pk + 119
	}
	r[2] = 133
	r[3] = 30
	r[4] = st.Pk
	if st.DK == 2 {
		r[4] += 128
	}
	//if st.Tc == 0 {
	//	logger.Error.Printf("Время цикла =0 %v", st)
	//	st.Tc = 30
	//}
	r[5] = st.Tc
	if st.Tc < 3 {
		return r
	}
	r[6] = 256 % st.Tc
	r[7] = (256 * 256) % st.Tc
	l := 0
	mgr := false
	mdk := false
	plus := false
	for _, s := range sts {
		if s.Number == 0 && s.Tf == 0 && s.Stop == 0 {
			break
		}
		if s.Plus {
			plus = true
		}
		if s.Tf == 1 {
			mgr = true
		}
		if s.Tf == 8 {
			mdk = true
		}

		l++
	}
	count := 0
	for _, s := range sts {
		if s.Trs {
			count++
		}
	}
	r[8] = 192
	if st.Shift != 0 && sts[0].Start != 0 {
		r[9] += 16 //Есть переход фаз
	}
	if st.TypePU == 1 {
		r[9] += 128 //Есть ЛПУ
	}
	if mgr {
		r[9] += 64 //Среди фаз есть МГР
	}
	if st.RazLen {
		r[9] += 32
	}
	if st.TwoT {
		r[9] += 2
	}
	if plus {
		r[9] += 8
	}
	if mdk && !st.RazLen {
		r[9] = 0
	}
	pos := 10
	tvpb := make([]Stage, 0)
	tvpflag := false
	for _, s := range st.Stages {
		if s.Tf == 2 || s.Tf == 4 {
			tvpb = make([]Stage, 0)
			if s.Trs {
				tvpflag = true
			}
			tvpb = append(tvpb, s)
			continue
		}
		if s.Tf == 5 || s.Tf == 6 || s.Tf == 7 {
			if s.Trs {
				tvpflag = true
			}
			tvpb = append(tvpb, s)
			continue
		}
	}
	if tvpflag {
		//Значит есть переход замещающих фаз
		for _, s := range st.Stages {
			for i, t := range tvpb {
				if s.Nline == t.Nline && s.Trs {
					t.Dt = s.Dt
					tvpb[i] = t
				}
			}
		}
		for _, s := range tvpb {
			r[pos] = s.Number
			if s.Tf == 2 {
				r[pos] += 160 // 2 - 1ТВП
			}
			if s.Tf == 3 {
				r[pos] += 96 // 3 - 2ТВП
			}
			if s.Tf == 4 {
				r[pos] += 224 // 4 - 1,2ТВП
			}
			if s.Tf == 5 {
				r[pos] += 128 // 5 - Зам 1 ТВП
			}
			if s.Tf == 6 {
				r[pos] += 64 //  6 - Зам 2 ТВП
			}
			if s.Tf == 7 {
				r[pos] += 16 //  7 - Зам
			}
			pos++
			r[pos] = s.Dt
			pos++
			r[8] += 1
		}

	} else {
		if st.Shift != 0 && sts[0].Start != 0 {
			//Есть сдвиг формируем запись сдвига
			//Находим переносные фазы
			for _, s := range sts {
				if !s.Trs {
					continue
				}
				r[pos] = s.Number
				pos++
				r[pos] = sts[0].Start
				pos++
				r[8] += 1
			}
		}
	}

	for _, s := range sts {
		if s.Number == 0 && s.Tf == 0 && s.Stop == 0 {
			break
		}

		r[pos] = s.Number
		if s.Tf == 2 {
			r[pos] += 160 // 2 - 1ТВП
		}
		if s.Tf == 3 {
			r[pos] += 96 // 3 - 2ТВП
		}
		if s.Tf == 4 {
			r[pos] += 224 // 4 - 1,2ТВП
		}
		if s.Tf == 5 {
			r[pos] += 128 // 5 - Зам 1 ТВП
		}
		if s.Tf == 6 {
			r[pos] += 64 //  6 - Зам 2 ТВП
		}
		if s.Tf == 7 {
			r[pos] += 16 //  7 - Зам
		}
		if s.Tf == 8 {
			if r[9] != 0 {
				r[pos] += 45 //МДК
			}
		}
		if s.Tf == 9 {
			r[pos] += 32 // 9 - ВДК
		}
		pos++
		for pos >= len(r) {
			//logger.Debug.Printf("Массив %v",st)
			pos--
		}
		r[pos] = s.Stop
		pos++
		r[8] += 1
		for pos >= len(r) {
			//logger.Debug.Printf("Массив %v",st)
			pos--
		}
	}
	return r
}

//FromBuffer создает план координации из буфера возвращает ошибку
func (sd *SetDK) FromBuffer(buffer []int) error {
	st := NewSetPk(1)
	err := st.FromBuffer(buffer)
	if err != nil {
		return err
	}
	if buffer[0] <= 112 {
		st.DK = 1
		sd.DK[st.Pk-1] = st
	}
	for i, d := range sd.DK {
		d.Description = fmt.Sprintf("План координации %d", d.Pk)
		sd.DK[i] = d
	}
	return nil
}

//FromBuffer создает план координации из буфера возвращает ошибку
func (st *SetPk) FromBuffer(buffer []int) error {
	if len(buffer) != 34 {
		return fmt.Errorf("неверная длина массива")
	}
	if buffer[2] != 133 {
		return fmt.Errorf("несовпал номер массива на сервере и номер массива")
	}
	if buffer[0] < 100 || buffer[0] > 111 {
		return fmt.Errorf("неверный номер массива %d", buffer[0])
	}
	// mdk:=false
	st.Pk = buffer[4] & 0x7f
	if buffer[4]&0x80 != 0 {
		st.DK = 2
	} else {
		st.DK = 1
	}
	st.Tc = buffer[5]
	Shift := false
	ShiftZerro := false
	if buffer[9]&16 != 0 {
		//есть переход фаз нужно читать сдвиг
		Shift = true
	}
	if buffer[9]&128 != 0 {
		st.TypePU = 1
	}
	mgr := false
	if buffer[9]&64 != 0 {
		mgr = true //Среди фаз есть МГР
	}
	if buffer[9]&32 != 0 {
		st.RazLen = true
	} else {
		st.RazLen = false
	}
	if buffer[9]&2 != 0 {
		st.TwoT = true
	} else {
		st.TwoT = false
	}
	// if buffer[9]==0 {
	// 	mdk=true
	// }
	plus := false
	if buffer[9]&8 != 0 {
		plus = true
	}
	pos := 10
	ss := make([]Stage, 12)
	for n := range ss {
		ss[n].Nline = n + 1
	}
	first := true
	for n := range ss {
		if buffer[pos] == 0 && buffer[pos+1] == 0 {
			break
		}
		ss[n].Number = buffer[pos] & 15
		m := buffer[pos] & 0xf0
		if m == 160 {
			ss[n].Tf = 2 // 2 - 1ТВП
		}
		if m == 96 {
			ss[n].Tf = 3 // 3 - 2ТВП
		}
		if m == 224 {
			ss[n].Tf = 4 // 4 - 1,2ТВП
			if plus {
				ss[n].Plus = true
			}
		}
		if m == 128 {
			ss[n].Tf = 5 // 5 - Зам 1 ТВП
		}
		if m == 64 {
			ss[n].Tf = 6 // 6 - Зам 2 ТВП
		}
		if m == 16 {
			ss[n].Tf = 7 //  7 - Зам
		}
		if m == 32 {
			ss[n].Tf = 9 //9 ВДК
		}
		if ss[n].Number == 0 {
			if (buffer[9] == 0 && first) || (buffer[9] == 32 && first) {
				ss[n].Tf = 8
				first = false
			} else {
				if buffer[9] == 32 && !first {
					ss[n].Tf = 9
				} else {
					if !mgr {
						logger.Error.Printf("есть фаза ноль но нет признака мгр ")
					}
					ss[n].Tf = 1
				}
			}
		}
		pos++
		ss[n].Stop = buffer[pos]
		pos++
	}
	//Пошли сплошные и мать его костыли пппппп
	st.Shift = 0
	si := 0
	if Shift {

		for i := 1; i < len(ss); i++ {
			if ss[i].Number == 1 {
				st.Shift = ss[i-1].Stop
				si = i
				break
			}
		}
	} else {
		if ss[0].Number != 1 {
			ShiftZerro = true
			for i := 1; i < len(ss); i++ {
				if ss[i].Number == 1 {
					st.Shift = ss[i-1].Stop
					si = i
					break
				}
			}
		}
	}
	//Перекатываем в Stage
	start := st.Shift
	j := 0
	for i := si; i < len(ss); i++ {
		st.Stages[j].Nline = j + 1
		st.Stages[j].Start = start
		st.Stages[j].Number = ss[i].Number
		st.Stages[j].Tf = ss[i].Tf
		st.Stages[j].Stop = ss[i].Stop
		start = ss[i].Stop
		if st.Stages[j].Tf == 0 && st.Stages[j].Stop == 0 && st.Stages[j].Number == 0 {
			st.Stages[j].Start = 0
			if Shift {
				st.Stages[j-1].Trs = true
				st.Stages[j-1].Dt = ss[0].Stop
			}
			break
		}
		if st.Stages[j].Tf == 5 || st.Stages[j].Tf == 6 || st.Stages[j].Tf == 7 {
			if j > 0 {
				st.Stages[j].Start = st.Stages[j-1].Start
			} else {
				st.Stages[j].Start = start
			}
		}
		j++
	}
	if Shift || ShiftZerro {
		sit := 1
		if ShiftZerro {
			sit = 0
			start = 0
		} else {
			start = ss[0].Stop
		}
		for i := sit; i < si; i++ {
			st.Stages[j].Nline = j + 1
			st.Stages[j].Start = start
			st.Stages[j].Number = ss[i].Number
			st.Stages[j].Tf = ss[i].Tf
			st.Stages[j].Stop = ss[i].Stop
			start = ss[i].Stop
			if st.Stages[j].Tf == 0 && st.Stages[j].Stop == 0 && st.Stages[j].Number == 0 {
				st.Stages[j].Start = 0
			}
			if st.Stages[j].Tf == 5 || st.Stages[j].Tf == 6 || st.Stages[j].Tf == 7 {
				if j > 0 {
					st.Stages[j].Start = st.Stages[j-1].Start
				} else {
					st.Stages[j].Start = start
				}
			}
			j++
		}
	}
	return nil
}
