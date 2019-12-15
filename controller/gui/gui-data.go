package gui

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"rura/ag-server/controller/device"
	"sort"
	"time"
)

//List список всех котроллеров
type List struct {
	Devices []OneDevice `json:"devs"`
}

//OneDevice краткое описание одного устройства
type OneDevice struct {
	ID            int       `json:"id"`
	Connection    bool      `json:"con"`
	Name          string    `json:"name"`
	LastOperation time.Time `json:"ltime"`
}

//Logs Возвращает логи устройства
type Logs struct {
	Logs []OneLog `json:"ls"`
}

//OneLog одна строка лога
type OneLog struct {
	Time      time.Time `json:"t"`
	Direction string    `json:"d"`
	Message   string    `json:"m"`
}

//GetList возращает json list
func getList() ([]byte, error) {
	var list List
	var one OneDevice
	list.Devices = make([]OneDevice, 0)
	for _, d := range device.Devs {
		// d.Mutex.Lock()
		one.ID = d.Controller.ID
		one.Connection = d.Status
		one.Name = d.Controller.Name
		one.LastOperation = d.Controller.LastOperation
		list.Devices = append(list.Devices, one)
		// d.Mutex.Unlock()
	}
	sort.Slice(list.Devices, func(i, j int) bool { return list.Devices[i].ID < list.Devices[j].ID })
	return json.MarshalIndent(&list, "", "   ")
}
func getLog(id int) ([]byte, error) {
	var logs Logs
	var l OneLog
	result := make([]byte, 0)
	logs.Logs = make([]OneLog, 0)
	d, is := device.Devs[id]
	if !is {
		return result, fmt.Errorf("нет такого устройства %d", id)
	}
	for _, ll := range d.LogInts {
		l.Time = ll.Time
		l.Direction = "from server "
		if ll.Source {
			l.Direction = "from device "
		}
		message := hex.EncodeToString(ll.Message)
		l.Message = ""
		for i := 0; i < len(message); i++ {
			l.Message += message[i : i+1]
			if i%2 == 1 {
				l.Message += " "
			}
		}

		logs.Logs = append(logs.Logs, l)
	}
	return json.MarshalIndent(&logs, "", "   ")
}
