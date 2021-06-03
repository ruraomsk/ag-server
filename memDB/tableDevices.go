package memDB

import (
	"encoding/json"
	"fmt"
	"github.com/ruraomsk/TLServer/logger"
	"github.com/ruraomsk/ag-server/pudge"
	"strconv"
	"time"
)

var TableDevices *Tx

func initDevices() {
	TableDevices = Create()
	TableDevices.name = "devices"
	TableDevices.ReadAll = func() map[string]interface{} {
		res := make(map[string]interface{})
		w := fmt.Sprintf("select device from devices;")
		rows, err := db.Query(w)
		if err != nil {
			logger.Error.Printf("запрос %s %s", w, err.Error())
			return res
		}
		var buffer []byte
		var ctrl pudge.Controller
		for rows.Next() {
			_ = rows.Scan(&buffer)
			_ = json.Unmarshal(buffer, &ctrl)
			key := strconv.Itoa(ctrl.ID)
			ctrl.StatusConnection = false
			ctrl.DK.EDK = 0
			ctrl.DK.PDK = false
			res[key] = ctrl
		}
		return res
	}
	TableDevices.UpdateFn = func(key string, value interface{}) string {
		ctrl := value.(pudge.Controller)
		val, _ := json.Marshal(value)
		return fmt.Sprintf("update devices set device='%s' where id=%d;", val, ctrl.ID)
	}
	TableDevices.AddFn = func(key string, value interface{}) string {
		ctrl := value.(pudge.Controller)
		val, _ := json.Marshal(value)
		return fmt.Sprintf("insert into devices (id, device) values (%d,'%s');", ctrl.ID, val)
	}
	TableDevices.DeleteFn = func(key string) string {
		id, _ := strconv.Atoi(key)
		return fmt.Sprintf("delete from devices where id=%d;", id)
	}
	return
}
func GetListControllers() []int {
	list := TableDevices.GetAllKeys()
	result := make([]int, 0)
	for _, id := range list {
		i, _ := strconv.Atoi(id)
		result = append(result, i)
	}
	return result
}
func GetController(id int) (pudge.Controller, error) {
	value, err := TableDevices.Get(strconv.Itoa(id))
	if err != nil {
		return pudge.Controller{}, err
	}
	return value.(pudge.Controller), nil
}
func SetController(controller pudge.Controller) {
	TableDevices.Set(strconv.Itoa(controller.ID), controller)
}
func NewController(cross pudge.Cross) pudge.Controller {
	c := pudge.Controller{ID: cross.IDevice}
	c.Name = cross.Name
	c.NK = 1
	c.PK = 1
	c.CK = 1
	c.LastOperation = time.Unix(0, 0)
	c.TechMode = 1
	c.DK = pudge.DK{TDK: 1}
	c.Base = true
	c.Statistics = make([]pudge.Statistic, 0)
	c.Arrays = make([]pudge.ArrayPriv, 0)
	c.LogLines = make([]pudge.LogLine, 0)
	return c
}
