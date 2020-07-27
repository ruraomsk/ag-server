package xcontrol

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/JanFant/TLServer/logger"
	"github.com/ruraomsk/ag-server/pudge"
	"github.com/ruraomsk/ag-server/setup"

	//Инициализатор постргресса
	_ "github.com/lib/pq"
)

//Corrector Проверяем корректность системы к управлению
func Corrector() error {
	logger.Info.Printf("Проверяем систему ....")
	dbinfo := fmt.Sprintf("host=%s user=%s password=%s dbname=%s sslmode=disable",
		setup.Set.DataBase.Host, setup.Set.DataBase.User,
		setup.Set.DataBase.Password, setup.Set.DataBase.DBname)
	conDB, err := sql.Open("postgres", dbinfo)
	if err != nil {
		logger.Error.Printf("Запрос на открытие %s %s", dbinfo, err.Error())
		return err
	}
	defer conDB.Close()
	if err = conDB.Ping(); err != nil {
		logger.Error.Printf("Ping %s", err.Error())
		return err
	}
	w := "select state from public.xctrl;"
	rows, err := conDB.Query(w)
	if err != nil {
		logger.Error.Printf("Запрос  %s %s", w, err.Error())
		return err
	}
	flag := false
	for rows.Next() {
		var status []string
		var v State
		var stat []byte
		status = make([]string, 0)
		err = rows.Scan(&stat)
		if err != nil {
			logger.Error.Printf("Запрос scan %s", err.Error())
			return err
		}

		err = json.Unmarshal(stat, &v)
		if err != nil {
			logger.Error.Printf("json %s %s", string(stat), err.Error())
			return err
		}
		w = fmt.Sprintf("select id,state from public.cross where region = %d and area=%d and subarea = %d;", v.Region, v.Area, v.SubArea)
		cross, err := conDB.Query(w)
		if err != nil {
			logger.Error.Printf("Запрос  %s %s", w, err.Error())
			return err
		}
		for cross.Next() {
			var id int
			var state []byte
			err = cross.Scan(&id, &state)
			if err != nil {
				logger.Error.Printf("Запрос scan %s", err.Error())
				return err
			}
			var c pudge.Cross
			err = json.Unmarshal(state, &c)
			if err != nil {
				logger.Error.Printf("json %s %s", string(state), err.Error())
				return err
			}
			for _, p := range v.Strategys {
				if p.PK == 0 {
					continue
				}
				if c.Arrays.SetDK.IsEmpty(1, p.PK) {
					flag = true
					s := fmt.Sprintf("Перекресток {%d %d %d} не имеет плана координации %d", v.Region, v.Area, id, p.PK)
					//logger.Error.Printf(s)
					status = append(status, s)
				}
			}
		}
		v.Status = status
		s, err := json.Marshal(&v)
		w = fmt.Sprintf("update public.xctrl set state='%s' where region=%d and area=%d and subarea=%d;",
			string(s), v.Region, v.Area, v.SubArea)
		_, err = conDB.Exec(w)
		if err != nil {
			logger.Error.Printf("Запрос %s %s", w, err.Error())
			return err
		}

	}
	if flag {
		//logger.Info.Print("найдены ошибки посмотрите протокол")
		return fmt.Errorf("найдены ошибки посмотрите протокол")
	}
	logger.Info.Print("Все проверено ошибок нет")
	return nil
}
