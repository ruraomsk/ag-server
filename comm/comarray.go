package comm

import (
	"bufio"
	"database/sql"
	"encoding/json"
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/lib/pq"

	"github.com/ruraomsk/ag-server/binding"
	"github.com/ruraomsk/ag-server/extcon"
	"github.com/ruraomsk/ag-server/logger"
	"github.com/ruraomsk/ag-server/pudge"
	"github.com/ruraomsk/ag-server/setup"
)

// DevPhases для передачи фаз
type DevPhases struct {
	ID int      `json:"idevice"`
	DK pudge.DK `json:"dk"`
}

var connectMap map[string]net.Conn
var cMutex sync.Mutex
var cSFDK chan pudge.CommandARM

func listenArmCommand() {
	ln, err := net.Listen("tcp", ":"+strconv.Itoa(setup.Set.CommServer.PortCommand))
	if err != nil {
		logger.Error.Printf("Ошибка открытия порта %s", err.Error())
		return
	}
	cSFDK = make(chan pudge.CommandARM)
	go sfdkControl()
	go routeSgdk()
	for {
		socket, err := ln.Accept()
		if err != nil {
			logger.Error.Printf("Ошибка accept %s", err.Error())
			continue
		}
		go workerCommand(socket)
	}
}
func listenChangeProtocol() {
	ln, err := net.Listen("tcp", ":"+strconv.Itoa(setup.Set.CommServer.PortProtocol))
	if err != nil {
		logger.Error.Printf("Ошибка открытия порта %s", err.Error())
		return
	}
	//defer ln.Close()
	for {
		socket, err := ln.Accept()
		if err != nil {
			logger.Error.Printf("Ошибка accept %s", err.Error())
			continue
		}
		go workerProtocol(socket)
	}
}

