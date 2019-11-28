package pudge

//CreateController создает устройство на основе справочников
func CreateController(id int, typeDevice int) (Controller, error) {
	var dev Controller
	dev.ID = id
	dev.StatusConnection = NotConnected
	dev.StatusDevice = 0
	// Потом дописать настройку координат и прочее
	return dev, nil
}
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
