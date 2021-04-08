// Package pudge  выполняет
// 	1. Ведется текущее состояние контроллеров
// 	2. Сидит прием по каналу запросов на чтение со стороны сервера АРМ
// 	3. Открывается канал приема запросов на запись от сервера коммуникации
// 	4. Если сервер коммуникации присылает запрос на запись нового состояния то
// 		делается проверка на существенное измение и если это так то новое состояние посылается в бд логгирования
// 	5. Открывается прием по каналу запросов от сервера АРМ после отправки копиии запроса в канал сервера канала данный
// 		запрос посылается серверу коомуникации
// 	6. По времени заданному в настройках делается полная копия состояния всех контроллеров в базу данных простой посылкой
// 		копии
package pudge

import (
	"math/rand"
	"reflect"
	"strconv"
	"time"

	"github.com/ruraomsk/TLServer/logger"
	"github.com/ruraomsk/ag-server/binding"
)

//Region указатель на номер перекрестка
type Region struct {
	Region int //Код региона
	Area   int //Код района
	ID     int //Номер перекрестка
}

//ToKey создает строковый ключ
func (r *Region) ToKey() string {
	return strconv.Itoa(r.Region) + ";" + strconv.Itoa(r.Area) + ";" + strconv.Itoa(r.ID)
}

//StatusConnection статус соединения
type StatusConnection int

