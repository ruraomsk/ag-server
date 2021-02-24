package dumper

import (
	"github.com/jasonlvhit/gocron"
	"github.com/ruraomsk/TLServer/logger"
	"github.com/ruraomsk/ag-server/pudge"
	"github.com/ruraomsk/ag-server/setup"
	"os"
	"os/exec"
	"runtime"
	"time"
)

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
	file.WriteString("SET PGPASSWORD=" + setup.Set.DataBase.Password + "\n")
	date := time.Now().Format(time.RFC3339)[0:10]
	path := setup.Set.Dumper.Path + "/dump" + date + ".sql"
	file.WriteString("pg_dump -U" + setup.Set.DataBase.User + " -d" + setup.Set.DataBase.DBname +
		" -C -c --column-inserts --if-exists --no-comments -f" + path + "\n")
	path = setup.Set.Dumper.Path + "/cross" + date + ".tar"
	file.WriteString("tar -cvf " + path + " " + setup.Set.Dumper.PathSVG + "\n")

	file.Close()
	cmd := exec.Command("save.bat")
	err = cmd.Run()
	if err != nil {
		logger.Error.Printf("Не могу выполнить save.bat %s", err.Error())
		return
	}
	logger.Info.Printf("Dump writed..")

}
func makeStatistics() {
	regs := pudge.GetCrosses()
	for _, reg := range regs {
		cross, is := pudge.GetCross(reg.Region, reg.Area, reg.ID)
		if !is {
			continue
		}
		ctrl, is := pudge.GetController(cross.IDevice)
		if !is {
			continue
		}
		arch := new(pudge.ArchStat)
		arch.Region = reg.Region
		arch.Area = reg.Area
		arch.ID = reg.ID
		arch.Date = time.Now().Add(-time.Hour * 24)
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
func Statistics() {
	logger.Info.Printf("Statistics starting..")
	go writerArch()
	gocron.Every(1).Day().At(setup.Set.XCtrl.ClearTime).Do(makeStatistics)
	<-gocron.Start()
	logger.Info.Printf("Statistics working..")

}
func Start() {

	if !setup.Set.Dumper.Make {
		logger.Info.Printf("Dumper dont start..")

		return
	}
	logger.Info.Printf("Dumper starting..")
	gocron.Every(1).Day().At(setup.Set.Dumper.Time).Do(makeDump)
	<-gocron.Start()
	logger.Info.Printf("Dumper working..")
}
