package device

import (
	"github.com/ruraomsk/ag-server/setup"
	"math/rand"
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

//setDKs производим контроль перехода на автоматическое управление
func (d *Device) setDKs() {
	if d.Controller.Base {
		//работаем в базовом режиме нужно проверить полноту настроек массивов
		d.makeAreas()
		if !d.isFullAreas() {
			return
		}
		d.Controller.Base = false
	}

}
func (d *Device) makeAreas() {
	// for _,ar:=range d.Controller.Arrays{
	// 	if
	// }
	return
}
func (d *Device) isFullAreas() bool {

	return false
}
