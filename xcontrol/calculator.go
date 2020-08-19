package xcontrol

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/JanFant/TLServer/logger"
	"github.com/ruraomsk/ag-server/setup"
	"math/rand"
	"sort"
	"time"
)

func (s *State) calculate() {
	s.Results = make([]Result, 0)
	for _ = range s.Calculates {
		r := new(Result)
		r.Ileft, r.Iright = rand.Intn(900), rand.Intn(900)
		s.Results = append(s.Results, *r)
	}
	if len(s.Results) < len(s.Calculates)/2 {
		//Точек получено в два раза меньше чем задано
		s.PKCalc = 0
		return
	}
	columns := make([]int, 3)
	for _, r := range s.Results {
		d := float64(r.Ileft) / float64(r.Iright)
		if d < s.LeftRel {
			columns[0] = columns[0] + 1
		}
		if d <= s.RightRel && d >= s.LeftRel {
			columns[1] = columns[1] + 1
		}
		if d > s.RightRel {
			columns[2] = columns[2] + 1
		}
	}
	col := 1
	if columns[0] > columns[1] && columns[0] > columns[2] {
		col = 0
	}
	if columns[2] > columns[1] && columns[2] > columns[0] {
		col = 2
	}
	max := 0
	for _, r := range s.Results {
		switch col {
		case 0:
			if r.Ileft > max {
				max = r.Ileft
			}
		case 1:
			if r.Ileft > max {
				max = r.Ileft
			}
			if r.Iright > max {
				max = r.Iright
			}
		case 2:
			if r.Iright > max {
				max = r.Iright
			}
		}
	}
	sort.Slice(s.Strategys, func(i, j int) bool { return s.Strategys[i].XLeft < s.Strategys[j].XLeft })
	for _, st := range s.Strategys {
		s.LastTime = time.Now()
		s.PKCalc = -1
		if max >= st.XLeft && max < st.XRight {
			switch col {
			case 0:
				s.PKCalc = st.PKL
			case 1:
				s.PKCalc = st.PKS
			case 2:
				s.PKCalc = st.PKR
			}
		}
	}
	if s.PKCalc < 0 {
		//Не нашли берем последний известный
		st := s.Strategys[len(s.Strategys)-1]
		switch col {
		case 0:
			s.PKCalc = st.PKL
		case 1:
			s.PKCalc = st.PKS
		case 2:
			s.PKCalc = st.PKR
		}
	}
}
func (s *State) change() {
	s.PKNow = s.PKCalc
}

//Calculator Посылает новые планы координации на устройства
func Calculator() {
	dbinfo := fmt.Sprintf("host=%s user=%s password=%s dbname=%s sslmode=disable",
		setup.Set.DataBase.Host, setup.Set.DataBase.User,
		setup.Set.DataBase.Password, setup.Set.DataBase.DBname)
	conDB, err := sql.Open("postgres", dbinfo)
	if err != nil {
		logger.Error.Printf("Запрос на открытие %s %s", dbinfo, err.Error())
		return
	}
	defer conDB.Close()
	if err = conDB.Ping(); err != nil {
		logger.Error.Printf("Ping %s", err.Error())
		return
	}
	//Обновим все записи установим счетчик
	defer conDB.Exec("rollback;")
	w := "select state from public.xctrl;"
	rows, err := conDB.Query(w)
	if err != nil {
		logger.Error.Printf("Запрос  %s %s", w, err.Error())
		return
	}
	for rows.Next() {
		var v State
		var vv []byte
		err = rows.Scan(&vv)
		if err != nil {
			logger.Error.Printf("Запрос scan %s", err.Error())
			return
		}
		err = json.Unmarshal(vv, &v)
		if err != nil {
			logger.Error.Printf("Запрос unmurhal %v %s", vv, err.Error())
			return
		}
		v.Remain = v.Step
		v.PKCalc = 0
		v.PKNow = 0
		v.PKLast = 0
		v.Results = make([]Result, 0)
		s, err := json.Marshal(&v)
		if err != nil {
			logger.Error.Printf("Запрос marhal %v %s", vv, err.Error())
			return
		}
		w = fmt.Sprintf("update public.xctrl set state='%s' where region=%d and area=%d and subarea=%d;",
			string(s), v.Region, v.Area, v.SubArea)
		_, err = conDB.Exec(w)
		if err != nil {
			logger.Error.Printf("Запрос %s %s", w, err.Error())
			return
		}
	}
	for true {
		time.Sleep(1 * time.Minute)
		//logger.Info.Printf("Управление по характерным точка стадия 1....")
		_, err = conDB.Exec("begin;")
		if err != nil {
			logger.Error.Printf("Запрос begin %s", err.Error())
			return
		}
		Corrector()
		w := "select state from public.xctrl;"
		rows, err := conDB.Query(w)
		if err != nil {
			logger.Error.Printf("Запрос  %s %s", w, err.Error())
			return
		}
		for rows.Next() {
			var v State
			var vv []byte
			err = rows.Scan(&vv)
			if err != nil {
				logger.Error.Printf("Запрос scan %s", err.Error())
				return
			}
			err = json.Unmarshal(vv, &v)
			if err != nil {
				logger.Error.Printf("Запрос unmurhal %v %s", vv, err.Error())
				return
			}
			if !v.Switch && v.PKNow == 0 {
				continue
			}
			v.Remain--
			if v.Remain <= 0 {
				v.LastTime = time.Now()
				if !v.Switch {
					v.PKNow = 0
				} else {
					//Собственно расчет
					v.calculate()
					v.change()
				}
				v.Remain = v.Step
			}
			s, err := json.Marshal(&v)
			if err != nil {
				logger.Error.Printf("Запрос marhal %v %s", vv, err.Error())
				return
			}
			w = fmt.Sprintf("update public.xctrl set state='%s' where region=%d and area=%d and subarea=%d;",
				string(s), v.Region, v.Area, v.SubArea)
			_, err = conDB.Exec(w)
			if err != nil {
				logger.Error.Printf("Запрос %s %s", w, err.Error())
				return
			}

		}
		_, err = conDB.Exec("commit;")
		if err != nil {
			logger.Error.Printf("Запрос commit %s", err.Error())
			return
		}
	}
	return
}
