package techComm

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/lib/pq"
	"github.com/ruraomsk/TLServer/logger"
	"github.com/ruraomsk/ag-server/memDB"
	"github.com/ruraomsk/ag-server/pudge"
	"github.com/ruraomsk/ag-server/setup"
	"strconv"
	"strings"
)

var mapMessages map[string]map[int]string
var ChanLog chan pudge.RecLogCtrl
var db *sql.DB
var err error

// WriterLog Ведем простое логирование
func WriterLog() {
	mapMessages = make(map[string]map[int]string)
	ChanLog = make(chan pudge.RecLogCtrl, 1000)
	info := fmt.Sprintf("host=%s user=%s password=%s dbname=%s sslmode=disable",
		setup.Set.DataBase.Host, setup.Set.DataBase.User,
		setup.Set.DataBase.Password, setup.Set.DataBase.DBname)
	db, err = sql.Open("postgres", info)
	if err != nil {
		logger.Error.Printf("запрос на открытие %s %s", info, err.Error())
		return
	}

	for {
		ch := <-ChanLog
		cr, err := memDB.GetCrossFromDevice(ch.ID)
		if err != nil {
			logger.Error.Printf("error %v", ch)
			continue
		}
		reg := pudge.Region{cr.Region, cr.Area, cr.ID}
		crm, is := mapMessages[reg.ToKey()]
		if !is {
			news := make(map[int]string)
			news[0] = ""
			news[1] = ""
			news[2] = ""
			mapMessages[reg.ToKey()] = news
			crm = mapMessages[reg.ToKey()]
		}
		switch ch.Type {
		case 0:
			if strings.Compare(crm[0], ch.LogString) != 0 {
				writeLogDB(cr, ch, 0)
				crm[0] = ch.LogString
			}
		case 1:
			if strings.Compare(crm[1], ch.LogString) != 0 {
				writeLogDB(cr, ch, 1)
				crm[1] = ch.LogString
			}
		case 2:
			if strings.Compare(crm[2], ch.LogString) != 0 {
				writeLogDB(cr, ch, 2)
				crm[2] = ch.LogString
			}
		case -1:
			if strings.Contains(ch.LogString, "Координированное управление ПК=") {
				if strings.Compare(crm[1], "Координированное управление") != 0 {
					temp := ch.LogString
					ch.LogString = "Координированное управление"
					writeLogDB(cr, ch, 1)
					crm[1] = ch.LogString
					ch.LogString = temp
				}
				if strings.Compare(crm[0], ch.LogString) != 0 {
					writeLogDB(cr, ch, 0)
					crm[0] = ch.LogString
				}
				break
			}
			if strings.Compare(crm[0], ch.LogString) != 0 {
				writeLogDB(cr, ch, 0)
				crm[0] = ch.LogString
			}
			if strings.Compare(crm[1], ch.LogString) != 0 {
				writeLogDB(cr, ch, 1)
				crm[1] = ch.LogString
			}

		}
		mapMessages[reg.ToKey()] = crm
	}
}
func writeLogDB(cr pudge.Cross, ch pudge.RecLogCtrl, tup int) {
	j := pudge.JSONLog{Region: strconv.Itoa(cr.Region), Area: strconv.Itoa(cr.Area), ID: cr.ID, Description: cr.Name, Type: tup}
	result, _ := json.Marshal(j)
	w := fmt.Sprintf("insert into public.logdevice (region,id,tm,crossinfo,txt) values(%d,%d,'%s','%s','%s');",
		cr.Region, ch.ID, string(pq.FormatTimestamp(ch.Time)), result, ch.LogString)
	_, err = db.Exec(w)
	if err != nil {
		logger.Error.Printf("Ошибка записи в БД логирования %s \n%s", err.Error(), w)
	}
}
