package device

import "rura/ag-server/pudge"

import "time"

import "net"

import "rura/ag-server/setup"

import "strconv"

import "rura/ag-server/logger"

import "rura/ag-server/transport"

import "fmt"

import "sync"

import "rura/ag-server/extcon"

import "strings"

//Device управляющая структура имитатора устройства
type Device struct {
	ID           int
	Controller   *pudge.Controller
	Status       bool
	StatusDevice bool //true УСДК включено
	HeadDevice   transport.HeaderDevice
	HeadServer   transport.HeaderServer
	dk1          bool //Управление по ДК1
	dk2          bool //Управление по ДК2
	flag         bool //если изменили из GUI
	Soc          net.Conn
	mutex        sync.Mutex
	needAns      []int
	context      *extcon.ExtContext
}

//Close ТИПА ЗАКРЫВАЕМ УСТРОЙСТВО
func (d *Device) Close() {
	d.Status = false
	d.mutex.Unlock()
	defer d.Soc.Close()
}

//StartDevice обслуживание одного устройства
func (d *Device) StartDevice() {
	logger.Info.Printf("Запускаем id %d", d.ID)

	ctrl := new(pudge.Controller)
	ctrl.ID = d.ID
	d.Status = false
	d.StatusDevice = true
	setDefault(ctrl)
	soc, err := net.Dial("tcp", setup.Set.Controller.IP+":"+strconv.Itoa(setup.Set.Controller.Port))
	if err != nil {
		logger.Error.Printf("Ошибка соединения с портом %s", err.Error())
		return
	}
	defer d.Close()
	d.context, _ = extcon.NewContext("device" + strconv.Itoa(d.ID))
	d.Controller = ctrl
	d.Soc = soc
	err = d.writeFirstMessage()
	if err != nil {
		logger.Error.Printf("Ошибка  передачи %s", err.Error())
		return
	}
	err = d.readMessageServer()
	if err != nil {
		logger.Error.Printf("Ошибка  приема %s", err.Error())
		return
	}
	// Начинаем основной цикл
	for {
		is, err := d.readMaybeMessageFromServer()
		if err != nil {
			logger.Error.Printf("Ошибка возможного приема %s", err.Error())
			return
		}
		if is {
			d.updateDevice()
			if len(d.needAns) != 0 {
				err = d.makeAndSendAnsware()
				if err != nil {
					logger.Error.Printf("Ошибка передачи ответа от %d %s", d.ID, err.Error())
					return
				}
			}
		}
		d.context.SetTimeOut(time.Duration(1 * time.Second))
		select {
		case <-d.context.Done():
			if !strings.Contains(d.context.GetStatus(), "timeout") {
				logger.Info.Printf("id %d завершает работу...", d.ID)
				return
			}
			if time.Now().Sub(d.Controller.LastOperation) > setup.Set.Server.KeepAlive {
				err = d.sendKeepAlive()
				if err != nil {
					logger.Error.Printf("при передаче keepalive %s", err.Error())
					return
				}
				d.context.SetTimeOut(time.Duration(1 * time.Second))
				continue
			}
			// d.sendIfChange()
			d.context.SetTimeOut(time.Duration(1 * time.Second))
			continue
		}
	}
}

