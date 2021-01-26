package dumper

import (
	"github.com/JanFant/TLServer/logger"
	"github.com/jasonlvhit/gocron"
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
