package setup

// Set переменная для хранения текущих настроек
var Set *Setup

// Setup общая структура для настройки всей системы
type Setup struct {
	Home       string     `toml:"home"`
	Debug      string     `toml:"debug"`
	Location   string     `toml:"location"`  //Локация временной зоны
	StepPudge  int        `toml:"steppudge"` //Шаг сохранения в секундах
	Secret     bool       `toml:"secret"`
	Version    int        `toml:"version"`
	LogSystem  LogSystem  `toml:"logsystem"`
	DataBase   DataBase   `toml:"dataBase"`
	CommServer CommServer `toml:"commServer"`
	XCtrl      XCtrl      `toml:"xctrl"`
	Saver      Saver      `toml:"saver"`
	Dumper     Dumper     `toml:"dumper"`
	Statistic  Statistic  `toml:"statistic"`
	Loader     Loader     `toml:"loader"`
	Camera     Camera     `toml:"camera"`
}
type Camera struct {
	Make bool   `toml:"make"`
	Path string `toml:"path"`
}
type LogSystem struct {
	Make      bool   `toml:"make"`
	StartTime string `toml:"start"`
	Period    int    `toml:"period"`
}

// Loader описание загрузчика
type Loader struct {
	Make    bool       `toml:"make"`     //Делать ли прием со стороны
	Port    int        `toml:"port"`     //Стартовый номер порта на прием
	SVGPort int        `toml:"svgPort"`  //номер порта на прием свг
	Path    string     `toml:"pathSVG"`  //Путь для каталога рисунков перекрестков
	Files   [][]string `toml:"filesSVG"` //Файлы которые описывают перекресток
}

// Statistic описание статистики
type Statistic struct {
	Make    bool       `toml:"make"`    //true если выполнять
	Regions [][]string `toml:"regions"` //Перечень регионов с временами старта во времени сервера и признаком какую дату использовать
}
type Saver struct {
	Make    bool       `toml:"make"`     //true если выполнять
	Remote  string     `toml:"remote"`   //TCP до сервера приема
	Svg     string     `toml:"svg"`      //TCP до сервера приема SVG
	File    string     `toml:"file"`     //Имя и путь для файла сохраниея команд SQL
	PreSQL  []string   `toml:"presql"`   //Команды выполняемые при передаче первого дампа
	Step    int        `toml:"step"`     //Интервал времени в секундах для расчетов
	Keys    [][]string `toml:"keys"`     //Имена твблиц с ключами для пересылки
	StepSVG int        `toml:"stepSVG"`  //Интервал времени в секундах для расчетов SVG
	Path    string     `toml:"pathSVG"`  //Путь для каталога рисунков перекрестков
	Files   [][]string `toml:"filesSVG"` //Файлы которые описывают перекресток
}

// CommServer настройки для сервера коммуникации
type CommServer struct {
	Port         int   `toml:"port"`          //Стартовый номер порта на прием
	PortCommand  int   `toml:"portc"`         //Порт приема команд от сервера АРМ
	PortArray    int   `toml:"porta"`         //Порт приема массивов привязки от сервера АРМ
	PortProtocol int   `toml:"portp"`         //Порт приема изменения протокола от сервера АРМ
	PortDevices  int   `toml:"portd"`         //Порт передачи номера фазы и времени фазы серверу АРМ
	TimeOutRead  int64 `toml:"read_timeout"`  //Таймаут на чтение если данные должны быть получены
	TimeOutWrite int64 `toml:"write_timeout"` //Таймаут на запись если данные должны быть переданы
	ID           int   `toml:"id"`
	MaxCon       int   `toml:"maxcon"`
}

// DataBase настройки базы данных postresql
type DataBase struct {
	Host     string `toml:"host"`
	Port     int    `toml:"port"`
	User     string `toml:"user"`
	Password string `toml:"password"`
	DBname   string `toml:"dbname"`
}

// XCtrl настройки подсистемы характерных точек
type XCtrl struct {
	Port        int     `toml:"port"` //Порт ожидания команд от системы
	Switch      bool    `toml:"switch"`
	StepDev     int     `toml:"stepdev"`  //Шаг опроса устройств в минутах
	StepCalc    int     `toml:"stepcalc"` //Шаг расчета
	ShiftDevice string  `toml:"shdev"`    //Смещение от шага секунды
	ShiftCtrl   string  `toml:"shctrl"`   //Смещение для запуска управления секунды
	NameUser    string  `toml:"nameuser"`
	FullHost    string  `toml:"fullhost"`
	Regions     [][]int `toml:"regions"` //Перечень регионов с временами старта во времени сервера и признаком какую дату использовать
}
type Dumper struct {
	Make    bool   `toml:"make"`
	Path    string `toml:"path"`
	Time    string `toml:"time"`
	PathSVG string `toml:"pathSVG"` //Путь для каталога рисунков перекрестков
}