func listenArmArray() {
	ln, err := net.Listen("tcp", ":"+strconv.Itoa(setup.Set.CommServer.PortArray))

	if err != nil {
		logger.Error.Printf("Ошибка открытия порта %s", err.Error())
		return
	}
	//defer ln.Close()
	for {
		socket, err := ln.Accept()
		if err != nil {
			logger.Error.Printf("Ошибка accept %s", err.Error())
			continue
		}
		go workerArray(socket)
	}
}
func listenSendingPhazes() {
	ln, err := net.Listen("tcp", ":"+strconv.Itoa(setup.Set.CommServer.PortDevices))
	connectMap = make(map[string]net.Conn)
	go workerDevices()
	if err != nil {
		logger.Error.Printf("Ошибка открытия порта %s", err.Error())
		return
	}
	//defer ln.Close()
	for {
		socket, err := ln.Accept()
		if err != nil {
			logger.Error.Printf("Ошибка accept %s", err.Error())
			continue
		}
		logger.Info.Printf("Новый клиент фаз устройства %s", socket.RemoteAddr().String())
		cMutex.Lock()
		connectMap[socket.RemoteAddr().String()] = socket
		cMutex.Unlock()
	}
}
func workerDevices() {
	timer := extcon.SetTimerClock(10 * time.Second)
	// writer := bufio.NewWriter(soc)
	for {
		select {
		case <-timer.C:
			for _, soc := range connectMap {
				_, err := fmt.Fprintf(soc, "0\n")
				if err != nil {
					logger.Error.Printf("Ошибка передачи tcp %s %s", soc.RemoteAddr().String(), err.Error())
					cMutex.Lock()
					delete(connectMap, soc.RemoteAddr().String())
					cMutex.Unlock()
				}
			}
		case d := <-sendPhases:
			array, err := json.Marshal(&d)
			if err != nil {
				logger.Error.Printf("Ошибка json %s", err.Error())
				continue
			}
			for _, soc := range connectMap {
				_, err = fmt.Fprintf(soc, string(array)+"\n")
				if err != nil {
					logger.Error.Printf("Ошибка передачи tcp %s %s", soc.RemoteAddr().String(), err.Error())
					cMutex.Lock()
					delete(connectMap, soc.RemoteAddr().String())
					cMutex.Unlock()
				}
			}
		}
	}

}
func workerCommand(soc net.Conn) {
	defer soc.Close()
	var command pudge.CommandARM
	phases := make(map[int]int)
	logger.Info.Printf("Новый клиент команд %s", soc.RemoteAddr().String())
	reader := bufio.NewReader(soc)
	for {
		c, err := reader.ReadString('\n')
		if err != nil {
			logger.Error.Printf("При чтении команд сервера АРМ %s %s ", soc.RemoteAddr().String(), err.Error())
			return
		}
		if c[0:1] == "0" {
			// logger.Info.Println("Keep alive")
			continue
		}
		err = json.Unmarshal([]byte(c), &command)
		if err != nil {
			logger.Error.Println("При конвертации команд сервера АРМ ", err.Error())
			continue
		}
		if command.Command == 4 {
			cSFDK <- command
			continue
		}
		if command.Command == 13 {
			//Команды включения и выключения для ХТ в id лежит регион,район и подрайон
			region := (command.ID >> 16) & 0xff
			area := (command.ID >> 8) & 0xff
			subarea := command.ID & 0xff
			mess := pudge.CommandXT{Region: pudge.Region{Region: region, Area: area, ID: subarea}, Command: command.Params}
			logger.Info.Printf("Команда сервера АРМ %v", mess)
			pudge.XTCommand <- mess
			continue
		}
		if command.Command == 12 {
			ctrl, is := pudge.GetController(command.ID)
			if !is {
				logger.Error.Printf("Команда сервера АРМ %v нет такого id %d", command, command.ID)
				continue
			}
			ctrl.LogLines = make([]pudge.LogLine, 0)
			pudge.SetController(ctrl)
			logger.Info.Printf("Почищен внутренний журнал %d", command.ID)
			continue
		}
		dev, is := getDevice(command.ID)
		if !is {
			if strings.Compare(command.User, setup.Set.XCtrl.NameUser) != 0 {
				logger.Error.Printf("Команда сервера АРМ %v нет такого id %d", command, command.ID)
			}
			continue
		}
		_, is = phases[command.ID]
		if !is {
			phases[command.ID] = 0
		}
		if command.Command < 0x0e {
			logger.Info.Printf("Команда %v %s", command, soc.RemoteAddr().String())
		}
		if command.Command == 1 {
			//Принудительная отправка всех массивов
			pudge.ChanLog <- pudge.LogRecord{ID: command.ID, Region: pudge.Region{Region: 0}, Type: 1, Time: time.Now(), Journal: pudge.UserDeviceStatus(command.User, -1, 0)}
			ctrl, _ := pudge.GetController(command.ID)
			ctrl.Arrays = pudge.MakeArrays(*binding.NewArrays())
			pudge.SetController(ctrl)
			logger.Info.Printf("id %d массив привязок поставлен на перезагрузку", command.ID)
		} else {
			if command.Command == 9 {
				if command.Params > 0 && command.Params < 9 {
					if command.Params != phases[command.ID] {
						pudge.ChanLog <- pudge.LogRecord{ID: command.ID, Region: pudge.Region{Region: 0}, Type: 1, Time: time.Now(), Journal: pudge.UserDeviceStatus(command.User, command.Command, command.Params)}
						phases[command.ID] = command.Params
					}
				} else {
					phases[command.ID] = 0
					pudge.ChanLog <- pudge.LogRecord{ID: command.ID, Region: pudge.Region{Region: 0}, Type: 1, Time: time.Now(), Journal: pudge.UserDeviceStatus(command.User, command.Command, command.Params)}
				}
			}
			if command.Command == 5 || command.Command == 6 || command.Command == 7 {
				phases[command.ID] = 0
				pudge.ChanLog <- pudge.LogRecord{ID: command.ID, Region: pudge.Region{Region: 0}, Type: 0, Time: time.Now(), Journal: pudge.UserTechStatus(command.ID, command.User, command.Command, command.Params)}
			}
			if isCorectCommand(command) {
				dev.CommandARM <- command
			} else {
				logger.Debug.Printf("Команда %v запрещена", command)
			}
		}
	}
}
func workerArray(soc net.Conn) {
	dbinfo := fmt.Sprintf("host=%s user=%s password=%s dbname=%s sslmode=disable",
		setup.Set.DataBase.Host, setup.Set.DataBase.User,
		setup.Set.DataBase.Password, setup.Set.DataBase.DBname)
	db, err := sql.Open("postgres", dbinfo)
	if err != nil {
		logger.Error.Printf("Запрос на открытие %s %s", dbinfo, err.Error())
		soc.Close()
		return
	}

	defer func() {
		soc.Close()
		db.Close()
	}()

	logger.Info.Printf("Новый клиент массивов %s", soc.RemoteAddr().String())
	reader := bufio.NewReader(soc)
	for {
		a, err := reader.ReadString('\n')
		if err != nil {
			logger.Error.Println("При чтении привязки от сервера АРМ ", err.Error())
			return
		}
		// logger.Debug.Printf("=%v=", a)
		if a[0:1] == "0" {
			// logger.Info.Println("Keep alive")
			continue
		}
		state := new(pudge.UserCross)
		err = json.Unmarshal([]byte(a), &state)
		if err != nil {
			logger.Error.Printf("При конвертации привязки сервера АРМ %s %s", a, err.Error())
			continue
		}
		// logger.Error.Println("Пришло state")
		if state.State.IDevice < 0 {
			//Удаление перекрестка
			last, is := pudge.GetCross(pudge.Region{Region: state.State.Region, Area: state.State.Area, ID: state.State.ID})
			if !is {
				//Перекрестка нет
				logger.Info.Printf("Попытка удаления удаленного перекрестка %d %d %d", state.State.Region, state.State.Area, state.State.ID)
				continue
			}
			logger.Info.Printf("Удаление перекрестка %d %d %d %d", state.State.Region, state.State.Area, state.State.ID, last.IDevice)
			// w := fmt.Sprintf("%s удаление перекрестка %d %d %d ", state.User, state.State.Region, state.State.Area, state.State.ID)
			pudge.DeleteCross(pudge.Region{Region: state.State.Region, Area: state.State.Area, ID: state.State.ID})
			continue
		}
		old, is := pudge.GetCross(pudge.Region{Region: state.State.Region, Area: state.State.Area, ID: state.State.ID})
		if !is {
			//Перекрестка нет нужно создать
			state.State.StatusDevice = 18
			w := fmt.Sprintf("%s добавил перекрестка %d %d %d %d", state.User, state.State.Region, state.State.Area, state.State.ID, state.State.IDevice)
			logger.Info.Print(w)
			pudge.SetCross(state.State)
			// time.Sleep(1 * time.Second)
			continue
		}
		os, _ := json.Marshal(&old)
		// logger.Info.Printf("Write new state ")
		if old.IDevice != state.State.IDevice {
			logger.Info.Printf("Отключаем старое устройство %d ", old.IDevice)
			// w := fmt.Sprintf("%s заменил устройство с %d на %d ", state.User, old.IDevice, state.State.ID)
			d, ok := getDevice(old.IDevice)
			if ok {
				//Остановим текущее
				d.ExitCommand <- 1
				time.Sleep(5 * time.Second)
				killDevice(old.IDevice)
				time.Sleep(time.Second)
			}
		}
		pudge.SetCross(state.State)
		w := fmt.Sprintf("%s изменил перекресток %d %d %d", state.User, state.State.Region, state.State.Area, state.State.ID)
		s := fmt.Sprintf("insert into public.history (region,area,id,login,tm,state) values (%d,%d,%d,'%s','%s','%s');",
			state.State.Region, state.State.Area, state.State.ID, state.User, string(pq.FormatTimestamp(time.Now())), string(os))
		_, _ = db.Exec(s)
		logger.Info.Print(w)
	}
}
func workerProtocol(soc net.Conn) {
	defer soc.Close()
	var protocol ChangeProtocol
	logger.Info.Printf("Новый клиент протокола %s", soc.RemoteAddr().String())
	reader := bufio.NewReader(soc)
	for {
		c, err := reader.ReadString('\n')
		if err != nil {
			logger.Error.Println("При чтении изменения протокола от АРМ ", err.Error())
			return
		}
		if c[0:1] == "0" {
			// logger.Info.Println("Keep alive")
			continue
		}
		err = json.Unmarshal([]byte(c), &protocol)
		if err != nil {
			logger.Error.Println("При конвертации изменения протокола АРМ ", err.Error())
			continue
		}
		dev, is := getDevice(protocol.ID)
		if !is {
			logger.Error.Printf("Команда протокола АРМ %v нет такого id %d", protocol, protocol.ID)
			continue
		}
		w := fmt.Sprintf("%s send command %v", protocol.User, protocol)
		logger.Info.Print(w)
		dev.ChangeProtocol <- protocol
	}
}

// func getDescription(toSend CommandARM) string {
// 	switch toSend.Command {
// 	case 4:
// 		if toSend.Params == 1 {
// 			return "Отправлен запрос на смену фаз"
// 		}
// 		return "Отключить запрос на смену фаз"
// 	case 5:
// 		if toSend.Params == 0 {
// 			return "Переход на автоматическое регулирование ПК"
// 		}
// 		return "Сменить ПК на " + strconv.Itoa(toSend.Params)
// 	case 6:
// 		if toSend.Params == 0 {
// 			return "Переход на автоматическое регулирование СК"
// 		}
// 		return "Сменить CК на " + strconv.Itoa(toSend.Params)
// 	case 7:
// 		if toSend.Params == 0 {
// 			return "Переход на автоматическое регулирование НК"
// 		}
// 		return "Сменить НК на " + strconv.Itoa(toSend.Params)
// 	}
// 	switch toSend.Params {
// 	case 0:
// 		return "Переход в Локальный режим"
// 	case 9:
// 		return "Переход в  Координированное управление"
// 	case 10:
// 		return "Включить жёлтое мигание"
// 	case 11:
// 		return "Отключить светофоры"
// 	}
// 	return "Включить фазу №" + strconv.Itoa(toSend.Params)
// }
