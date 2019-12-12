package pudge

import "strconv"

import "rura/ag-server/logger"

func (c *Controller) isChanged(cc Controller) bool {
	if c.StatusConnection != cc.StatusConnection {
		return true
	}
	if c.StatusDevice != cc.StatusDevice {
		return true
	}

	return false
}

//IsConnected возвращает на связи ли устройство
func (c *Controller) IsConnected() bool {
	return c.StatusConnection != Connected
}

//IsRegistred возвращает истину если данный id зарегистрирован
func IsRegistred(id int) (bool, string) {
	mutex.Lock()
	defer mutex.Unlock()
	w := "select region,id from public.\"cross\" where idevice=" + strconv.Itoa(id) + ";"
	rows, err := conDevGis.Query(w)
	if err != nil {
		logger.Error.Println(err.Error())
		return false, ""
	}
	var region int
	var idr int
	defer rows.Close()
	if rows.NextResultSet() {
		rows.Next()
		rows.Scan(&region, &idr)
		w := "select describ from public.dev_gis where region=" + strconv.Itoa(region) + " and id=" + strconv.Itoa(idr) + ";"
		rs, err := conDevGis.Query(w)
		if err != nil {
			logger.Error.Println(err.Error())
			return false, ""
		}
		var res string
		rs.Next()
		rs.Scan(&res)
		return true, res
	}
	return false, ""
}
