package api

import (
	"bufio"
	"crypto/md5"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/JanFant/TLServer/logger"
	"github.com/ruraomsk/ag-server/pudge"
	"github.com/ruraomsk/ag-server/setup"
	"hash"
	"io"
	"net"
	"reflect"
	"strconv"
)

var users map[string]int

type ConnectQuery struct {
	Login    string `json:"login"`
	Password string `json:"password"`
	Region   int    `json:"region"`
}
type Request struct {
	//MagicWord   string `json:"magic"`
	PortCommand int `json:"portc"`
	PortOut     int `json:"portout"`
}
type CommandOut struct {
	Login string `json:"login"`
	//Command string `json:"command"` //command start or next
}

func listenApiCommand() {
	ln, err := net.Listen("tcp", ":"+strconv.Itoa(setup.Set.ApiServer.PortAPI))
	if err != nil {
		logger.Error.Printf("Ошибка открытия порта %s", err.Error())
		return
	}
	defer ln.Close()
	for {
		socket, err := ln.Accept()
		if err != nil {
			logger.Error.Printf("Ошибка accept %s", err.Error())
			continue
		}
		go workerConnect(socket)
	}
}
func listenOutCommand() {
	ln, err := net.Listen("tcp", ":"+strconv.Itoa(setup.Set.ApiServer.PortOut))
	if err != nil {
		logger.Error.Printf("Ошибка открытия порта %s", err.Error())
		return
	}
	defer ln.Close()
	for {
		socket, err := ln.Accept()
		if err != nil {
			logger.Error.Printf("Ошибка accept %s", err.Error())
			continue
		}
		go workerOutConnect(socket)
	}
}

//APIState структура для возврата состояния перкрестков
type APIState struct {
	Region       int    `json:"region"`  //Регион
	Area         int    `json:"area"`    //Район
	SubArea      int    `json:"subarea"` //подрайон
	ID           int    `json:"id"`      //Номер перекрестка
	IDevice      int    `json:"idevice"` // Назначенное на перекресток устройство
	Dgis         string `json:"dgis"`    //Координаты перекрестка
	Name         string `json:"name"`
	StatusDevice int    `json:"status"` // Статус устройства
	PK           int    `json:"pk"`     //Номер плана координации
	CK           int    `json:"ck"`     //Номер суточной карты
	NK           int    `json:"nk"`     //Номер недельной карты
}
type table struct {
	records map[string]*record
	sending []string
}

type record struct {
	hash hash.Hash
}

func (t *table) newCycle() {
	t.sending = make([]string, 0)
}
func (t *table) key(state pudge.Cross) string {
	return fmt.Sprintf("%d;%d;%d", state.Region, state.Area, state.ID)
}
func (t *table) addRecord(state pudge.Cross) {
	rapi := new(APIState)
	rapi.newValue(state)
	b, err := json.Marshal(&rapi)
	if err != nil {
		logger.Error.Printf("Преобразование %v ошибка %s", rapi, err.Error())
		return
	}
	str := string(b) + "\n"
	hnew := md5.New()
	_, _ = io.WriteString(hnew, str)
	r, is := t.records[t.key(state)]
	if is {
		if !reflect.DeepEqual(&r.hash, &hnew) {
			t.sending = append(t.sending, str)
			r.hash = hnew
		}
	} else {
		r := new(record)
		t.sending = append(t.sending, str)
		r.hash = hnew
		t.records[t.key(state)] = r
	}
}
func (a *APIState) newValue(s pudge.Cross) {
	a.Region = s.Region
	a.Area = s.Area
	a.ID = s.ID
	a.Dgis = s.Dgis
	a.Name = s.Name
	a.NK = s.NK
	a.CK = s.CK
	a.PK = s.PK
	a.StatusDevice = s.StatusDevice
}
func workerOutConnect(soc net.Conn) {
	defer soc.Close()
	var command CommandOut
	tab := new(table)
	logger.Info.Printf("Новый клиент комманд обновления %s", soc.RemoteAddr().String())
	dbinfo := fmt.Sprintf("host=%s user=%s password=%s dbname=%s sslmode=disable",
		setup.Set.DataBase.Host, setup.Set.DataBase.User,
		setup.Set.DataBase.Password, setup.Set.DataBase.DBname)
	dbb, err := sql.Open("postgres", dbinfo)
	if err != nil {
		logger.Error.Printf("Запрос на открытие %s %s", dbinfo, err.Error())
		return
	}
	defer dbb.Close()
	if err = dbb.Ping(); err != nil {
		logger.Error.Printf("Ping %s", err.Error())
		return
	}
	reader := bufio.NewReader(soc)
	writer := bufio.NewWriter(soc)
	for {
		c, err := reader.ReadString('\n')
		if err != nil {
			logger.Error.Println("При чтении команд сервера обновлений ", err.Error())
			return
		}
		if c[0:1] == "0" {
			// logger.Info.Println("Keep alive")
			continue
		}
		err = json.Unmarshal([]byte(c), &command)
		if err != nil {
			logger.Error.Println("При конвератации команд сервера обновлений ", err.Error())
			continue
		}
		region, is := users[command.Login]
		if !is {
			logger.Error.Printf("Нет такого пользователя %s ", command.Login)
			return
		}
		tab.newCycle()
		rows, err := dbb.Query(fmt.Sprintf("select state from public.cross where region=%d;", region))
		if err != nil {
			logger.Error.Printf("Error %s", err.Error())
			return
		}
		for rows.Next() {
			var state pudge.Cross
			_ = rows.Scan(&state)
			tab.addRecord(state)
		}
		_ = rows.Close()
		for _, w := range tab.sending {
			_, _ = writer.WriteString(w)
		}
		_, _ = writer.WriteString(emptyCross())
	}
}
func emptyCross() string {
	r := new(APIState)
	b, _ := json.Marshal(&r)
	return string(b) + "\n"
}
func workerConnect(soc net.Conn) {
	defer soc.Close()
	var connect ConnectQuery
	logger.Info.Printf("Новый клиент комманд %s", soc.RemoteAddr().String())
	reader := bufio.NewReader(soc)
	c, err := reader.ReadString('\n')
	if err != nil {
		logger.Error.Println("При чтении подключения к серверу API ", err.Error())
		return
	}
	err = json.Unmarshal([]byte(c), &connect)
	if err != nil {
		logger.Error.Println("При конвератации подключения к серверу API ", err.Error())
		return
	}
	logger.Info.Printf("Подключился к API %s ", connect.Login)
	users[connect.Login] = connect.Region
	req := Request{PortCommand: setup.Set.CommServer.PortCommand, PortOut: setup.Set.ApiServer.PortOut}
	res, err := json.Marshal(&req)
	if err != nil {
		logger.Error.Println("При создании ответа от сервера API ", err.Error())
		return
	}
	_, _ = soc.Write(res)
	_, _ = soc.Write([]byte("\n"))
}
func Start() {
	logger.Info.Print("Сервер команд API ")
	users = make(map[string]int)
	go listenApiCommand()
	go listenOutCommand()
}
