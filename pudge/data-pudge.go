// Package pudge выполняет
//  1. Ведется текущее состояние контроллеров
//  2. Сидит прием по каналу запросов на чтение со стороны сервера АРМ
//  3. Открывается канал приема запросов на запись от сервера коммуникации
//  4. Если сервер коммуникации присылает запрос на запись нового состояния то
//     делается проверка на существенное измение и если это так то новое состояние посылается в бд логгирования
//  5. Открывается прием по каналу запросов от сервера АРМ после отправки копиии запроса в канал сервера канала данный
//     запрос посылается серверу коомуникации
//  6. По времени заданному в настройках делается полная копия состояния всех контроллеров в базу данных простой посылкой
//     копии
package pudge

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/ruraomsk/ag-server/binding"
	"github.com/ruraomsk/ag-server/logger"
	"github.com/ruraomsk/ag-server/setup"
)

// CommandARM Команды от Сервера АРМ
type CommandARM struct {
	ID      int    `json:"id"`
	User    string `json:"user"`
	Command int    `json:"cmd"`
	Params  int    `json:"param"`
}
type CommandXT struct {
	Region  Region
	Command int
}

// StatusCtrl //Описывает статус устройства
type StatusCtrl struct {
	ID          int    `json:"id"`
	Description string `json:"description"`
	Control     bool   `json:"control"`
}

// Region указатель на номер перекрестка
type Region struct {
	Region int //Код региона
	Area   int //Код района
	ID     int //Номер перекрестка
}

// ToKey создает строковый ключ
func (r *Region) ToKey() string {
	return fmt.Sprintf("%d:%d:%d", r.Region, r.Area, r.ID)
}
func FromKeyToRegion(key string) Region {
	r := strings.Split(key, ":")
	region, _ := strconv.Atoi(r[0])
	area, _ := strconv.Atoi(r[1])
	id, _ := strconv.Atoi(r[2])
	return Region{Region: region, Area: area, ID: id}
}
func (r *Region) LocalTime() time.Time {
	for _, v := range setup.Set.XCtrl.Regions {
		if v[0] == r.Region {
			return time.Now().Add(time.Duration(v[1]) * time.Hour)
		}
	}
	return time.Now()
}

// StatusConnection статус соединения
type StatusConnection int

// DK диагностика состояния по ДК
type DK struct {
	// 1 2 Ручное управление
	// 3 Зеленая улица
	// 4 Диспетчерское управление
	// 5 6 Локальное управление
	// 8 9 Координированное управление
	RDK int `json:"rdk"` //Режим ДК
	// от 1 до 8 номера рабочих фаз
	// 9 	13 промежуточный такт
	// 10 	14  желтое мигание
	// 11 	15 отключен светофор
	// 12 кругом краснный
	//
	FDK int `json:"fdk"` //Фаза ДК
	//	1 - ДК
	//	2 - ВПУ
	//	3 - инженерный пульт (ИП УСДК)
	//	4 - УСДК/ДКА
	//	5 - инженерная панель (ИП ДКА)
	//	6 - система (ЭВМ)
	//	7 - система (ЭВМ)
	// 	8 - система (ЭВМ)
	// 	9 - система (ЭВМ)
	DDK   int  `json:"ddk"`   //Устройство ДК
	EDK   int  `json:"edk"`   //Неисправность ДК
	PDK   bool `json:"pdk"`   //Признак переходного периода ДК
	EEDK  int  `json:"eedk"`  //дополнительный код неисправности
	ODK   bool `json:"odk"`   //Открыта дверь ДК
	LDK   int  `json:"ldk"`   //Номер фазы на которой сгорели лампы
	FTUDK int  `json:"ftudk"` //Фаза ТУ ДК на момент передачи
	TDK   int  `json:"tdk"`   //Время отработки ТУ в секундах
	FTSDK int  `json:"ftsdk"` //Фаза ТС ДК
	TTCDK int  `json:"ttcdk"` //Время от начала фазы ТС в секундах

}

