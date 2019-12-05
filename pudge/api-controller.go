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
func IsRegistred(id int) bool {
	w := "select * from dev_gis where id=" + strconv.Itoa(id) + ";"
	rows, err := conDevGis.Query(w)
	if err != nil {
		logger.Error.Println(err.Error())
		return false
	}
	rows.Close()
	if rows.Next() {
		return true
	}
	return false
}
