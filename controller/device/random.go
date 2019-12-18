package device

import (
	"math/rand"
	"rura/ag-server/setup"
)

//В этом месте производим разные измения состояния устройства
func (d *Device) randomChange() bool {
	if !setup.Set.Controller.Random {
		return false
	}

	switch rand.Intn(4) {
	case 0:
		//Ничего не правим
		return false
	case 1:
		// Неисправность входов
		return d.Controller.Input.MakeError()
	case 2:
		// Неисправность ошибок
		return d.Controller.Error.MakeError()
	case 3:
		// Неисправность GPS
		return d.Controller.GPS.MakeError()

	}
	return false
}
