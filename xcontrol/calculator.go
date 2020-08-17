package xcontrol

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/JanFant/TLServer/logger"
	"github.com/ruraomsk/ag-server/setup"
	"math/rand"
	"time"
)

func (s *State) calculate() {
	s.XNumber = rand.Intn(500)
}
func (s *State) change() {
	for _, st := range s.Strategys {
		if s.XNumber >= st.XLeft && s.XNumber < st.XRight {
			s.PKNow = st.PK
			s.LastTime = time.Now()
		}
	}
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
		v.PKNow = 0
		v.PKLast = 0
		v.XNumber = 0
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
		_, err = conDB.Exec("commit;")
		if err != nil {
			logger.Error.Printf("Запрос commit %s", err.Error())
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