func setDefault(c *pudge.Controller) {
	c.LastOperation = time.Unix(0, 0)
	c.TexRezim = 1
	c.Base = true
	c.PK = 1
	c.CK = 1
	c.NK = 1
	var cc pudge.StatusCommandDU
	cc.IsPK = true
	cc.IsPKS = true
	cc.IsNK = true
	c.StatusCommandDU = cc
	var dk pudge.DK
	dk.RDK = 1
	dk.FDK = 1
	dk.DDK = 2
	dk.EDK = 0
	dk.PDK = false
	dk.EEDK = 0
	dk.ODK = false
	dk.LDK = 0
	dk.FTUDK = 1
	dk.TDK = 10
	dk.TTCDK = 20
	c.DK1 = dk
	c.DK2 = dk
	c.TMax = 0
	var m pudge.Model
	m.VPCPD = 101
	m.VPBS = 2
	m.C12 = true
	m.STP = true
	m.DKA = true
	m.DTA = true
	c.Model = m
	var er pudge.ErrorDevice
	er.V220DK1 = false
	er.V220DK2 = false
	er.RTC = false
	er.TVP1 = false
	er.TVP2 = false
	er.FRAM = false
	c.Error = er
	var gps pudge.GPS
	gps.Ok = true
	c.GPS = gps
	var input pudge.Input
	input.V1 = false
	c.Input = input
	c.Statistics = make([]pudge.Statistic, 0)
	c.Arrays = make([]pudge.ArrayPriv, 0)
	c.LogLines = make([]pudge.LogLine, 0)
}
func (d *Device) writeFirstMessage() error {
	code := 0
	if d.Controller.Base {
		code = 0xac
	}
	d.HeadDevice = transport.CreateHeaderDevice(d.Controller.ID, 30, 0, code)
	mss := make([]transport.SubMessage, 0)
	var ms transport.SubMessage
	ms.Set0x10Device(d.Controller)
	mss = append(mss, ms)
	d.HeadDevice.UpackMessages(mss)
	buffer := d.HeadDevice.MakeBuffer()
	n, err := d.Soc.Write(buffer)
	if err == nil && n != len(buffer) {
		err = fmt.Errorf("id %d передано %d байт вместо %d", d.Controller.ID, n, len(buffer))
	}
	d.Controller.LastOperation = time.Now()
	return err
}
func (d *Device) readMaybeMessageFromServer() (bool, error) {
	d.Soc.SetReadDeadline(time.Now().Add(setup.Set.Server.TimeOutRead))
	buf := make([]byte, 13)
	n, err := d.Soc.Read(buf)
	if strings.Contains(err.Error(), "i/o timeout") {
		return false, nil
	}
	if err == nil && n != len(buf) {
		err = fmt.Errorf("id %d при чтении сообщения от сервера прочитано %d байт нужно %d", d.ID, n, len(buf))
	}
	if err != nil {
		return false, err
	}
	buf2 := make([]byte, buf[12]+2)
	n, err = d.Soc.Read(buf2)
	if strings.Contains(err.Error(), "i/o timeout") {
		return false, nil
	}
	if err == nil && n != len(buf2) {
		err = fmt.Errorf("id %d при чтении сообщения от сервера прочитано %d байт нужно %d", d.ID, n, len(buf2))
	}
	if err != nil {
		return false, err
	}
	buffer := append(buf, buf2...)
	err = d.HeadServer.Parse(buffer)
	if err != nil {
		return true, fmt.Errorf("id %d при разборе  сообщения от сервера %s", d.ID, err.Error())
	}
	d.Controller.LastOperation = time.Now()
	return true, err
}

