package logsys

import (
	"bytes"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/jasonlvhit/gocron"
	"github.com/ruraomsk/ag-server/logger"
	"github.com/ruraomsk/ag-server/setup"
)

var needHistory = `
	CREATE TABLE if not exists public.loghistory
	(
	    tm timestamp with time zone NOT NULL,
    	id integer NOT NULL,
    	crossinfo jsonb NOT NULL,
    	txt text COLLATE pg_catalog."default" NOT NULL,
    	region integer NOT NULL 
	)
	WITH (
		autovacuum_enabled = TRUE
	)
	TABLESPACE pg_default;
	
	ALTER TABLE public.loghistory
		OWNER to postgres;
`

func makeMoveLog() {
	dbinfo := fmt.Sprintf("host=%s user=%s password=%s dbname=%s sslmode=disable",
		setup.Set.DataBase.Host, setup.Set.DataBase.User,
		setup.Set.DataBase.Password, setup.Set.DataBase.DBname)
	var dblog *sql.DB
	var err error
	for {
		dblog, err = sql.Open("postgres", dbinfo)
		if err != nil {
			logger.Error.Printf("Запрос на открытие %s %s", dbinfo, err.Error())
			time.Sleep(10 * time.Second)
			continue
		}
		if err = dblog.Ping(); err != nil {
			logger.Error.Printf("Ping %s", err.Error())
			time.Sleep(time.Second * 10)
			continue
		}
		break
	}
	dbmov, err := sql.Open("postgres", dbinfo)
	if err != nil {
		logger.Error.Printf("Запрос на открытие %s %s", dbinfo, err.Error())
		return
	}
	_, _ = dbmov.Exec(needHistory)
	defer func() {
		dblog.Close()
		dbmov.Close()

	}()
	w := "select region,id,tm,crossinfo,txt,journal from public.logdevice;"
	log, err := dblog.Query(w)
	if err != nil {
		logger.Error.Printf("запрос %s %s", w, err.Error())
		return
	}
	last := time.Now().Add(-(time.Duration(setup.Set.LogSystem.Period*24) * time.Hour))
	del := make([]string, 0)
	var region, id int
	var tm, crossinfo, txt, journal []byte
	for log.Next() {
		log.Scan(&region, &id, &tm, &crossinfo, &txt, &journal)
		t, err := getTime(string(tm))
		if err != nil {
			logger.Error.Printf("разбор %s %s", tm, err.Error())
			return
		}
		if t.Before(last) {
			d := fmt.Sprintf("delete from public.logdevice where region=%d and id=%d and tm='%s' and crossinfo='%s';", region, id, string(tm), string(crossinfo))
			del = append(del, d)
			txt = bytes.Replace(txt, []byte("'"), []byte(" "), -1)
			if isWorkRegion(region) {
				if len(journal) != 0 {
					w = fmt.Sprintf("insert into public.loghistory (region,id,tm,crossinfo,txt,journal) values (%d,%d,'%s','%s','%s','%s')",
						region, id, string(tm), string(crossinfo), string(txt), string(journal))
				} else {
					w = fmt.Sprintf("insert into public.loghistory (region,id,tm,crossinfo,txt,journal) values (%d,%d,'%s','%s','%s','{}')",
						region, id, string(tm), string(crossinfo), string(txt))

				}
				_, err = dbmov.Exec(w)
				if err != nil {
					logger.Error.Printf("запрос %s %s", w, err.Error())
					return
				}
			}
		}
	}
	log.Close()
	for _, d := range del {
		_, err = dblog.Exec(d)
		if err != nil {
			logger.Error.Printf("запрос %s %s", w, err.Error())
			return
		}
	}
}
func getTime(tm string) (time.Time, error) {
	layout := "2006-01-02 15:04:05.999999999Z07:00"
	tm = strings.Replace(tm, "T", " ", 1)
	return time.Parse(layout, tm)
}
func Start() {
	logger.Info.Printf("Move log device starting..")
	makeMoveLog()
	_ = gocron.Every(1).Day().At(setup.Set.LogSystem.StartTime).Do(makeMoveLog)
	<-gocron.Start()
	logger.Info.Printf("Move log device working..")
}

func isWorkRegion(region int) bool {
	for _, v := range setup.Set.XCtrl.Regions {
		if v[0] == region {
			return true
		}
	}
	return false
}
