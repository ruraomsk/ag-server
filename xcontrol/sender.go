package xcontrol

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/ruraomsk/TLServer/logger"
	"github.com/ruraomsk/ag-server/comm"
	"github.com/ruraomsk/ag-server/setup"
	"net"
	"strconv"
	"time"
)

//Sender Посылает новые планы координации на устройства
func Sender() {
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
	soc, err := net.Dial("tcp", "localhost:"+strconv.Itoa(setup.Set.CommServer.PortCommand))
	if err != nil {
		logger.Error.Printf("Соединение с сервером команд %s", err.Error())
		return
	}
	_, err = soc.Write([]byte("0\n"))
	if err != nil {
		logger.Error.Printf("Передача keep alive на сервер команд %s", err.Error())
		return
	}
	defer conDB.Exec("rollback;")
	for true {
		time.Sleep(time.Duration(setup.Set.XCtrl.StepCalc) * time.Second)
		//logger.Info.Printf("Управление по характерным точка стадия 2....")
		_, err = conDB.Exec("begin;")
		if err != nil {
			logger.Error.Printf("Запрос begin %s", err.Error())
			return
		}
		comms := make([]comm.CommandARM, 0)
		rows, err := conDB.Query("select state from public.xctrl;")
		if err != nil {
			logger.Error.Printf("Запрос  select state from public.xctrl; %s", err.Error())
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
			if !v.Release {
				v.PKNow = 0
			}
			if v.PKNow != v.PKLast {
				v.LastTime = time.Now()
				w := fmt.Sprintf("select idevice from public.cross where region = %d and area=%d and subarea = %d;", v.Region, v.Area, v.SubArea)
				cross, err := conDB.Query(w)
				if err != nil {
					logger.Error.Printf("Запрос  %s %s", w, err.Error())
					return
				}
				logger.Info.Printf("Регион %d район %d подрайон %d новый план %d", v.Region, v.Area, v.SubArea, v.PKNow)
				for cross.Next() {
					var idevice int
					err = cross.Scan(&idevice)
					c := comm.CommandARM{ID: idevice, User: "XCtrl", Command: 5, Params: v.PKNow}
					comms = append(comms, c)
				}
				v.PKLast = v.PKNow
				v.LastTime = time.Now()
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
		}

		_, err = conDB.Exec("commit;")
		if err != nil {
			logger.Error.Printf("Запрос commit %s", err.Error())
			return
		}
		if len(comms) == 0 {
			_, err = soc.Write([]byte("0\n"))
			if err != nil {
				logger.Error.Printf("Передача keep alive на сервер команд %s", err.Error())
				return
			}
		} else {
			for _, com := range comms {
				//logger.Info.Printf("send %v",com)
				c, err := json.Marshal(com)
				if err != nil {
					logger.Error.Printf("Конвертация команды %v %s", com, err.Error())
					return
				}
				c = append(c, '\n')
				_, err = soc.Write(c)
				if err != nil {
					logger.Error.Printf("Передача %s на сервер команд %s", string(c), err.Error())
					return
				}
				time.Sleep(1 * time.Millisecond)
			}
		}
	}

	return
}
