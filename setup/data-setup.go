package setup

//Set переменная для хранения текущих настроек
var Set *Setup

//Setup общая структура для настройки всей системы
type Setup struct {
	Home       string     `toml:"home"`
	Location   string     `toml:"location"`  //Локация временной зоны
	StepPudge  int        `toml:"steppudge"` //Шаг сохранения в секундах
	DataBase   DataBase   `toml:"dataBase"`
	CommServer CommServer `toml:"commServer"`
	Controller Controller `toml:"controller"`
	XCtrl      XCtrl      `toml:"xctrl"`
	Saver      Saver      `toml:"saver"`
}
type Saver struct {
	Keys [][]string `toml:"keys"`
}

//CommServer настройки для сервера коммуникации
type CommServer struct {
	Port         int   `toml:"port"`          //Стартовый номер порта на прием
	PortCommand  int   `toml:"portc"`         //Порт приема команд от сервера АРМ
	PortArray    int   `toml:"porta"`         //Порт приема массивов привязки от сервера АРМ
	PortProtocol int   `toml:"portp"`         //Порт приема изменения протокола от сервера АРМ
	TimeOutRead  int64 `toml:"read_timeout"`  //Таймаут на чтение если данные должны быть получены
	TimeOutWrite int64 `toml:"write_timeout"` //Таймаут на запись если данные должны быть переданы
	ID           int   `toml:"id"`
}

//DataBase настройки базы данных postresql
type DataBase struct {
	Host     string `toml:"host"`
	Port     int    `toml:"port"`
	User     string `toml:"user"`
	Password string `toml:"password"`
	DBname   string `toml:"dbname"`
}

//Controller настройки имитатора
type Controller struct {
	IP      string `toml:"ip"`
	GuiPort int    `toml:"guiport"`
	Port    int    `toml:"port"`
	Step    int    `toml:"step"` //Интервал времени в секундах для расчетов
	Random  bool   `toml:"random"`
}

//XCtrl настройки подсистемы характерных точек
type XCtrl struct {
	Switch    bool `toml:"switch"`
	Calculate bool `toml:"calculate"`
	StepCalc  int  `toml:"stepCalc"`
	StepSend  int  `toml:"stepSend"`
}