// Compare сравнивание истина если равны
func (d *DK) Compare(dd *DK) bool {
	return reflect.DeepEqual(d, dd)
}

// Model Описание модели устройства
type Model struct {
	VPCPDL int  `json:"vpcpdl"` //Версия ПО платы ПСПД до точки
	VPCPDR int  `json:"vpcpdr"` //Версия ПО платы ПСПД после точки
	VPBSL  int  `json:"vpbsl"`  //Версия ПО платы ПБС до точки
	VPBSR  int  `json:"vpbsr"`  //Версия ПО платы ПБС после точки
	C12    bool //Субблок С12
	STP    bool //Разрешение накопление статистики по ТП
	DKA    bool //Контроллер ДК-А
	DTA    bool //Детектор транспорта
}

// Compare сравнивание истина если равны
func (m *Model) Compare(mm *Model) bool {
	return reflect.DeepEqual(m, mm)
}

// ErrorDevice описание ошибок устройства
type ErrorDevice struct {
	V220DK1 bool //Срабатывание входа контроля 220В DK1
	V220DK2 bool //Срабатывание входа контроля 220В DK2
	RTC     bool // Неисправность часов RTC
	TVP1    bool //Неисправность ТВП1
	TVP2    bool //Неисправность ТВП2
	FRAM    bool //Неисправность FRAM
}

// Compare сравнивание истина если равны
func (e *ErrorDevice) Compare(ee *ErrorDevice) bool {
	return reflect.DeepEqual(e, ee)
}

// GPS описание состояния модуля GPS устройства
type GPS struct {
	Ok   bool //Исправно
	E01  bool // Нет связи с приемником
	E02  bool // Ошибка CRC
	E03  bool // Нет валидного времени
	E04  bool // Мало спутников
	Seek bool // Поиск спутников после включения
}

// Compare сравнивание истина если равны
func (g *GPS) Compare(gg *GPS) bool {
	return reflect.DeepEqual(g, gg)
}

// Input описание состояния входов устройства
type Input struct {
	V1 bool     //Неисправность входа 1
	V2 bool     //Неисправность входа 2
	V3 bool     //Неисправность входа 3
	V4 bool     //Неисправность входа 4
	V5 bool     //Неисправность входа 5
	V6 bool     //Неисправность входа 6
	V7 bool     //Неисправность входа 7
	V8 bool     //Неисправность входа 8
	S  [16]bool //Неисправность статистики
}

// Compare сравнивание истина если равны
func (i *Input) Compare(ii *Input) bool {
	return reflect.DeepEqual(i, ii)
}

// IsBroken возвращает true если есть хотя бы одна неисправность
func (i *Input) IsBroken() bool {
	iv := i.V1 || i.V2 || i.V3 || i.V4 || i.V5 || i.V6 || i.V7 || i.V8
	for _, ii := range i.S {
		iv = iv || ii
	}
	return iv
}

// ArchStat архивная статистика
type ArchStat struct {
	Region     int         `json:"region"` //Регион
	Area       int         `json:"area"`   //Район
	ID         int         `json:"id"`     //Номер перекрестка
	Date       time.Time   `json:"date"`
	Statistics []Statistic //Накопленная статистика
}

// Statistic статистика
type Statistic struct {
	Period int //Номер периода усреднения от начала суток
	Type   int //Тип статистики 1-интенсивность 2 скорость 3 расширенная
	TLen   int //Величина времени усреднения мин
	Hour   int //Час окончания периода
	Min    int //Минуты окончания периода
	Datas  []DataStat
}

// Compare сравнивание истина если равны
func (s *Statistic) Compare(ss *Statistic) bool {
	return reflect.DeepEqual(s, ss)
}

