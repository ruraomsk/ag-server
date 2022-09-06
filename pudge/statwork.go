package pudge

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/ruraomsk/ag-server/extcon"
	"github.com/ruraomsk/ag-server/logger"
	"github.com/ruraomsk/ag-server/setup"
)

type RecordStat struct {
	Region Region
	Stat   Statistic
}

var StatisticChan chan RecordStat

func StatisticStart() {
	StatisticChan = make(chan RecordStat, 1000)
	dbinfo := fmt.Sprintf("host=%s user=%s password=%s dbname=%s sslmode=disable",
		setup.Set.DataBase.Host, setup.Set.DataBase.User,
		setup.Set.DataBase.Password, setup.Set.DataBase.DBname)
	con, err := sql.Open("postgres", dbinfo)
	if err != nil {
		logger.Error.Printf("Запрос на открытие %s %s", dbinfo, err.Error())
		return
	}
	defer con.Close()
	context, _ := extcon.NewContext("writerStat")
	timer := extcon.SetTimerClock(time.Duration(1 * time.Minute))
	for {
		select {
		case rs := <-StatisticChan:
			w := fmt.Sprintf("select stat from public.statistics where date='%s' and region=%d and area=%d and id=%d;",
				rs.Region.LocalTime().Format("2006-01-02"), rs.Region.Region, rs.Region.Area, rs.Region.ID)
			var state ArchStat
			rows, _ := con.Query(w)
			needInsert := true
			for rows.Next() {
				var buf []byte
				rows.Scan(&buf)
				json.Unmarshal(buf, &state)
				needInsert = false
				state.Statistics = append(state.Statistics, rs.Stat)
			}
			rows.Close()
			if needInsert {
				state.Region = rs.Region.Region
				state.Area = rs.Region.Area
				state.ID = rs.Region.ID
				state.Date = rs.Region.LocalTime()
				state.Statistics = make([]Statistic, 0)
				state.Statistics = append(state.Statistics, rs.Stat)
				js, _ := json.Marshal(&state)
				w = fmt.Sprintf("INSERT INTO public.statistics(region, area, id, date, stat) VALUES (%d, %d, %d, '%s', '%s');",
					state.Region, state.Area, state.ID, state.Date.Format("2006-01-02"), string(js))
			} else {
				state.Date = rs.Region.LocalTime()
				js, _ := json.Marshal(state)
				w = fmt.Sprintf("Update public.statistics set stat='%s' where date='%s' and region=%d and area=%d and id=%d;",
					string(js), state.Date.Format("2006-01-02"), state.Region, state.Area, state.ID)

			}
			_, err = con.Exec(w)
			if err != nil {
				logger.Error.Printf("%s %s", w, err.Error())
			}

		case <-context.Done():
			logger.Info.Print("writerArch is stoped...")
			return
		case <-timer.C:
			//Пинганем БД чтобы соединение не закрылось
			if err = con.Ping(); err != nil {
				logger.Error.Printf("Ping %s", err.Error())
				return
			}

		}
	}
}
