package dumper

import (
	"io/ioutil"
	"os"
	"os/exec"
	"runtime"
	"time"

	"github.com/jasonlvhit/gocron"
	"github.com/ruraomsk/ag-server/logger"
	"github.com/ruraomsk/ag-server/setup"
)

func dumpClean(path string) {
	files, err := ioutil.ReadDir(path)
	if err != nil {
		logger.Error.Printf("Error reading directory with dump files %s", path)
	}
	for _, file := range files {
		if file.ModTime().Add(time.Hour * 24 * 30).Before(time.Now()) {
			_ = os.Remove(path + "/" + file.Name())
		}
	}
}

func makeDump() {
	dumpClean(setup.Set.Dumper.Path)
	if runtime.GOOS == "linux" {

		save := "#!/bin/bash\n"
		save += "PGPASSWORD=" + setup.Set.DataBase.Password + "\n"
		save += "export PGPASSWORD\n"
		save += "su -l " + setup.Set.DataBase.User + " -c "
		date := time.Now().Format(time.RFC3339)[0:10]
		path := setup.Set.Dumper.Path + "/dump" + date + ".sql"
		save += "\"pg_dump " + " -d" + setup.Set.DataBase.DBname +
			" -C -c  -f" + path + "\"\n"
		path = setup.Set.Dumper.Path + "/cross" + date + ".tar"
		save += "tar -cvf " + path + " " + setup.Set.Dumper.PathSVG + "\n"

		chmod := exec.Command("bash", save)
		err := chmod.Run()
		if err != nil {
			logger.Error.Printf("Не могу %s %s", save, err.Error())
		} else {
			logger.Info.Printf("Dump writed..")
		}
		return
	}
	save := "SET PGPASSWORD=" + setup.Set.DataBase.Password + "\n"
	date := time.Now().Format(time.RFC3339)[0:10]
	path := setup.Set.Dumper.Path + "/dump" + date + ".sql"
	save += "pg_dump -U" + setup.Set.DataBase.User + " -d" + setup.Set.DataBase.DBname +
		" -C -c --column-inserts --if-exists --no-comments -f" + path + "\n"
	path = setup.Set.Dumper.Path + "/cross" + date + ".tar"
	save += "tar -cvf " + path + " " + setup.Set.Dumper.PathSVG + "\n"
	os.Remove("save.bat")
	file, err := os.Create("save.bat")
	if err != nil {
		logger.Error.Printf("Не могу записать save.bat %s", err.Error())
		return
	}
	file.WriteString(save)
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
	logger.Info.Printf("Dumper starting..")
	_ = gocron.Every(1).Day().At(setup.Set.Dumper.Time).Do(makeDump)
	<-gocron.Start()
	logger.Info.Printf("Dumper working..")
}