func (d *Device) readMessageServer() error {
	buf := make([]byte, 13)
	n, err := d.Soc.Read(buf)
	if err == nil && n != len(buf) {
		err = fmt.Errorf("id %d при чтении сообщения от сервера прочитано %d байт нужно %d", d.ID, n, len(buf))
	}
	if err != nil {
		return err
	}
	buf2 := make([]byte, buf[12]+2)
	n, err = d.Soc.Read(buf2)
	if err == nil && n != len(buf2) {
		err = fmt.Errorf("id %d при чтении сообщения от сервера прочитано %d байт нужно %d", d.ID, n, len(buf2))
	}
	if err != nil {
		return err
	}
	buffer := append(buf, buf2...)
	err = d.HeadServer.Parse(buffer)
	if err != nil {
		return fmt.Errorf("id %d при разборе  сообщения от сервера %s", d.ID, err.Error())
	}
	d.Controller.LastOperation = time.Now()
	return err
}
func (d *Device) updateDevice() {
	d.needAns = make([]int, 0)
	mss := d.HeadServer.ParseMessage()
	for _, ms := range mss {
		if ms.Type != 0 {
			//Прислали массив привязки
			num, array := ms.GetArray()
			flag := false
			for n, ar := range d.Controller.Arrays {
				if ar.Number == num {
					flag = true
					d.mutex.Lock()
					d.Controller.Arrays[n].Array = array
					d.mutex.Unlock()
				}
			}
			if !flag {
				var arr pudge.ArrayPriv
				arr.Number = num
				arr.Array = array
				d.mutex.Lock()
				d.Controller.Arrays = append(d.Controller.Arrays, arr)
				d.mutex.Unlock()
			}
			d.mutex.Lock()
			d.needAns = append(d.needAns, int(d.HeadServer.Number))
			d.mutex.Unlock()
			continue
		}
		switch ms.GetCodeCommandServer() {
		case 0x02:
			//Управление УСДК
			d.mutex.Lock()
			d.StatusDevice = ms.Get0x02Server()
			d.needAns = append(d.needAns, int(d.HeadServer.Number))
			d.mutex.Unlock()
		case 0x03:
			//Запрос состояния устройства
			d.mutex.Lock()
			d.needAns = append(d.needAns, -1)
			d.mutex.Unlock()
		case 0x04:
			//Запрос на смену фаз
			d.mutex.Lock()
			bb := ms.Get0x04Server()
			d.dk1 = bb[0]
			d.dk2 = bb[1]
			d.mutex.Unlock()
		case 0x05:
			//Запрос смена плана координации
			d.mutex.Lock()
			d.Controller.PK = ms.Get0x05Server()
			d.mutex.Unlock()
		case 0x06:
			//Смена суточной карты
			d.mutex.Lock()
			d.Controller.CK = ms.Get0x06Server()
			d.mutex.Unlock()
		case 0x07:
			//Смена недельной карты
			d.mutex.Lock()
			d.Controller.NK = ms.Get0x06Server()
			d.mutex.Unlock()
		case 0x09:
			//Режим работы ДК1
			d.mutex.Lock()
			d.Controller.DK1.RDK = ms.Get0x09Server()
			d.mutex.Unlock()
		case 0x0A:
			//Режим работы ДК2
			d.mutex.Lock()
			d.Controller.DK2.RDK = ms.Get0x0AServer()
			d.mutex.Unlock()
		case 0x0B:
			//Передача массива привязки
			d.mutex.Lock()
			ii := ms.Get0x0BServer()
			d.needAns = append(d.needAns, -2)
			d.needAns = append(d.needAns, ii[0])
			d.needAns = append(d.needAns, ii[1])
			d.mutex.Unlock()
		}
	}
}
func (d *Device) makeAndSendAnsware() error {
	if len(d.needAns) == 0 {
		return nil
	}
	code := 0
	if d.Controller.Base {
		code = 0xac
	}
	d.HeadDevice = transport.CreateHeaderDevice(d.Controller.ID, 30, 0, code)
	mss := make([]transport.SubMessage, 0)
	var ms transport.SubMessage
	d.mutex.Lock()
	for i := 0; i < len(d.needAns); {
		if d.needAns[i] > 0 {
			ms.Set0x01Device(d.needAns[i], time.Now().Minute(), time.Now().Second(), 0, 0)
			mss = append(mss, ms)
			i++
			continue
		}
		if d.needAns[i] == -1 {
			//Нужно ответить на запрос о состоянии устройства
			ms.Set0x0FDevice(d.Controller)
			mss = append(mss, ms)
			ms.Set0x12Device(d.Controller)
			mss = append(mss, ms)
			ms.Set0x11Device(d.Controller)
			mss = append(mss, ms)
			i++
			continue
		}
		if d.needAns[i] == -2 {
			i++
			for _, ar := range d.Controller.Arrays {
				if ar.Number == d.needAns[i] {
					ms.SetArray(ar.Number, ar.Array)
					break
				}
			}
			i += 2
			mss = append(mss, ms)
		}

	}
	d.mutex.Unlock()
	d.HeadDevice.UpackMessages(mss)
	buffer := d.HeadDevice.MakeBuffer()
	d.Soc.SetWriteDeadline(time.Now().Add(setup.Set.Server.TimeOutWrite))
	n, err := d.Soc.Write(buffer)
	if err == nil && n != len(buffer) {
		return fmt.Errorf("при отправке id %d передано %d байт вместо %d", d.ID, n, len(buffer))
	}
	if err != nil {
		return nil
	}
	d.Controller.LastOperation = time.Now()
	return nil
}
func (d *Device) sendKeepAlive() error {
	code := 0
	if d.Controller.Base {
		code = 0xac
	}
	d.mutex.Lock()
	d.HeadDevice = transport.CreateHeaderDevice(d.Controller.ID, 30, 0, code)
	mss := make([]transport.SubMessage, 0)
	var ms transport.SubMessage
	//Нужно ответить на запрос о состоянии устройства
	ms.Set0x0FDevice(d.Controller)
	mss = append(mss, ms)
	ms.Set0x12Device(d.Controller)
	mss = append(mss, ms)
	ms.Set0x11Device(d.Controller)
	mss = append(mss, ms)
	d.HeadDevice.UpackMessages(mss)
	buffer := d.HeadDevice.MakeBuffer()
	d.mutex.Unlock()
	d.Soc.SetWriteDeadline(time.Now().Add(setup.Set.Server.TimeOutWrite))
	n, err := d.Soc.Write(buffer)
	if err == nil && n != len(buffer) {
		return fmt.Errorf("при keepAlive id %d передано %d байт вместо %d", d.ID, n, len(buffer))
	}
	if err != nil {
		return nil
	}
	d.Controller.LastOperation = time.Now()
	return nil
}
