package create

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"io/ioutil"
	"rura/ag-server/logger"
	"rura/ag-server/setup"
	"strings"
)

//SqlCreate просмотр каталога и исполнить все запросы с расширением create
func SQLCreate(path string) error {
	dbinfo := fmt.Sprintf("host=%s user=%s password=%s dbname=%s sslmode=disable",
		setup.Set.DataBase.Host, setup.Set.DataBase.User,
		setup.Set.DataBase.Password, setup.Set.DataBase.DBname)
	con, err := sql.Open("postgres", dbinfo)
	if err != nil {
		logger.Error.Printf("Запрос на открытие %s %s", dbinfo, err.Error())
		return err
	}
	defer con.Close()
	if err = con.Ping(); err != nil {
		logger.Error.Printf("Ping %s", err.Error())
		return err
	}
	dirs, err := ioutil.ReadDir(path)
	if err != nil {
		logger.Error.Printf("Ошибка чтения содержимого кталога %s %s", path, err.Error())
		return err
	}
	for _, dir := range dirs {
		if dir.IsDir() {
			continue
		}
		if !strings.HasSuffix(dir.Name(), ".sql") {
			continue
		}
		nfile := path + "/" + dir.Name()
		cmd, err := ioutil.ReadFile(nfile)
		if err != nil {
			logger.Error.Printf("Error reading file %s! %s\n", path, err.Error())
			return err
		}
		logger.Info.Printf("Обрабатываем файл %s", nfile)
		_, err = con.Exec(string(cmd))

		if err != nil {
			logger.Error.Printf("Error create  %s\n", err.Error())
			return err
		}

	}
	return nil
}
