package memDB

import (
	"encoding/json"
	"fmt"
	"github.com/ruraomsk/TLServer/logger"
	"github.com/ruraomsk/ag-server/binding"
	"github.com/ruraomsk/ag-server/pudge"
)

var CrossesTable *Tx

func initCrosses() {
	CrossesTable = Create()
	CrossesTable.name = "crosses"
	CrossesTable.ReadAll = func() map[string]interface{} {
		res := make(map[string]interface{})
		w := fmt.Sprintf("select state from public.\"cross\";")
		rows, err := db.Query(w)
		if err != nil {
			logger.Error.Printf("запрос %s %s", w, err.Error())
			return res
		}
		var buffer []byte
		var cross pudge.Cross
		for rows.Next() {
			_ = rows.Scan(&buffer)
			_ = json.Unmarshal(buffer, &cross)
			cross.StatusDevice = 18
			key := pudge.Region{Region: cross.Region, Area: cross.Area, ID: cross.ID}
			res[key.ToKey()] = cross
		}
		return res
	}
	CrossesTable.UpdateFn = func(key string, value interface{}) string {
		state := value.(pudge.Cross)
		val, _ := json.Marshal(value)
		return fmt.Sprintf("update public.\"cross\" set subarea=%d,idevice=%d,dgis='%s',describ='%s',status='%d',state='%s' "+
			"where region=%d and area=%d and id=%d;", state.SubArea, state.IDevice, state.Dgis, state.Name, state.StatusDevice, val,
			state.Region, state.Area, state.ID)
	}
	CrossesTable.AddFn = func(key string, value interface{}) string {
		state := value.(pudge.Cross)
		val, _ := json.Marshal(value)
		return fmt.Sprintf("insert into public.\"cross\" (region, area, subarea, id, idevice, dgis, describ, status, state) values "+
			"(%d,%d,%d,%d,%d,'%s','%s',%d,'%s');",
			state.Region, state.Area, state.ID, state.SubArea, state.IDevice, state.Dgis, state.Name, state.StatusDevice, val)
	}
	CrossesTable.DeleteFn = func(key string) string {
		reg := pudge.FromKeyToRegion(key)
		return fmt.Sprintf("delete from public.\"cross\" where region=%d and area=%d and id=%d;", reg.Region, reg.Area, reg.ID)
	}
	return
}
func GetCrossFind(region, area, id int) (pudge.Cross, error) {
	reg := pudge.Region{Region: region, Area: area, ID: id}
	value, err := CrossesTable.Get(reg.ToKey())
	if err != nil {
		return pudge.Cross{}, err
	}
	return value.(pudge.Cross), err

}
func GetCross(key string) (pudge.Cross, error) {
	value, err := CrossesTable.Get(key)
	if err != nil {
		return pudge.Cross{}, err
	}
	return value.(pudge.Cross), err
}
func SetCross(cross pudge.Cross) {
	reg := pudge.Region{Region: cross.Region, Area: cross.Area, ID: cross.ID}
	CrossesTable.Set(reg.ToKey(), cross)
}
func IsDeviceSetting(device int) bool {
	for _, value := range CrossesTable.MDB.Data {
		cross := value.(pudge.Cross)
		if device == cross.IDevice {
			return true
		}
	}
	return false
}
func GetCrossFromDevice(device int) (pudge.Cross, error) {
	CrossesTable.Lock()
	defer CrossesTable.Unlock()
	for _, value := range CrossesTable.MDB.Data {
		cross := value.(pudge.Cross)
		if device == cross.IDevice {
			return cross, nil
		}
	}
	return pudge.Cross{}, fmt.Errorf("нет связанного перекрестка с %d", device)
}
func NewCross(region, area, id int) pudge.Cross {
	cross := pudge.Cross{Region: region, Area: area, ID: id}
	cross.Arrays = *binding.NewArrays()
	return cross
}
func DeleteCross(region, area, id int) {
	reg := pudge.Region{Region: region, Area: area, ID: id}
	CrossesTable.Delete(reg.ToKey())
}
