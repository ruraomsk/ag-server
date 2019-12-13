package pudge

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

func isRegistred(id int) string {
	mutex.Lock()
	defer mutex.Unlock()
	name, is := ids[id]
	if is {
		return name
	}
	return ""
}