// Compare сравнивание истина если равны
func (d *DataStat) Compare(dd *DataStat) bool {
	return reflect.DeepEqual(d, dd)
}

// DataStat статистика по канально
type DataStat struct {
	Chanel   int `json:"ch"` //Номер канала
	Status   int `json:"st"` // Состояние 0-исправен 1-обрыв 2 - замыкание
	Intensiv int `json:"in"` //Интенсивность или скорость для Type 1 или 2 и только интенсивность для 3
	Speed    int `json:"sp"` //Скорость для типа 3
	Density  int `json:"d"`  //Плотность потока авт/км
	Occupant int `json:"o"`  //Занятость зоны в %
	GP       int `json:"g"`  //Средний зазор между авто с/авт
}

// StatusCommandDU команды ДУ
type StatusCommandDU struct {
	IsPK       bool //Назначен ПК
	IsCK       bool // назначена карта выбора по времени суток
	IsNK       bool //Назначена недельная карта
	IsDUDK1    bool //на 1 ДК есть команда ДУ
	IsDUDK2    bool //на 2 ДК есть команда ДУ
	IsReqSFDK1 bool //Есть запрос на передачу фаз по 1 ДК СФДК
	IsReqSFDK2 bool //Есть запрос на передачу фаз по 2 ДК СФДК
}

// Compare сравнивание истина если равны
func (s *StatusCommandDU) Compare(ss *StatusCommandDU) bool {
	return reflect.DeepEqual(s, ss)
}

// LogLine запись лога устройства
type LogLine struct {
	Time   time.Time
	Record string
	Info   string
}

// Compare сравнивание истина если равны
func (l *LogLine) Compare(ll *LogLine) bool {
	return reflect.DeepEqual(l, ll)
}

//Arrays описание и хранение всех настроечных массивов

// UserCross структура для передачи нового состояния перекрестка
type UserCross struct {
	User  string `json:"user"`
	State Cross  `json:"state"`
}

// Cross описание перекрестка
type Cross struct {
	Region       int     `json:"region"`  //Регион
	Area         int     `json:"area"`    //Район
	SubArea      int     `json:"subarea"` //подрайон
	ID           int     `json:"id"`      //Номер перекрестка
	IDevice      int     `json:"idevice"` // Назначенное на перекресток устройство
	Dgis         string  `json:"dgis"`    //Координаты перекрестка
	ConType      string  `json:"contype"` //Тип соединения устройства
	NumDev       int     `json:"numdev"`  //Номер устройства (УСДК,ДК-А,С12УСДК)
	Scale        float64 `json:"scale"`   //Масштаб
	Name         string  `json:"name"`
	Phone        string  `json:"phone"`  //Телефон
	WiFi         string  `json:"wifi"`   //IP для подключения ВПУ и прочего
	StatusDevice int     `json:"status"` // Статус устройства
	Arm          string
	WriteToDB    bool `json:"-"`  //Если истина то еще не записана в БД
	PK           int  `json:"pk"` //Номер плана координации
	CK           int  `json:"ck"` //Номер суточной карты
	NK           int  `json:"nk"` //Номер недельной карты
	Model        Model
	//Statistics   []Statistic    `json:"statis"` //Накопленная статистика
	Arrays binding.Arrays `json:"arrays"` //Файлы привязки
}

// ArrayPriv собственно массив привязки
type ArrayPriv struct {
	Number int
	NElem  int
	Array  []int
}

// Compare сравнивание истина если равны
func (a *ArrayPriv) Compare(aa *ArrayPriv) bool {
	return reflect.DeepEqual(a, aa)
}

// RecLogCtrl структура передачи инормации системного лога устройства
// Время всегда ставится системное
type LogRecord struct {
	ID        int // Уникальный номер контроллера
	Region    Region
	Type      int       // Тип уровня 0-технология 1 устройство 2 двери и лампы
	Time      time.Time //Время события
	LogString string    //Собственно сообщение
	Journal   Journal   //Запись журнала
}

