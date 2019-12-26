package device

import (
	"encoding/hex"
	"math/rand"
	"net"
	"rura/ag-server/extcon"
	"rura/ag-server/logger"
	"rura/ag-server/pudge"
	"rura/ag-server/setup"
	"rura/ag-server/transport"
	"strconv"
	"sync"
	"time"
)

//Devs список всех устройств
var Devs map[int]*Device

//LogInt одна запись внутреннего лога обменов
type LogInt struct {
	Time    time.Time
	Source  bool //true если это сообщение устройства иначе сообщение сервера
	Message []byte
}

//Device управляющая структура имитатора устройства
type Device struct {
	ID           int
	Name         string
	Controller   *pudge.Controller
	Status       bool
	StatusDevice bool //true УСДК включено
	HeadDevice   transport.HeaderDevice
	HeadServer   transport.HeaderServer
	dk1          bool //Управление по ДК1
	dk2          bool //Управление по ДК2
	flag         bool //если изменили из GUI
	Soc          net.Conn
	Mutex        sync.Mutex
	needAns      []int
	context      *extcon.ExtContext
	LogInts      []LogInt
	Random       bool
	hout         chan transport.HeaderDevice
	hin          chan transport.HeaderServer
}

//LogToList вывод лога для GUI
func (d *Device) LogToList() []string {
	result := make([]string, 0)
	for _, ll := range d.LogInts {
		t := ll.Time.Format("15:04:05")
		d := "s"
		if ll.Source {
			d = "d"
		}
		message := hex.EncodeToString(ll.Message)
		m := ""
		for i := 0; i < len(message); i++ {
			m += message[i : i+1]
			if i%2 == 1 {
				m += " "
			}
		}
		r := m + ";" + t + d
		result = append(result, r)

	}
	return result
}

//ToList Вывод для GUI
func (d *Device) ToList() []string {
	result := make([]string, 0)
	r := "На линии;" + strconv.FormatBool(d.Status)
	result = append(result, r)
	r = "УСДК включен;" + strconv.FormatBool(d.StatusDevice)
	result = append(result, r)
	r = "Управление по ДК1;" + strconv.FormatBool(d.dk1)
	result = append(result, r)
	r = "Управление по ДК2;" + strconv.FormatBool(d.dk2)
	result = append(result, r)
	return result
}
func (d *Device) addLog(source bool, buffer []byte) {
	l := new(LogInt)
	l.Time = time.Now()
	l.Source = source
	l.Message = buffer
	d.Mutex.Lock()
	d.LogInts = append(d.LogInts, *l)
	d.Mutex.Unlock()

}

//Close ТИПА ЗАКРЫВАЕМ УСТРОЙСТВО
func (d *Device) Close() {
	d.Status = false

	// d.Mutex.Unlock()
	d.Soc.Close()
}

//StartDevice обслуживание одного устройства
func (d *Device) StartDevice() {
	// logger.Info.Printf("Запускаем id %d", d.ID)
	rand.Seed(int64(d.ID))
	time.Sleep(time.Duration(rand.Intn(60)+1) * time.Second)
	d.Status = false
	d.StatusDevice = true
	d.LogInts = make([]LogInt, 0)
	soc, err := net.Dial("tcp", setup.Set.Controller.IP+":"+strconv.Itoa(setup.Set.Controller.Port))
	if err != nil {
		logger.Error.Printf("Ошибка соединения с портом %s", err.Error())
		return
	}
	d.hout = make(chan transport.HeaderDevice)
	d.hin = make(chan transport.HeaderServer)
	defer close(d.hout)
	defer close(d.hin)
	defer d.Close()
	go transport.GetMessagesFromService(soc, d.hin)
	go transport.SendMessagesToServer(soc, d.hout)
	d.Soc = soc

	d.writeFirstMessage()
	d.Status = true
	// Начинаем основной цикл
	d.Mutex.Lock()
	d.context, _ = extcon.NewContext("device" + strconv.Itoa(d.ID))
	d.Mutex.Unlock()
	timer := extcon.SetTimerClock(time.Duration(time.Duration(setup.Set.Controller.Step) * time.Second))
	for {
		d.Mutex.Lock()
		if d.context == nil {
			logger.Error.Printf("Пропал контекст %d ", d.ID)
			d.context, _ = extcon.NewContext("device" + strconv.Itoa(d.ID))
		}
		d.Mutex.Unlock()
		select {
		case d.HeadServer = <-d.hin:
			buffer := d.HeadServer.MakeBuffer()
			d.addLog(false, buffer)
			d.Controller.LastOperation = time.Now()
			d.updateDevice()
			if len(d.needAns) != 0 {
				err = d.makeAndSendAnsware()
				if err != nil {
					logger.Error.Printf("Ошибка передачи ответа от %d %s", d.ID, err.Error())
					return
				}
			}
		case <-timer.C:
			if time.Now().Sub(d.Controller.LastOperation) > setup.Set.Server.KeepAlive {
				//Возможно сделать несиправности
				if d.randomChange() {
					// logger.Info.Println("Изменено устройство ", d.ID)
				}
				err = d.sendKeepAlive()
				if err != nil {
					logger.Error.Printf("при передаче keepalive %s", err.Error())
					return
				}
			}
		case <-d.context.Done():
			logger.Info.Printf("id %d завершает работу...", d.ID)
			return
		}
	}
}

