package gui

import (
	"encoding/json"
	"rura/ag-server/controller/device"
	"sort"
)

//List список всех котроллеров
type List struct {
	Devices []OneDevice `json:"devs"`
}

//OneDevice краткое описание одного устройства
type OneDevice struct {
	ID         int  `json:"id"`
	Connection bool `json:"con"`
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
		list.Devices = append(list.Devices, one)
		// d.Mutex.Unlock()
	}
	sort.Slice(list.Devices, func(i, j int) bool { return list.Devices[i].ID < list.Devices[j].ID })
	return json.MarshalIndent(&list, "", "   ")
}
