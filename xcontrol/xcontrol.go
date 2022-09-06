package xcontrol

import (
	"bufio"
	"database/sql"
	"encoding/json"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/jasonlvhit/gocron"
	"github.com/ruraomsk/ag-server/logger"
	"github.com/ruraomsk/ag-server/pudge"

	//Инициализатор постргресса
	_ "github.com/lib/pq"
	"github.com/ruraomsk/ag-server/extcon"
	"github.com/ruraomsk/ag-server/setup"
)

//Данный пакет производит управление по характерным точкам
// Разбит на два раздела
// 	в первом разделе производится расчет характерной точки и выбор стратегии
// 	во втором разделе производится выполнение выбранной стратегии для каждого района и подрайона

var command chan int
var commARM chan pudge.CommandARM
var work = false
var dbb *sql.DB
var err error
var mainTable *Table
var stats []ExtState
var UserName string
var FirstCalculate bool
var viewer = false

func listenCommand() {
	ln, err := net.Listen("tcp", ":"+strconv.Itoa(setup.Set.XCtrl.Port))
	if err != nil {
		logger.Error.Printf("Ошибка открытия порта %s", err.Error())
		command <- -1
		return
	}
	defer ln.Close()
	for {
		socket, err := ln.Accept()
		if err != nil {
			logger.Error.Printf("Ошибка accept %s", err.Error())
			continue
		}
		go worker(socket)
	}
}

func worker(soc net.Conn) {
	defer soc.Close()

	logger.Info.Printf("Новый клиент сервера ХТ %s", soc.RemoteAddr().String())
	reader := bufio.NewReader(soc)
	writer := bufio.NewWriterSize(soc, 124*1024*1024)
	for {
		cmd, err := reader.ReadString('\n')
		if err != nil {
			logger.Error.Printf("При чтении команд сервера ХТ от клиента %s %s", soc.RemoteAddr().String(), err.Error())
			return
		}
		cmd = strings.Replace(cmd, "\n", "", 1)
		if cmd[0:1] == "0" {
			//logger.Info.Println("Keep alive")
			continue
		}
		if !viewer {
			_, _ = writer.WriteString("BAD\n")
			_ = writer.Flush()
			continue
		}
		//logger.Info.Printf("От сервера %s пришла команда %s", soc.RemoteAddr().String(), cmd)
		if strings.Contains(cmd, "restart") {
			command <- 1
			continue
		}
		if strings.HasPrefix(cmd, "setup") {
			result, err := json.Marshal(setup.Set.XCtrl)
			if err != nil {
				logger.Error.Println(err.Error())
				_, _ = writer.WriteString("{}")
			} else {
				_, _ = writer.WriteString(string(result))
			}
			_, _ = writer.WriteString("\n")
			_ = writer.Flush()
			continue
		}
		if strings.HasPrefix(cmd, "crosslist") {
			_, _ = writer.WriteString(mainTable.listTables())
			_, _ = writer.WriteString("\n")
			_ = writer.Flush()
			continue
		}
		if strings.HasPrefix(cmd, "statelist") {
			_, _ = writer.WriteString(listStates())
			_, _ = writer.WriteString("\n")
			_ = writer.Flush()
			continue
		}
		if strings.HasPrefix(cmd, "stateset") {
			ls := strings.Split(cmd, ",")
			region, _ := strconv.Atoi(ls[1])
			area, _ := strconv.Atoi(ls[2])
			id, _ := strconv.Atoi(ls[3])
			command, _ := strconv.Atoi(ls[4])
			changeState(pudge.Region{Region: region, Area: area, ID: id}, command)
			_, _ = writer.WriteString("Ok\n")
			_ = writer.Flush()
			continue
		}
		if strings.HasPrefix(cmd, "devicecmd") {
			ls := strings.Split(cmd, ",")
			idevice, _ := strconv.Atoi(ls[1])
			code, _ := strconv.Atoi(ls[2])
			command, _ := strconv.Atoi(ls[3])
			//logger.Debug.Printf("in %v",comm.CommandARM{ID: idevice, User: UserName, Command: code, Params: command})
			commARM <- pudge.CommandARM{ID: idevice, User: UserName, Command: code, Params: command}
			_, _ = writer.WriteString("Ok\n")
			_ = writer.Flush()
			continue
		}
		if strings.HasPrefix(cmd, "crossget") {
			ls := strings.Split(cmd, ",")
			region, _ := strconv.Atoi(ls[1])
			area, _ := strconv.Atoi(ls[2])
			id, _ := strconv.Atoi(ls[3])
			_, _ = writer.WriteString(mainTable.getXCross(pudge.Region{
				Region: region,
				Area:   area,
				ID:     id,
			}))
			_, _ = writer.WriteString("\n")
			_ = writer.Flush()
			continue
		}
		if strings.HasPrefix(cmd, "stateget") {
			ls := strings.Split(cmd, ",")
			region, _ := strconv.Atoi(ls[1])
			area, _ := strconv.Atoi(ls[2])
			id, _ := strconv.Atoi(ls[3])
			_, _ = writer.WriteString(getState(pudge.Region{
				Region: region,
				Area:   area,
				ID:     id,
			}))
			_, _ = writer.WriteString("\n")
			_ = writer.Flush()
			continue
		}
		if strings.HasPrefix(cmd, "dataget") {
			ls := strings.Split(cmd, ",")
			region, _ := strconv.Atoi(ls[1])
			area, _ := strconv.Atoi(ls[2])
			id, _ := strconv.Atoi(ls[3])
			_, _ = writer.WriteString(getData(pudge.Region{
				Region: region,
				Area:   area,
				ID:     id,
			}, ls[4]))
			_, _ = writer.WriteString("\n")
			_ = writer.Flush()
			continue
		}
		if strings.HasPrefix(cmd, "messages") {
			_, _ = writer.WriteString(getMessages())
			_, _ = writer.WriteString("\n")
			_ = writer.Flush()
			continue
		}
	}

}
func startCron() {
	<-gocron.Start()
}