//DK диагностика состояния по ДК
type DK struct {
	RDK int `json:"rdk"` //Режим ДК
	// 1 2 Ручное управление
	// 3 Зеленая улица
	// 4 Диспетчерское управление
	// 5 6 Локальное управление
	// 8 9 Координированное управление
	FDK int `json:"fdk"` //Фаза ДК
	// от 1 до 8 номера рабочих фаз
	// 9 промежуточный такт
	// 10 желтое мигание
	// 11 отключен светофор
	// 12 кругом краснный
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

//Compare сравнивание истина если равны
func (d *DK) Compare(dd *DK) bool {
	return reflect.DeepEqual(d, dd)
}

//Model Описание модели устройства
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

func randBool() bool {
	if rand.Intn(2) == 1 {
		return true
	}
	return false
}

//MakeError случайным образом создает ошибку
func (e *ErrorDevice) MakeError() bool {
	switch rand.Intn(7) {
	case 0:
		return false
	case 1:
		e.V220DK1 = randBool()
	case 2:
		e.V220DK2 = randBool()
	case 3:
		e.RTC = randBool()
	case 4:
		e.TVP1 = randBool()
	case 5:
		e.TVP2 = randBool()
	case 6:
		e.FRAM = randBool()
	}
	return true
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

//MakeError порождает ошибки или испавность
func (g *GPS) MakeError() bool {
	switch rand.Intn(7) {
	case 0:
		return false
	case 1:
		g.Ok = randBool()
	case 2:
		g.E01 = randBool()
	case 3:
		g.E02 = randBool()
	case 4:
		g.E03 = randBool()
	case 5:
		g.E04 = randBool()
	case 6:
		g.Seek = randBool()
	}
	return true
}

//Compare сравнивание истина если равны
func (g *GPS) Compare(gg *GPS) bool {
	return reflect.DeepEqual(g, gg)
}

//Input описание состояния входов устройства
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

//MakeError порождает ошибки или испавность
func (i *Input) MakeError() bool {
	switch rand.Intn(9) {
	case 0:
		return false
	case 1:
		i.V1 = randBool()
	case 2:
		i.V2 = randBool()
	case 3:
		i.V3 = randBool()
	case 4:
		i.V4 = randBool()
	case 5:
		i.V5 = randBool()
	case 6:
		i.V6 = randBool()
	case 7:
		i.V7 = randBool()
	case 8:
		i.V8 = randBool()
	}
	return true
}

//Compare сравнивание истина если равны
func (i *Input) Compare(ii *Input) bool {
	return reflect.DeepEqual(i, ii)
}

//ArchStat архивная статистика
type ArchStat struct {
	Region     int         `json:"region"` //Регион
	Area       int         `json:"area"`   //Район
	ID         int         `json:"id"`     //Номер перекрестка
	Date       time.Time   `json:"date"`
	Statistics []Statistic //Накопленная статистика
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
	IsCK       bool // назначена карта выбора по времени суток
	IsNK       bool //Назначена недельная карта
	IsDUDK1    bool //на 1 ДК есть команда ДУ
	IsDUDK2    bool //на 2 ДК есть команда ДУ
	IsReqSFDK1 bool //Есть запрос на передачу фаз по 1 ДК СФДК
	IsReqSFDK2 bool //Есть запрос на передачу фаз по 2 ДК СФДК
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

//Arrays описание и хранение всех настроечных массивов

//UserCross структура для передачи нового состояния перекрестка
type UserCross struct {
	User  string `json:"user"`
	State Cross  `json:"state"`
}

//Cross описание перекрестка
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
	WiFi         string  `json:"wifi""`  //IP для подключения ВПУ и прочего
	StatusDevice int     `json:"status"` // Статус устройства
	WriteToDB    bool    `json:"-"`      //Если истина то еще не записана в БД
	PK           int     `json:"pk"`     //Номер плана координации
	CK           int     `json:"ck"`     //Номер суточной карты
	NK           int     `json:"nk"`     //Номер недельной карты
	Model        Model
	//Statistics   []Statistic    `json:"statis"` //Накопленная статистика
	Arrays binding.Arrays `json:"arrays"` //Файлы привязки
}

//ArrayPriv собственно массив привязки
type ArrayPriv struct {
	Number int
	NElem  int
	Array  []int
}

//Compare сравнивание истина если равны
func (a *ArrayPriv) Compare(aa *ArrayPriv) bool {
	return reflect.DeepEqual(a, aa)
}

//RecLogCtrl структура передачи инормации системного лога устройства
// Время всегда ставится системное
type RecLogCtrl struct {
	ID        int       // Уникальный номер контроллера
	Type      int       // Тип уровня -1 тенология и устройство 0-технология 1 устройство 2 двери и лампы
	Time      time.Time //Время события
	LogString string    //Собственно сообщение
}

//Status описание статуса устройства
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
	FromDevice15Min     uint64
	FromDevice1Hour     uint64
	ToDevice15Min       uint64
	ToDevice1Hour       uint64
	LastToDevice15Min   uint64
	LastToDevice1Hour   uint64
	LastFromDevice15Min uint64
	LastFromDevice1Hour uint64
}

//Controller внутренне представление контроллера
type Controller struct {
	ID               int       `json:"id"`       // Уникальный номер контроллера
	Name             string    `json:"name"`     //Имя перекрестка если привязан
	StatusConnection bool      `json:"scon"`     // Статус соединения
	LastOperation    time.Time `json:"ltime"`    // Время последней операции обмена с устройством
	TimeDevice       time.Time `json:"dtime"`    // Время устройства
	WriteToDB        bool      `json:"-"`        //Если истина то еще не записана в БД
	TechMode         int       `json:"techmode"` //Технологический режим

	// Технологический режим

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

	//Local           bool   `json:"local"` //Если истина то контроллер находится в режиме загрузки файлов привязки
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
	Statistics      []Statistic
	Arrays          []ArrayPriv `json:"arrays"` //Файлы привязки
	LogLines        []LogLine
	Traffic         Traffic
}

//Compare сравнивание истина если равны
func (cc *Controller) Compare(ccc *Controller) bool {
	return reflect.DeepEqual(cc, ccc)
}

//SetDefault Заполнить по умолчанию
func SetDefault(c *Controller, key Region) {
	cr, is := crosses[key.ToKey()]
	if !is {
		logger.Error.Fatalf("нет такого %s", key)
	}
	c.Name = cr.Name
	c.ID = cr.IDevice
	c.NK = 1
	c.PK = 1
	c.CK = 1
	c.LastOperation = time.Now()
	c.TechMode = 1
	c.DK.TDK = 1
	c.Base = true
	var m Model
	m.VPCPDL = 0
	m.VPCPDR = 0
	m.VPBSL = 0
	m.VPBSR = 0

	c.Model = m
	c.Statistics = make([]Statistic, 0)
	c.Arrays = make([]ArrayPriv, 0)
	c.LogLines = make([]LogLine, 0)
}

//NewCross создание нового описания перекрестка
func NewCross() *Cross {
	r := new(Cross)
	//r.Statistics = make([]Statistic, 0)
	r.Arrays = *binding.NewArrays()
	return r
}
