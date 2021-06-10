package dumper

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/jasonlvhit/gocron"
	"github.com/ruraomsk/TLServer/logger"
	"github.com/ruraomsk/ag-server/extcon"
	"github.com/ruraomsk/ag-server/pudge"
	"github.com/ruraomsk/ag-server/setup"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"time"
)

var conDB *sql.DB
var err error
var workStatistic bool

func makeDump() {
	if runtime.GOOS == "linux" {
		logger.Error.Printf("Нет реализации для Linux!")
		return
	}
	file, err := os.Create("save.bat")
	if err != nil {
		logger.Error.Printf("Не могу создать файл save.bat %s", err.Error())
		return
	}
	_, _ = file.WriteString("SET PGPASSWORD=" + setup.Set.DataBase.Password + "\n")
	date := time.Now().Format(time.RFC3339)[0:10]
	path := setup.Set.Dumper.Path + "/dump" + date + ".sql"
	_, _ = file.WriteString("pg_dump -U" + setup.Set.DataBase.User + " -d" + setup.Set.DataBase.DBname +
		" -C -c --column-inserts --if-exists --no-comments -f" + path + "\n")
	path = setup.Set.Dumper.Path + "/cross" + date + ".tar"
	_, _ = file.WriteString("tar -cvf " + path + " " + setup.Set.Dumper.PathSVG + "\n")

	file.Close()
	time.Sleep(5 * time.Second)
	cmd := exec.Command("save.bat")
	err = cmd.Run()
	if err != nil {
		logger.Error.Printf("Не могу выполнить save.bat %s", err.Error())
		return
	}
	logger.Info.Printf("Dump writed..")

}
func makeStatistics(region, flag int) {
	regs := pudge.GetCrosses()
	for _, reg := range regs {
		if !workStatistic {
			return
		}
		if reg.Region != region {
			continue
		}
		cross, is := pudge.GetCross(reg.Region, reg.Area, reg.ID)
		if !is {
			continue
		}
		ctrl, is := pudge.GetController(cross.IDevice)
		if !is {
			//Значит чистим в БД
			dev, err := conDB.Query(fmt.Sprintf("SELECT device FROM public.devices where id=%d; ", cross.IDevice))
			if err != nil {
				logger.Error.Printf("При чтении из БД устройства %d %s", cross.IDevice, err.Error())
				return
			}
			defer dev.Close()
			for dev.Next() {
				var js []byte
				c := new(pudge.Controller)
				err = dev.Scan(&js)
				if err != nil {
					logger.Error.Printf("При чтении из БД устройства %d %s", cross.IDevice, err.Error())
				}
				err = json.Unmarshal(js, &c)
				if err != nil {
					logger.Error.Printf("При чтении из БД устройства %d %s", cross.IDevice, err.Error())
				}
				c.Statistics = make([]pudge.Statistic, 0)
				js, err = json.Marshal(c)
				if err != nil {
					logger.Error.Printf("При записи в БД устройства %d %s", cross.IDevice, err.Error())
				}
				_, err = conDB.Exec("update  devices set device='" + string(js) + "' where id=" + strconv.Itoa(c.ID) + ";")
				if err != nil {
					logger.Error.Printf("При записи в БД устройства %d %s", cross.IDevice, err.Error())
					return
				}
			}
			dev.Close()
			continue
		}
		arch := new(pudge.ArchStat)
		arch.Region = reg.Region
		arch.Area = reg.Area
		arch.ID = reg.ID
		switch flag {
		case -1:
			arch.Date = time.Now().Add(-time.Hour * 24)
		case 0:
			arch.Date = time.Now()
		case 1:
			arch.Date = time.Now().Add(time.Hour * 24)
		}
		arch.Statistics = make([]pudge.Statistic, 0)
		for _, s := range ctrl.Statistics {
			arch.Statistics = append(arch.Statistics, s)
		}
		if len(arch.Statistics) != 0 {
			ctrl.Statistics = make([]pudge.Statistic, 0)
			pudge.SetController(ctrl)
			writeArch <- *arch
		}
	}
}
func goCron() {
	<-gocron.Start()

}
func Statistics() {
	logger.Info.Printf("Statistics is collected..")
	dbinfo := fmt.Sprintf("host=%s user=%s password=%s dbname=%s sslmode=disable",
		setup.Set.DataBase.Host, setup.Set.DataBase.User,
		setup.Set.DataBase.Password, setup.Set.DataBase.DBname)
	conDB, err = sql.Open("postgres", dbinfo)
	if err != nil {
		logger.Error.Printf("Запрос на открытие %s %s", dbinfo, err.Error())
		return
	}
	defer conDB.Close()
	if err = conDB.Ping(); err != nil {
		logger.Error.Printf("Ping %s", err.Error())
		return
	}
	workStatistic = true
	go writerArch()
	context, _ := extcon.NewContext("Statistic")
	timer := extcon.SetTimerClock(time.Duration(1 * time.Minute))

	for _, reg := range setup.Set.Statistic.Regions {
		region, _ := strconv.Atoi(reg[0])
		flag, _ := strconv.Atoi(reg[2])
		_ = gocron.Every(1).Day().At(reg[1]).Do(makeStatistics, region, flag)
	}
	go goCron()
	for {
		select {
		case <-context.Done():
			workStatistic = false
			time.Sleep(3 * time.Second)
			logger.Info.Print("Statistic is stopped...")
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
func Start() {
	logger.Info.Printf("Dumper starting..")
	_ = gocron.Every(1).Day().At(setup.Set.Dumper.Time).Do(makeDump)
	<-gocron.Start()
	logger.Info.Printf("Dumper working..")
}