// Status описание статуса устройства
type Status struct {
	StatusV220    int  `json:"s220"`     // Состояние питания 0 норма  25 - авария 220В 26 - выключен
	StatusGPS     int  `json:"sGPS"`     // Состояние GPS
	StatusServer  int  `json:"sServer"`  // Состояние связи с сервером
	StatusPSPD    int  `json:"sPSPD"`    // Состояние связи с платой ПСПД
	ErrorLastConn int  `json:"elc"`      // Код последней причины разрыва связи
	Ethernet      bool `json:"ethernet"` // true если связь через Ethernet
	TObmen        int  `json:"tobm"`     // Интервал обмена с сервером (минуты)
	LevelGSMNow   int  `json:"lnow"`     // уровень сигнала GSM  в текущей сессии
	LevelGSMLast  int  `json:"llast"`    // уровень сигнала GSM  в предыдущей сессии
	Motiv         int  `json:"motiv"`    // Мотив разрыва связи
}
type Traffic struct {
	FromDevice1Hour     uint64
	ToDevice1Hour       uint64
	LastToDevice1Hour   uint64
	LastFromDevice1Hour uint64
}

// Controller внутренне представление контроллера
type Controller struct {
	ID               int       `json:"id"`    // Уникальный номер контроллера
	Name             string    `json:"name"`  //Имя перекрестка если привязан
	StatusConnection bool      `json:"scon"`  // Статус соединения
	LastMyOperation  time.Time `json:"-"`     // Время последней операции обмена с устройством
	ConnectTime      time.Time `json:"ctime"` // Время подключения
	TimeDevice       time.Time `json:"dtime"` // Время устройства
	LastOperation    time.Time `json:"ltime"` // Время последней успешной операции обмена с устройством
	WriteToDB        bool      `json:"-"`     //Если истина то еще не записана в БД
	// 1 - выбор ПК по времени по суточной карте ВР-СК;
	// 2 - выбор ПК по недельной карте ВР-НК;
	// 3 - выбор ПК по времени по суточной карте, назначенной
	// оператором ДУ-СК;
	// 4 - выбор ПК по недельной карте, назначенной оператором
	// ДУ-НК;
	// 5 - план по запросу оператора ДУ-ПК;
	// 6 - резервный план (отсутствие точного времени) РП;
	// 7 – коррекция привязки с ИП;
	// 8 – коррекция привязки с сервера;
	// 9 – выбор ПК по годовой карте;
	// 10 – выбор ПК по ХТ;
	// 11 – выбор ПК по картограмме;
	// 12 – противозаторовое управление.
	TechMode int `json:"techmode"` //Технологический режим

	Base            bool   `json:"base"` //Если истина то работает по базовой привязке
	PK              int    `json:"pk"`   //Номер плана координации
	CK              int    `json:"ck"`   //Номер суточной карты
	NK              int    `json:"nk"`   //Номер недельной карты
	IPHost          string `json:"ip"`   //IP адресс контроллера
	StatusCommandDU StatusCommandDU
	DK              DK
	TMax            int64 `json:"tmax"` //Максимальное время ожидания ответа от сервера в секундах
	TimeOut         int64 `json:"tout"` //TimeOut на чтение от контроллера в секундах
	Model           Model
	Error           ErrorDevice
	GPS             GPS
	Input           Input
	Status          Status
	// Statistics      []Statistic
	Arrays   []ArrayPriv `json:"arrays"` //Файлы привязки
	LogLines []LogLine
	Traffic  Traffic
}

// Compare сравнивание истина если равны
func (cc *Controller) Compare(ccc *Controller) bool {
	return reflect.DeepEqual(cc, ccc)
}

