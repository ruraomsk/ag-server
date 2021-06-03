package techComm

import (
	"github.com/ruraomsk/ag-server/extcon"
	"github.com/ruraomsk/ag-server/pudge"
	"github.com/ruraomsk/ag-server/transport"
	"net"
	"time"
)

type CommandARM struct {
	ID      int    `json:"id"`
	User    string `json:"user"`
	Command int    `json:"cmd"`
	Params  int    `json:"param"`
}

//CommandArray Привязка от Сервера АРМ
type CommandArray struct {
	ID     int   `json:"id"`
	Number int   `json:"number"`
	NElem  int   `json:"nelem"`
	Elems  []int `json:"elems"`
}

//ChangeProtocol для изменения протокола
type ChangeProtocol struct {
	ID       int    `json:"id"` // Уникальный номер контроллера
	User     string `json:"user"`
	F0x32    bool   `json:"f0x32"`    //Есть команда смены IP сервера
	IP       string `json:"ip"`       // Собственно IP в формате 000.000.000.000
	Port     int    `json:"port"`     // Номер порта сервера
	F0x33    bool   `json:"f0x33"`    //Есть команда смены интервала контроля  сервера
	Long     int    `json:"long"`     //Новый интервал в минутах
	F0x34    bool   `json:"f0x34"`    //Есть команда смены режима обмена
	Type     bool   `json:"type"`     // True  - Экономичный режим false - стандартный
	F0x35    bool   `json:"f0x35"`    //Есть команда смены интервала ожидания ответа
	Interval int    `json:"interval"` //Новый	 интервал в минутах
	Ignore   bool   `json:"ignore"`   //True игнорировать команду разрыва от ПСПД
}

type Device struct {
	ID             int //Идентификатор устройства
	Work           bool
	Socket         net.Conn               //Сокет для обмена с устройством
	Context        *extcon.ExtContext     //Расширенный контекст для управления устройством
	NumDev         uint8                  //Номер сообщения для подтверждения
	NumServ        uint8                  //Номер сообщения от сервера
	WaitNum        uint8                  //Номер ожидаемого сообщения
	CommandARM     chan CommandARM        //Канал приема команд управления
	CommandArray   chan []pudge.ArrayPriv //Канал приема привязок
	ChangeProtocol chan ChangeProtocol    //Канал приема изменения протокола обмена
	ErrorTCP       chan int
	Messages       DequeServer
	LastMessage    transport.HeaderServer
	LastToDevice   time.Time
	CountLost      int                         //Счетчик ожиданий ответа на номер
	tOut           time.Duration               //Тайм аут на передачу
	tIn            time.Duration               //Тайм аут на прием
	hOut           chan transport.HeaderServer //Канал для передачи на устройство
	hIn            chan transport.HeaderDevice //Канал приема от устройства
	Traffic        pudge.Traffic
}

func (d *Device) addNumber() {
	d.NumServ++
	if d.NumServ >= 250 {
		d.NumServ = 1
	}
}

type DequeServer struct {
	array []transport.HeaderServer
	size  int
}

func (d *DequeServer) Push(value transport.HeaderServer) {
	if d.size == 0 {
		d.array = make([]transport.HeaderServer, 0)
	}
	d.array = append(d.array, value)
	d.size = len(d.array)
}
func (d *DequeServer) Pop() transport.HeaderServer {
	if d.size == 0 {
		d.array = make([]transport.HeaderServer, 0)
		return transport.HeaderServer{}
	}
	r := d.array[0]
	d.array = d.array[1:]
	d.size = len(d.array)
	return r
}
func (d *DequeServer) Size() int {
	if d.size == 0 {
		d.array = make([]transport.HeaderServer, 0)
	}
	return d.size
}
