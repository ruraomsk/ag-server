package pudge

/*
	В этом разделе ведутся следующие работы
		1. Ведется текущее состояние контроллеров
		2. Сидит прием по каналу запросов на чтение со стороны сервера АРМ
		3. Открывается канал приема запросов на запись от сервера коммуникации
		4. Если сервер коммуникации присылает запрос на запись нового состояния то
			делается проверка на существенное измение и если это так то новое состояние посылается в бд логгирования
		5. Открывается прием по каналу запросов от сервера АРМ после отправки копиии запроса в канал сервера канала данный
			запрос посылается серверу коомуникации
		6. По времени заданному в настройках делается полная копия состояния всех контроллеров в базу данных простой посылкой
			копии
*/

import (
	"reflect"
	"strconv"
	"time"
)

//Region указатель на номер перекрестка
type Region struct {
	Region int //Код региона
	ID     int //Номер перекрестка
}

func (r *Region) toKey() string {
	return strconv.Itoa(r.Region) + ";" + strconv.Itoa(r.ID)
}

//Controllers возврат выбранных контроллеров
type Controllers struct {
	Contrs []Controller
}

//StatusConnection статус соединения
type StatusConnection int

const (
	//Connected Ok
	Connected StatusConnection = iota
	//NotConnected not Ok
	NotConnected
	//Undefine Undefine
	Undefine
)

