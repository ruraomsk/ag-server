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

//Данный пакет производит управление по характерным точкам
// Разбит на два раздела
// 	в первом разделе производится расчет характерной точки и выбор стратегии
// 	во втором разделе производится выполнение выбранной стратегии для каждого района и подрайона

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
	w := "select region,area,subarea,switch,pknow,pklast,xnum,strat,calc from public.xctrl;"
	rows, err := conDB.Query(w)
	if err != nil {
		logger.Error.Printf("Запрос  %s %s", w, err.Error())
		return err
	}
	flag := false
	for rows.Next() {
		var v StateSubArea
		var strat, calc []byte
		err = rows.Scan(&v.Region, &v.Area, &v.SubArea, &v.Switch, &v.PKNow, &v.PKLast, &v.XNumber, &strat, &calc)
		if err != nil {
			logger.Error.Printf("Запрос scan %s", err.Error())
			return err
		}

		err = json.Unmarshal(strat, &v.Strategys)
		if err != nil {
			logger.Error.Printf("json %s %s", string(strat), err.Error())
			return err
		}
		err = json.Unmarshal(calc, &v.Calculates)
		if err != nil {
			logger.Error.Printf("json %s %s", string(calc), err.Error())
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
			err = rows.Scan(&id, &state)
			if err != nil {
				logger.Error.Printf("Запрос scan %s", err.Error())
				return err
			}
			var c pudge.Cross
			err = json.Unmarshal(state, &c)
			if err != nil {
				logger.Error.Printf("json %s %s", string(calc), err.Error())
				return err
			}
			for _, p := range v.Strategys {
				if c.Arrays.SetDK.IsEmpty(1, p.PK) {
					flag = true
					logger.Error.Printf("Перекресток {%d %d %d} не имеет плана координации %d", v.Region, v.Area, id, p.PK)
				}
			}

		}
	}
	if flag {
		logger.Info.Print("Найдены ошибки посмотрите протокол")
		return fmt.Errorf("Найдены ошибки посмотрите протокол")
	}
	logger.Info.Print("Все проверено ошибок нет")
	return nil
}
