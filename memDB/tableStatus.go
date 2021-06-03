package memDB

import (
	"fmt"
	"github.com/ruraomsk/TLServer/logger"
	"github.com/ruraomsk/ag-server/pudge"
	"strconv"
)

var StatusTable *Tx

func initStatus() {
	StatusTable = Create()
	StatusTable.writable = false
	StatusTable.name = "status"
	StatusTable.ReadAll = func() map[string]interface{} {
		res := make(map[string]interface{})
		w := fmt.Sprintf("select id,description,control from status;")
		rows, err := db.Query(w)
		if err != nil {
			logger.Error.Printf("запрос %s %s", w, err.Error())
			return res
		}
		var id int
		var desc string
		var control bool
		for rows.Next() {
			_ = rows.Scan(&id, &desc, &control)
			key := strconv.Itoa(id)
			res[key] = pudge.StatusCtrl{ID: id, Description: desc, Control: control}
		}
		return res
	}
}
func GetStatus(id int) string {
	value, err := StatusTable.Get(strconv.Itoa(id))
	if err != nil {
		return ""
	}
	return value.(pudge.StatusCtrl).Description
}
func GetControls(id int) bool {
	value, err := StatusTable.Get(strconv.Itoa(id))
	if err != nil {
		return false
	}
	return value.(pudge.StatusCtrl).Control
}