//Start главный модуль регулятора
func Start(ready, stop chan interface{}) {
	context, _ := extcon.NewContext("xctrl")
	logger.Info.Print("Модуль управления по характерным старт... ")
	dbinfo := fmt.Sprintf("host=%s user=%s password=%s dbname=%s sslmode=disable",
		setup.Set.DataBase.Host, setup.Set.DataBase.User,
		setup.Set.DataBase.Password, setup.Set.DataBase.DBname)
	for {
		dbb, err = sql.Open("postgres", dbinfo)
		if err != nil {
			logger.Error.Printf("Запрос на открытие %s %s", dbinfo, err.Error())
			time.Sleep(time.Second * 10)
			continue
		}
		if err = dbb.Ping(); err != nil {
			logger.Error.Printf("Ping %s", err.Error())
			time.Sleep(time.Second * 10)
			continue
		}
		break

	}
	defer dbb.Close()
	work = true
	mainTable = new(Table)
	stats = make([]ExtState, 0)
	UserName = setup.Set.XCtrl.NameUser
	err := makeTable()
	if err != nil {
		logger.Error.Printf("Ошибка создания таблицы %s", err.Error())
		return
	}
	loadTable()
	clearError()
	go listenCommand()
	go workerXTCommand()
	command = make(chan int)
	commARM = make(chan pudge.CommandARM, 100)

	go sender()
	fmt.Println("Можно загружать просмотр...")
	logger.Info.Println("Можно загружать просмотр...")
	ready <- 0
	for {
		t := time.Now()
		if t.Minute()%setup.Set.XCtrl.StepDev == 0 && t.Second() == 0 {
			break
		}
		time.Sleep(1 * time.Second)
	}
	logger.Info.Print("Скорректировали время запуска... ")
	for _, reg := range setup.Set.Statistic.Regions {
		region, _ := strconv.Atoi(reg[0])
		_ = gocron.Every(1).Day().At(reg[1]).Do(clearRegion, region)
	}
	_ = gocron.Every(1).Day().At("0:00").Do(clearError)
	go startCron()
	for {
		viewer = false
		err := makeTable()
		if err != nil {
			logger.Error.Printf("Ошибка создания таблицы %s", err.Error())
			return
		}
		FirstCalculate = true
		calculate()
		FirstCalculate = false
		logger.Info.Print("Модуль управления по характерным точкам запущен... ")
		needRestart := false
		secondsTicker := time.NewTicker(time.Second)
		for !needRestart {
			select {
			case <-context.Done():
				work = false
				time.Sleep(5 * time.Second)
				return
			case <-stop:
				return
			case <-command:
				needRestart = true
			case <-secondsTicker.C:
				t := time.Now()
				if t.Minute()%setup.Set.XCtrl.StepDev == 0 && t.Second() == 0 {
					loadTable()
					calculate()
				}
			}
		}

	}

}
