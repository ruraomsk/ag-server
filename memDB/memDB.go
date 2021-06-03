package memDB

import (
	"database/sql"
	"fmt"
	"github.com/ruraomsk/TLServer/logger"
	"github.com/ruraomsk/ag-server/extcon"
	"github.com/ruraomsk/ag-server/setup"
	"time"
)

var memoryDB []*Tx
var db *sql.DB
var id int
var err error

func Start(ready chan interface{}, stop chan interface{}) {
	memoryDB = make([]*Tx, 0)
	info := fmt.Sprintf("host=%s user=%s password=%s dbname=%s sslmode=disable",
		setup.Set.DataBase.Host, setup.Set.DataBase.User,
		setup.Set.DataBase.Password, setup.Set.DataBase.DBname)
	db, err = sql.Open("postgres", info)
	if err != nil {
		logger.Error.Printf("запрос на открытие %s %s", info, err.Error())
		stop <- 0
		return
	}
	needHistoryCross(db)
	initCrosses()
	initDevices()
	initStatus()
	memoryDB = append(memoryDB, CrossesTable)
	memoryDB = append(memoryDB, TableDevices)
	memoryDB = append(memoryDB, StatusTable)

	for _, t := range memoryDB {
		t.Lock()
		t.MDB.Data = t.ReadAll()
		ss := t.GetAllKeys()
		if t.writable {
			for _, s := range ss {
				t.updated[s] = true
			}
		}
		t.Lock()
	}
	ready <- 0
	logger.Info.Println("memDB start work")
	updateTablesTicker := time.NewTicker(time.Duration(setup.Set.StepPudge) * time.Second)
	p, _ := extcon.NewContext("memDB")
	for true {
		select {
		case <-updateTablesTicker.C:
			for _, t := range memoryDB {
				if t.Save() != nil {
					return
				}
			}
		case <-p.Done():
			for _, t := range memoryDB {
				_ = t.Save()
			}
			db.Close()
			logger.Info.Println("memDB end work")
			return

		}
	}
}
