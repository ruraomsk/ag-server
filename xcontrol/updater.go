package xcontrol

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/JanFant/TLServer/logger"
	"github.com/ruraomsk/ag-server/setup"

	//Инициализатор постргресса
	_ "github.com/lib/pq"
)

//Updater заполняем отладочными данными таблицу
func Updater() {
	logger.Info.Printf("Добавляем данные ....")
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
	subs := make(map[key]State)
	w := "select region,area,subarea,id from public.cross;"
	rows, err := conDB.Query(w)

	var id int

	if err != nil {
		logger.Error.Printf("Запрос  %s %s", w, err.Error())
		return
	}
	for rows.Next() {
		v := newState()
		err = rows.Scan(&v.Region, &v.Area, &v.SubArea, &id)
		if err != nil {
			logger.Error.Printf("Запрос чтения %s %s", w, err.Error())
			return
		}

		k := key{Region: v.Region, Area: v.Area, SubArea: v.SubArea}
		_, is := subs[k]
		if !is {
			for i := 0; i < len(v.Calculates); i++ {
				v.Calculates[i].Area = v.Area
				v.Calculates[i].Region = v.Region
				v.Calculates[i].ID = id
			}
			subs[k] = v
			// fmt.Printf("%v\n", v)
		}
	}
	w = "delete from public.xctrl;"
	_, err = conDB.Exec(w)
	if err != nil {
		logger.Error.Printf("Запрос  %s %s", w, err.Error())
		return
	}
	for _, vv := range subs {
		// fmt.Printf("==%v\n", vv)
		state, err := json.Marshal(vv)
		if err != nil {
			logger.Error.Printf("json  %s", err.Error())
			return
		}

		w = fmt.Sprintf("insert into public.xctrl (region,area,subarea,state) values (%d,%d,%d,'%s');",
			vv.Region, vv.Area, vv.SubArea, string(state))
		_, err = conDB.Exec(w)
		if err != nil {
			logger.Error.Printf("Запрос  %s %s", w, err.Error())
			return
		}
		// fmt.Println(w)
	}
	logger.Info.Printf("Звершено добавление данных ....")
}
func newState() State {
	v := new(State)

	v.LastTime = time.Now()
	v.PKLast = 0
	v.PKNow = 0
	v.Switch = false
	v.XNumber = 0

	v.Strategys = make([]Strategy, 0)
	v.Calculates = make([]Calc, 0)
	v.Status = make([]string, 0)

	v.Strategys = append(v.Strategys, Strategy{XLeft: 0, XRight: 100, PK: 1})
	v.Strategys = append(v.Strategys, Strategy{XLeft: 100, XRight: 200, PK: 2})
	v.Strategys = append(v.Strategys, Strategy{XLeft: 200, XRight: 300, PK: 3})
	v.Strategys = append(v.Strategys, Strategy{XLeft: 300, XRight: 99999, PK: 4})
	v.Calculates = append(v.Calculates, Calc{Chanal: 1, Mult: 1})
	v.Calculates = append(v.Calculates, Calc{Chanal: 2, Mult: 1})
	v.Calculates = append(v.Calculates, Calc{Chanal: 3, Mult: 1})
	return *v
}
