package techComm

import (
	"fmt"
	"github.com/ruraomsk/ag-server/extcon"
	"github.com/ruraomsk/ag-server/memDB"
	"github.com/ruraomsk/ag-server/pudge"
	"github.com/ruraomsk/ag-server/transport"
	"net"
	"sync"
)

var devices = make(map[int]*Device)
var muDevice sync.Mutex

func getNameContext(id int) string {
	return fmt.Sprintf("device%d", id)
}

//GetChanArray возвращает канал для присылки массивов для данного устройства
func GetChanArray(id int) (chan []pudge.ArrayPriv, error) {
	if !isDeviceWork(id) {
		err := fmt.Errorf("нет канала слать массив на %d", id)
		return nil, err
	}
	d := getDevice(id)
	return d.CommandArray, nil
}
func getChanCommand(id int) (chan CommandARM, error) {
	if !isDeviceWork(id) {
		err := fmt.Errorf("нет канала слать команды на %d", id)
		return nil, err
	}
	d := getDevice(id)
	return d.CommandARM, nil
}
func isDeviceWork(id int) bool {
	muDevice.Lock()
	defer muDevice.Unlock()
	dev, is := devices[id]
	if !is {
		//logger.Info.Printf("Устройство %d не на связи", id)
		return false
	}
	return dev.Work
}
func stopDevice(id int) {
	extcon.StopForName(getNameContext(id))
}
func getChanProtocol(id int) (chan ChangeProtocol, error) {
	if !isDeviceWork(id) {
		err := fmt.Errorf("нет канала слать  протокол на %d", id)
		d := make(chan ChangeProtocol)
		return d, err
	}
	d := getDevice(id)
	return d.ChangeProtocol, nil
}
func getController(id int) (pudge.Controller, error) {
	cross, err := memDB.GetCrossFromDevice(id)
	if err != nil {
		return pudge.Controller{}, err
	}
	ctrl, err := memDB.GetController(id)
	if err != nil {
		//Не было раньше создаем запись
		ctrl = memDB.NewController(cross)
	}
	if ctrl.Name != cross.Name {
		ctrl.Name = cross.Name
	}
	return ctrl, nil
}
func newDevice(c pudge.Controller, socket net.Conn) *Device {
	d := new(Device)
	d.ID = c.ID
	d.Context, _ = extcon.NewContext(getNameContext(d.ID))
	d.CommandARM = make(chan CommandARM)
	d.CommandArray = make(chan []pudge.ArrayPriv)
	d.ChangeProtocol = make(chan ChangeProtocol)
	d.ErrorTCP = make(chan int)
	d.hOut = make(chan transport.HeaderServer)
	d.hIn = make(chan transport.HeaderDevice)
	d.Socket = socket
	return d
}
func addDevice(device *Device) {
	muDevice.Lock()
	devices[device.ID] = device
	muDevice.Unlock()
}
func deleteDevice(id int) {
	muDevice.Lock()
	delete(devices, id)
	muDevice.Unlock()

}
func getDevice(id int) *Device {
	return devices[id]
}