//DK диагностика состояния по ДК
type DK struct {
	RDK   int  `json:"rdk"`   //Режим ДК
	FDK   int  `json:"fdk"`   //Фаза ДК
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

//ToList для вывода в гуи
func (d *DK) ToList(result []string) []string {
	r := "Режим ДК;" + strconv.Itoa(d.RDK)
	result = append(result, r)
	r = "Фаза ДК;" + strconv.Itoa(d.FDK)
	result = append(result, r)
	r = "Устройство ДК;" + strconv.Itoa(d.DDK)
	result = append(result, r)
	r = "Код неисправности ДК;" + strconv.Itoa(d.EDK)
	result = append(result, r)
	r = "Признак переходного периода ДК;" + strconv.FormatBool(d.PDK)
	result = append(result, r)
	r = "Дополнительный код неисправности ДК;" + strconv.Itoa(d.EEDK)
	result = append(result, r)
	r = "Открыта дверь ДК;" + strconv.FormatBool(d.ODK)
	result = append(result, r)
	r = "Номер фазы на которой сгорели лампы ДК;" + strconv.Itoa(d.LDK)
	result = append(result, r)
	r = "Фаза ТУ ДК на момент передачи;" + strconv.Itoa(d.FTUDK)
	result = append(result, r)
	r = "Время отработки ТУ в секундах;" + strconv.Itoa(d.TDK)
	result = append(result, r)
	r = "Фаза ТС ДК;" + strconv.Itoa(d.FTSDK)
	result = append(result, r)
	r = "Время от начала фазы в секундах" + strconv.Itoa(d.TTCDK)
	result = append(result, r)
	return result
}

//Compare сравнивание истина если равны
func (d *DK) Compare(dd *DK) bool {
	return reflect.DeepEqual(d, dd)
}

//Model Описание модели устройства
type Model struct {
	VPCPD int  //Версия ПО платы ПСПД
	VPBS  int  //Версия ПО платы ПБС
	C12   bool //Субблок С12
	STP   bool //Разрешение накопление статистики по ТП
	DKA   bool //Контроллер ДК-А
	DTA   bool //Детектор транспорта
}

//ToList для вывода в гуи
func (m *Model) ToList(result []string) []string {
	r := "Версия ПО пдаты ПСПД;" + strconv.Itoa(m.VPCPD)
	result = append(result, r)
	r = "Версия ПО пдаты ПБС;" + strconv.Itoa(m.VPBS)
	result = append(result, r)
	r = "Субблок С12;" + strconv.FormatBool(m.C12)
	result = append(result, r)
	r = "Разрешение накопления статистики по ТП;" + strconv.FormatBool(m.STP)
	result = append(result, r)
	r = "Контроллер ДК-А;" + strconv.FormatBool(m.DKA)
	result = append(result, r)
	r = "Детектор транспорта;" + strconv.FormatBool(m.DTA)
	result = append(result, r)
	return result
}

//Compare сравнивание истина если равны
func (m *Model) Compare(mm *Model) bool {
	return reflect.DeepEqual(m, mm)
}

//ErrorDevice описание ошибок устройства
type ErrorDevice struct {
	V220DK1 bool //Срабатывание входа контроля 220В DK1
	V220DK2 bool //Срабатывание входа контроля 220В DK2
	RTC     bool // Неисправность часов RTC
	TVP1    bool //Неисправность ТВП1
	TVP2    bool //Неисправность ТВП2
	FRAM    bool //Неисправность FRAM
}

//ToList для вывода в гуи
func (e *ErrorDevice) ToList(result []string) []string {
	r := "Срабатывание входа контроля 220В DK1;" + strconv.FormatBool(e.V220DK1)
	result = append(result, r)
	r = "Срабатывание входа контроля 220В DK2;" + strconv.FormatBool(e.V220DK2)
	result = append(result, r)
	r = "Неисправность часов RTC;" + strconv.FormatBool(e.RTC)
	result = append(result, r)
	r = "Неисправность ТВП1;" + strconv.FormatBool(e.TVP1)
	result = append(result, r)
	r = "Неисправность ТВП2;" + strconv.FormatBool(e.TVP2)
	result = append(result, r)
	r = "Неисправность FRAM;" + strconv.FormatBool(e.FRAM)
	result = append(result, r)
	return result
}

//Compare сравнивание истина если равны
func (e *ErrorDevice) Compare(ee *ErrorDevice) bool {
	return reflect.DeepEqual(e, ee)
}

//GPS описание состояния модуля GPS устройства
type GPS struct {
	Ok   bool //Исправно
	E01  bool // Нет связи с приемником
	E02  bool // Ошибка CRC
	E03  bool // Нет валидного времени
	E04  bool // Мало спутников
	Seek bool // Поиск спутников после включения
}

//ToList для вывода в гуи
func (g *GPS) ToList(result []string) []string {
	r := "Исправно;" + strconv.FormatBool(g.Ok)
	result = append(result, r)
	r = "Нет связи с приемником;" + strconv.FormatBool(g.E01)
	result = append(result, r)
	r = "Ошибка CRC;" + strconv.FormatBool(g.E02)
	result = append(result, r)
	r = "Нет валидного времени;" + strconv.FormatBool(g.E03)
	result = append(result, r)
	r = "Мало спутников;" + strconv.FormatBool(g.E04)
	result = append(result, r)
	r = "Поиск спутников после включения;" + strconv.FormatBool(g.Seek)
	return result
}

//Compare сравнивание истина если равны
func (g *GPS) Compare(gg *GPS) bool {
	return reflect.DeepEqual(g, gg)
}

//Input описание состояния входов устройства
type Input struct {
	V1 bool //Неисправность входа 1
	V2 bool //Неисправность входа 2
	V3 bool //Неисправность входа 3
	V4 bool //Неисправность входа 4
	V5 bool //Неисправность входа 5
	V6 bool //Неисправность входа 6
	V7 bool //Неисправность входа 7
	V8 bool //Неисправность входа 8
}

//Compare сравнивание истина если равны
func (i *Input) Compare(ii *Input) bool {
	return reflect.DeepEqual(i, ii)
}

//Statistic статистика
type Statistic struct {
	Period int //Номер периода усреднения от начала суток
	Type   int //Тип статистики 1-интенсивность скорость
	TLen   int //Величина времения усреднения мин
	Hour   int //Час окончания периода
	Min    int //Минуты окончания периода
	Datas  []DataStat
}

//Compare сравнивание истина если равны
func (s *Statistic) Compare(ss *Statistic) bool {
	return reflect.DeepEqual(s, ss)
}

//DataStat статистика по канально
type DataStat struct {
	Chanel   int //Номер канала
	Status   int // Состояние 0-исправен 1-обрыв 2 - замыкание
	Intensiv int //Интенсивность
}

//StatusCommandDU команды ДУ
type StatusCommandDU struct {
	IsPK       bool //Назначен ПК
	IsPKS      bool // назначена карта выбора по времени суток
	IsNK       bool //Назначена недельная карта
	IsDUDK1    bool //на 1 ДК есть команда ДУ
	IsDUDK2    bool //на 2 ДК есть команда ДУ
	IsReqSFDK1 bool //Есть запрос на передачу фаз по 1 ДК СФДК
	IsReqSFDK2 bool //Есть запрос на передачу фаз по 2 ДК СФДК
}

//ToList для вывода в гуи
func (s *StatusCommandDU) ToList(result []string) []string {
	r := "Назначен ПК;" + strconv.FormatBool(s.IsPK)
	result = append(result, r)
	r = "Назначена карта выбора по времени суток;" + strconv.FormatBool(s.IsPKS)
	result = append(result, r)
	r = "Назначена недельная карта;" + strconv.FormatBool(s.IsNK)
	result = append(result, r)
	r = "На 1 ДК есть команда ДУ;" + strconv.FormatBool(s.IsDUDK1)
	result = append(result, r)
	r = "На 2 ДК есть команда ДУ;" + strconv.FormatBool(s.IsDUDK2)
	result = append(result, r)
	r = "Есть запрос на передачу фаз по 1 ДК СФДК;" + strconv.FormatBool(s.IsReqSFDK1)
	result = append(result, r)
	r = "Есть запрос на передачу фаз по 2 ДК СФДК;" + strconv.FormatBool(s.IsReqSFDK2)
	result = append(result, r)
	return result
}

//Compare сравнивание истина если равны
func (s *StatusCommandDU) Compare(ss *StatusCommandDU) bool {
	return reflect.DeepEqual(s, ss)
}

//LogLine запись лога устройства
type LogLine struct {
	Time   time.Time
	Record int
	Info   int
}

//Compare сравнивание истина если равны
func (l *LogLine) Compare(ll *LogLine) bool {
	return reflect.DeepEqual(l, ll)
}

//ArrayPriv Массим привязки
type ArrayPriv struct {
	Number int
	Array  []int
}

//Compare сравнивание истина если равны
func (a *ArrayPriv) Compare(aa *ArrayPriv) bool {
	return reflect.DeepEqual(a, aa)
}

//Cross описание перекрестка
type Cross struct {
	Region       int         `json:"region"`
	ID           int         `json:"id"`
	IDevice      int         `json:"idevice"`
	Name         string      `json:"name"`
	StatusDevice int         `json:"status"` // Статус устройства
	WriteToDB    bool        `json:"-"`      //Если истина то еще не записана в БД
	PK           int         `json:"pk"`     //Номер плана координации
	CK           int         `json:"ck"`     //Номер суточной карты
	NK           int         `json:"nk"`     //Номер недельной карты
	Statistics   []Statistic //Накопленная статистика
	Arrays       []ArrayPriv //Файлы привязки

}

//Controller внутренне представление контроллера
type Controller struct {
	ID               int              `json:"id"`    // Уникальный номер контроллера
	Name             string           `json:"name"`  //Имя перекрестка если привязан
	StatusConnection StatusConnection `json:"scon"`  // Статус соединения
	LastOperation    time.Time        `json:"ltime"` // Время последней операции обмена с устройством
	WriteToDB        bool             `json:"-"`     //Если истина то еще не записана в БД
	TexRezim         int              `json:"rezim"` //Технологический режим
	Base             bool             `json:"base"`  //Если истина то работает по базовой привязке
	PK               int              `json:"pk"`    //Номер плана координации
	CK               int              `json:"ck"`    //Номер суточной карты
	NK               int              `json:"nk"`    //Номер недельной карты
	StatusCommandDU  StatusCommandDU
	DK1              DK
	DK2              DK
	TMax             int `json:"tmax"` //Максимальное время ожидания ответа от сервера в секундах
	Model            Model
	Error            ErrorDevice
	GPS              GPS
	Input            Input
	Statistics       []Statistic
	Arrays           []ArrayPriv
	LogLines         []LogLine
}

//ToList для вывода в гуи
func (c *Controller) ToList() []string {
	result := make([]string, 0)
	r := "Технологический режим;" + strconv.Itoa(c.TexRezim)
	result = append(result, r)
	r = "Работаем в базовой привязке;" + strconv.FormatBool(c.Base)
	result = append(result, r)
	r = "Номер плана координации;" + strconv.Itoa(c.PK)
	result = append(result, r)
	r = "Номер суточной карты;" + strconv.Itoa(c.CK)
	result = append(result, r)
	r = "Номер недельной карты;" + strconv.Itoa(c.CK)
	result = append(result, r)
	r = "Максимальное время ожидания от сервера;" + strconv.Itoa(c.TMax)
	result = append(result, r)
	r = "ДК1; "
	result = append(result, r)
	result = c.DK1.ToList(result)
	r = "ДК2; "
	result = append(result, r)
	result = c.DK2.ToList(result)
	r = "МОДЕЛЬ УСТРОЙСТВА; "
	result = append(result, r)
	result = c.Model.ToList(result)
	r = "ОШИБКИ УСТРОЙСТВА; "
	result = append(result, r)
	result = c.Error.ToList(result)
	r = "СОСТОЯНИЕ GPS ПРИЕМНИКА; "
	result = append(result, r)
	result = c.GPS.ToList(result)
	r = "ТЕКУЩИЕ КОМАНДЫ ДУ; "
	result = append(result, r)
	result = c.StatusCommandDU.ToList(result)
	return result
}

//Compare сравнивание истина если равны
func (c *Controller) Compare(cc *Controller) bool {
	return reflect.DeepEqual(c, cc)
}

//SetDefault Заполнить по умолчанию
func SetDefault(c *Controller) {
	c.LastOperation = time.Unix(0, 0)
	c.TexRezim = 1
	c.Base = true
	c.PK = 1
	c.CK = 2
	c.NK = 3
	var cc StatusCommandDU
	cc.IsPK = true
	cc.IsPKS = true
	cc.IsNK = true
	c.StatusCommandDU = cc
	var dk DK
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
	var m Model
	m.VPCPD = 101
	m.VPBS = 2
	m.C12 = true
	m.STP = true
	m.DKA = true
	m.DTA = true
	c.Model = m
	var er ErrorDevice
	er.V220DK1 = false
	er.V220DK2 = false
	er.RTC = false
	er.TVP1 = false
	er.TVP2 = false
	er.FRAM = false
	c.Error = er
	var gps GPS
	gps.Ok = true
	c.GPS = gps
	var input Input
	input.V1 = false
	c.Input = input
	c.Statistics = make([]Statistic, 0)
	c.Arrays = make([]ArrayPriv, 0)
	c.LogLines = make([]LogLine, 0)
}
