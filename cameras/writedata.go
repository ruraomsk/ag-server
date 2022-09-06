package cameras

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/ruraomsk/ag-server/logger"
	"github.com/ruraomsk/ag-server/pudge"
)

func writeToStatistics(con *sql.DB, w ExtStatistic) {
	var st pudge.ArchStat
	//var dv pudge.Controller
	var s []byte

	q := fmt.Sprintf("select stat from public.statistics where region=%d and area=%d and id=%d and date='%s';",
		w.Region, w.Area, w.ID, w.Date.Format("2006-01-02"))
	rows, err := con.Query(q)
	if err != nil {
		logger.Error.Printf("Запрос %s %s", q, err.Error())
		return
	}
	if rows.Next() {
		rows.Scan(&s)
		err = json.Unmarshal(s, &st)
		if err != nil {
			logger.Error.Printf("Разбор %s %s", string(s), err.Error())
			rows.Close()
			return
		}
		for _, as := range st.Statistics {
			if as.Hour == w.Statistic.Hour && as.Min == w.Statistic.Min {
				//Такая запись уже есть
				logger.Error.Printf("%s %d-%d-%d %d:%d такая запись уже есть", w.Date.Format("2006-01-02"), w.Region, w.Area, w.ID, as.Hour, as.Min)
				rows.Close()
				return
			}
		}
		st.Statistics = append(st.Statistics, w.Statistic)
		s, err = json.Marshal(&st)
		if err != nil {
			logger.Error.Printf("Marshal %v %s", st, err.Error())
			rows.Close()
			return
		}
		q := fmt.Sprintf("update public.statistics set stat='%s' where region=%d and area=%d and id=%d and date='%s';", string(s),
			w.Region, w.Area, w.ID, w.Date.Format("2006-01-02"))
		//logger.Info.Println(q)
		_, err = con.Query(q)
		if err != nil {
			logger.Error.Printf("Запрос %s %s", q, err.Error())
		}
		rows.Close()
	}
	rows.Close()
	st.ID = w.ID
	st.Area = w.Area
	st.Region = w.Region
	st.Date = w.Date
	st.Statistics = make([]pudge.Statistic, 0)
	st.Statistics = append(st.Statistics, w.Statistic)
	s, _ = json.Marshal(&st)
	q = fmt.Sprintf("insert into public.statistics (stat,region,area,id,date) values ('%s',%d,%d,%d,'%s');", string(s),
		w.Region, w.Area, w.ID, w.Date.Format("2006-01-02"))
	//logger.Info.Println(q)
	_, err = con.Query(q)
	if err != nil {
		logger.Error.Printf("Запрос %s %s", q, err.Error())
	}
}
func writeToDevices(con *sql.DB, w ExtStatistic, id int) {
	// var dv pudge.Controller
	// var s []byte
	//Значит пишем в БД
	// f := fmt.Sprintf("select device from public.devices where id=%d ;", id)
	// fs, err := con.Query(f)
	// if err != nil {
	// 	logger.Error.Printf("Запрос %s %s", f, err.Error())
	// 	return
	// }
	// if fs.Next() {
	// 	fs.Scan(&s)
	// 	err = json.Unmarshal(s, &dv)
	// 	if err != nil {
	// 		logger.Error.Printf("Разбор %s %s", string(s), err.Error())
	// 		fs.Close()
	// 		return
	// 	}
	// 	for _, as := range dv.Statistics {
	// 		if as.Hour == w.Statistic.Hour && as.Min == w.Statistic.Min {
	// 			logger.Error.Printf("%d-%d-%d %d:%d такая запись уже есть", w.Region, w.Area, w.ID, as.Hour, as.Min)
	// 			fs.Close()
	// 			return
	// 		}
	// 	}
	// 	dv.Statistics = append(dv.Statistics, w.Statistic)
	// 	s, err = json.Marshal(&dv)
	// 	if err != nil {
	// 		logger.Error.Printf("Marshal %v %s", dv, err.Error())
	// 		fs.Close()
	// 		return
	// 	}
	// 	q := fmt.Sprintf("update public.devices set device='%s' where id=%d;", string(s), id)
	// 	_, err = con.Exec(q)
	// 	if err != nil {
	// 		logger.Error.Printf("Запрос %s %s", q, err.Error())
	// 	}
	// }
	// fs.Close()
}

func writedata() {
	// dbinfo := fmt.Sprintf("host=%s user=%s password=%s dbname=%s sslmode=disable",
	// 	setup.Set.DataBase.Host, setup.Set.DataBase.User,
	// 	setup.Set.DataBase.Password, setup.Set.DataBase.DBname)
	// con, err := sql.Open("postgres", dbinfo)
	// if err != nil {
	// 	logger.Error.Printf("Запрос на открытие %s %s", dbinfo, err.Error())
	// 	return
	// }
	// defer con.Close()
	// if err = con.Ping(); err != nil {
	// 	logger.Error.Printf("Ping %s", err.Error())
	// 	return
	// }
	// for {
	// 	w := <-writer
	// 	cross, is := pudge.GetCross(w.Region, w.Area, w.ID)
	// 	if !is {
	// 		writeToStatistics(con, w)
	// 		continue
	// 	}
	// dv, is := pudge.GetController(cross.IDevice)
	// if !is {
	// 	writeToDevices(con, w, cross.IDevice)
	// 	continue
	// }
	// found := false
	// for _, as := range dv.Statistics {
	// 	if as.Hour == w.Statistic.Hour && as.Min == w.Statistic.Min {
	// 		logger.Error.Printf("%d-%d-%d %d:%d такая запись уже есть", w.Region, w.Area, w.ID, as.Hour, as.Min)
	// 		found = true
	// 		break
	// 	}
	// }
	// if !found {
	// 	dv.Statistics = append(dv.Statistics, w.Statistic)
	// 	pudge.SetController(dv)
	// }
	// }
}
