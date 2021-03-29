package loader

import (
	"bufio"
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"github.com/ruraomsk/TLServer/logger"
	"github.com/ruraomsk/ag-server/extcon"
	"github.com/ruraomsk/ag-server/setup"
	"net"
	"strconv"
	"strings"
	"time"
)

func StartSQL(stop chan int) {
	ln, err := net.Listen("tcp", ":"+strconv.Itoa(setup.Set.Loader.Port))
	if err != nil {
		logger.Error.Printf("Ошибка открытия порта %s", err.Error())
		stop <- 1
		return
	}
	defer ln.Close()
	for {
		socket, err := ln.Accept()
		if err != nil {
			logger.Error.Printf("Ошибка accept %s", err.Error())
			continue
		}
		go workerSQL(socket, stop)
	}
}

func workerSQL(soc net.Conn, stop chan int) {
	defer soc.Close()
	logger.Info.Printf("Новый клиент SQL сервера %s", soc.RemoteAddr().String())
	dbinfo := fmt.Sprintf("host=%s user=%s password=%s dbname=%s sslmode=disable",
		setup.Set.DataBase.Host, setup.Set.DataBase.User,
		setup.Set.DataBase.Password, setup.Set.DataBase.DBname)
	dbb, err := sql.Open("postgres", dbinfo)
	if err != nil {
		logger.Error.Printf("Запрос на открытие %s %s", dbinfo, err.Error())
		stop <- 1
		return
	}
	defer dbb.Close()
	if err = dbb.Ping(); err != nil {
		logger.Error.Printf("Ping %s", err.Error())
		stop <- 1
		return
	}
	reader := bufio.NewReader(soc)
	writer := bufio.NewWriter(soc)
	for {
		c, err := reader.ReadString('\n')
		if err != nil {
			logger.Error.Printf("При чтении команд SQL сервера %s", err.Error())
			return
		}
		if c[0:1] == "0" {
			logger.Info.Println("Keep alive")
			continue
		}
		responce := false
		if strings.HasPrefix(c, "==RESPONSE NEED==") {
			responce = true
			c = strings.Replace(c, "==RESPONSE NEED==", "", 1)
		}
		_, err = dbb.Exec(c)
		if err != nil {
			w := fmt.Sprintf("Sql %s error %s", c, err.Error())
			logger.Error.Printf(w)
			if responce {
				writer.WriteString("w" + "\n")
				writer.Flush()
			}
			continue
		}
		if responce {
			writer.WriteString("ok\n")
			writer.Flush()
		}
	}
}
func RemoteLoader() {
	if !setup.Set.Loader.Make {
		logger.Info.Printf("Remote loader is switch off")
		return
	}
	logger.Info.Printf("Remote loader is started...")
	stop := make(chan int)
	go StartSQL(stop)
	go StartSVG(stop)
	context, _ := extcon.NewContext("Remote loader")
	for {
		select {
		case <-context.Done():
			time.Sleep(3 * time.Second)
			logger.Info.Print("Remote loader is stopped...")
			return
		case <-stop:
			logger.Info.Print("Remote loader is aborted...")
			return
		}
	}

}