func (d *Device) writeFirstMessage() {
	code := 0
	if d.Controller.Base {
		code = 0xac
	}
	d.HeadDevice = transport.CreateHeaderDevice(d.Controller.ID, 30, 0, code)
	mss := make([]transport.SubMessage, 0)
	var ms transport.SubMessage
	ms.Set0x1DDevice(d.Controller)
	mss = append(mss, ms)
	d.HeadDevice.UpackMessages(mss)
	buffer := d.HeadDevice.MakeBuffer()
	d.hout <- d.HeadDevice
	d.addLog(true, buffer)
	d.Controller.LastOperation = time.Now()
	return
}

func (d *Device) updateDevice() {
	// d.Mutex.Lock()
	// defer d.Mutex.Unlock()
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
					d.Controller.Arrays[n].Array = array
				}
			}
			if !flag {
				var arr pudge.ArrayPriv
				arr.Number = num
				arr.Array = array
				d.Controller.Arrays = append(d.Controller.Arrays, arr)
			}
			d.needAns = append(d.needAns, int(d.HeadServer.Number))
			continue
		}
		switch ms.GetCodeCommandServer() {
		case 0x02:
			//Управление УСДК
			d.StatusDevice = ms.Get0x02Server()
			d.needAns = append(d.needAns, int(d.HeadServer.Number))
		case 0x03:
			//Запрос состояния устройства
			d.needAns = append(d.needAns, -1)
		case 0x04:
			//Запрос на смену фаз
			bb := ms.Get0x04Server()
			d.dk1 = bb[0]
			d.dk2 = bb[1]
		case 0x05:
			//Запрос смена плана координации
			d.Controller.PK = ms.Get0x05Server()
		case 0x06:
			//Смена суточной карты
			d.Controller.CK = ms.Get0x06Server()
		case 0x07:
			//Смена недельной карты
			d.Controller.NK = ms.Get0x06Server()
		case 0x09:
			//Режим работы ДК1
			d.Controller.DK1.RDK = ms.Get0x09Server()
		case 0x0A:
			//Режим работы ДК2
			d.Controller.DK2.RDK = ms.Get0x0AServer()
		case 0x0B:
			//Передача массива привязки
			ii := ms.Get0x0BServer()
			d.needAns = append(d.needAns, -2)
			d.needAns = append(d.needAns, ii[0])
			d.needAns = append(d.needAns, ii[1])
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
	d.HeadDevice.UpackMessages(mss)
	buffer := d.HeadDevice.MakeBuffer()
	d.addLog(true, buffer)
	d.Controller.LastOperation = time.Now()
	d.hout <- d.HeadDevice
	return nil
}
func (d *Device) sendKeepAlive() error {
	// d.Mutex.Lock()
	// defer d.Mutex.Unlock()
	code := 0
	if d.Controller.Base {
		code = 0xac
	}
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
	d.hout <- d.HeadDevice
	d.addLog(true, buffer)
	d.Controller.LastOperation = time.Now()
	return nil
}
