package dumper

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/ruraomsk/TLServer/logger"
	"github.com/ruraomsk/ag-server/extcon"
	"github.com/ruraomsk/ag-server/pudge"
	"github.com/ruraomsk/ag-server/setup"
	"time"
)

var writeArch chan pudge.ArchStat

func writerArch() {
	writeArch = make(chan pudge.ArchStat, 1000)
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
	context, _ := extcon.NewContext("writerArch")
	timer := extcon.SetTimerClock(time.Duration(1 * time.Minute))
	for {
		select {
		case arch := <-writeArch:
			js, _ := json.Marshal(arch)
			if len(arch.Statistics) != 0 {
				w := fmt.Sprintf("INSERT INTO public.statistics(region, area, id, date, stat) VALUES (%d, %d, %d, '%s', '%s');",
					arch.Region, arch.Area, arch.ID, arch.Date.Format("2006-01-02"), string(js))
				_, err = conDB.Exec(w)
				if err != nil {
					logger.Error.Printf("%s %s", w, err.Error())
				}
			}
		case <-context.Done():
			logger.Info.Print("writerArch is stoped...")
			return
		case <-timer.C:
			//Пинганем БД чтобы соединение не закрылось
			if err = conDB.Ping(); err != nil {
				logger.Error.Printf("Ping %s", err.Error())
				return
			}

		}
	}
}