// JSONLog структура для хранения адреса
type Journal struct {
	Device string `json:"device"`
	Arm    string `json:"arm"`
	Status string `json:"status"`
	Rezim  string `json:"rez"`
	Phase  string `json:"phase"`
	NK     string `json:"nk"`
	CK     string `json:"ck"`
	PK     string `json:"pk"`
	Note   string `json:"note"`
}

// SetDefault Заполнить по умолчанию
func SetDefault(c *Controller, key Region) {
	mutexCtrl.Lock()
	defer mutexCtrl.Unlock()
	cr, is := crosses[key]
	if !is {
		logger.Error.Fatalf("нет такого %s", key.ToKey())
	}
	c.Name = cr.Name
	c.ID = cr.IDevice
	c.NK = 1
	c.PK = 1
	c.CK = 1
	c.LastMyOperation = time.Now()
	c.ConnectTime = time.Unix(0, 0)
	c.TechMode = 1
	c.DK.TDK = 1
	c.Base = true
	var m Model
	m.VPCPDL = 0
	m.VPCPDR = 0
	m.VPBSL = 0
	m.VPBSR = 0
	c.Traffic = Traffic{}
	c.Model = m

	c.Arrays = MakeArrays(*binding.NewArrays())
	c.LogLines = make([]LogLine, 0)
}

// NewCross создание нового описания перекрестка
func NewCross() *Cross {
	r := new(Cross)
	//r.Statistics = make([]Statistic, 0)
	r.Arrays = *binding.NewArrays()
	return r
}
func MakeArrays(ar binding.Arrays) []ArrayPriv {
	r := make([]ArrayPriv, 0)
	if !ar.StatDefine.IsEmpty() {
		buffer := ar.StatDefine.ToBuffer() //
		r = appBuffer(r, buffer)
	}
	if !ar.PointSet.IsEmpty() {
		buffer := ar.PointSet.ToBuffer() //
		r = appBuffer(r, buffer)
	}
	if !ar.UseInput.IsEmpty() {
		buffer := ar.UseInput.ToBuffer() //
		r = appBuffer(r, buffer)

	}
	if !ar.TimeDivice.IsEmpty() {
		buffer := ar.TimeDivice.ToBuffer() //
		r = appBuffer(r, buffer)
	}
	if !ar.SetupDK.IsEmpty() {
		buffer := ar.SetupDK.ToBuffer() //
		r = appBuffer(r, buffer)
	}
	if !ar.SetCtrl.IsEmpty() {
		buffer := ar.SetCtrl.ToBuffer() //
		r = appBuffer(r, buffer)
	}
	if !ar.SetTimeUse.IsEmpty() {
		buffer := ar.SetTimeUse.ToBuffer(157) //
		r = appBuffer(r, buffer)
		buffer = ar.SetTimeUse.ToBuffer(148) //
		r = appBuffer(r, buffer)
	}
	for i := 1; i < 13; i++ {
		r = appBuffer(r, ar.SetDK.DK[i-1].ToBuffer())
	}
	for _, ns := range ar.WeekSets.WeekSets { //
		if !ns.IsEmpty() {
			buffer := ns.ToBuffer()
			r = appBuffer(r, buffer)
		}
	}
	for _, ss := range ar.DaySets.DaySets { //
		if !ss.IsEmpty() {
			buffer := ss.ToBuffer()
			r = appBuffer(r, buffer)
		}
	}
	for _, ys := range ar.MonthSets.MonthSets { //
		if !ys.IsEmpty() {
			buffer := ys.ToBuffer()
			r = appBuffer(r, buffer)
		}
	}

	return r
}
func appBuffer(res []ArrayPriv, buffer []int) []ArrayPriv {
	return append(res, makePriv(buffer))
}
func makePriv(buffer []int) ArrayPriv {
	r := new(ArrayPriv)
	r.Array = make([]int, 0)
	r.Number = buffer[2]
	r.NElem = buffer[4]
	for i := 3; i < len(buffer); i++ {
		r.Array = append(r.Array, buffer[i])
	}
	return *r
}
