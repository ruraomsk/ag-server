package loader

import (
	"bufio"
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"github.com/ruraomsk/ag-server/extcon"
	"github.com/ruraomsk/ag-server/logger"
	"github.com/ruraomsk/ag-server/secret"
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
	logger.Info.Printf("Новый клиент SQL сервера %s", soc.RemoteAddr().String())
	dbinfo := fmt.Sprintf("host=%s user=%s password=%s dbname=%s sslmode=disable",
		setup.Set.DataBase.Host, setup.Set.DataBase.User,
		setup.Set.DataBase.Password, setup.Set.DataBase.DBname)
	dbb, err := sql.Open("postgres", dbinfo)
	if err != nil {
		logger.Error.Printf("Запрос на открытие %s %s", dbinfo, err.Error())
		soc.Close()
		stop <- 1
		return
	}
	defer func() {
		dbb.Close()
		soc.Close()
	}()

	if err = dbb.Ping(); err != nil {
		logger.Error.Printf("Ping %s", err.Error())
		stop <- 1
		return
	}
	reader := bufio.NewReader(soc)
	writer := bufio.NewWriter(soc)
	for {
		_ = soc.SetReadDeadline(time.Now().Add(10 * time.Minute))
		c, err := reader.ReadString('\n')
		//&& strings.Compare(err.Error(),"EOF")!=0
		if err != nil {
			logger.Error.Printf("При чтении команд SQL %s сервера %s", soc.RemoteAddr().String(), err.Error())
			return
		}
		if c[0:1] == "0" {
			//logger.Info.Printf("Keep alive from %s", soc.RemoteAddr().String())
			continue
		}
		if setup.Set.Secret {
			c = secret.DecodeString(c)
		}
		responce := false
		if strings.HasPrefix(c, "==RESPONSE NEED==") {
			responce = true
			c = strings.Replace(c, "==RESPONSE NEED==", "", 1)
		}
		if secret.IsSQLValid(c) {
			_, err = dbb.Exec(c)
		} else {
			err = fmt.Errorf("выражение не валидно ")
		}
		//soc.SetWriteDeadline(time.Now().Add(time.Duration(60 * time.Minute)))
		if err != nil {
			w := fmt.Sprintf("Sql %s error %s", c, err.Error())
			logger.Error.Printf(w)
			if responce {
				writer.WriteString(w + "\n")
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
