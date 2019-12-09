package setup

import "time"

//Setup общая структура для настройки всей системы
type Setup struct {
	DataBase   DataBase
	Server     Server
	Pudge      Pudge
	CommServer CommServer
	Controller Controller
	Location   string `json:"location"` //Локация временной зоны
}

//CommServer настройки для сервера коммуникации
type CommServer struct {
	Port        int `json:"port"`  //Стартовый номер порта на прием
	PortCommand int `json:"portc"` //Порт приема команд от сервера АРМ
	PortArray   int `json:"porta"` //Порт приема массивов привязки от сервера АРМ

	TimeOutRead  time.Duration `json:"read_timeout"`    //Таймаут на чтение если данные должны быть получены
	TimeOutWrite time.Duration `json:"write_timeout"`   //Таймаут на запись если данные должны быть переданы
	KeepAlive    time.Duration `json:"time_keep_alive"` //Интервал времени в течении которого должен прийти keepalive от устройства
	IDServer     uint8         `json:"id"`              //ID сервера или 0xa7 или 0x8d
}

//Server настройки для сервера армов
type Server struct {
	Port         int           `json:"port"`            //Стартовый номер порта на прием
	TimeOutRead  time.Duration `json:"read_timeout"`    //Таймаут на чтение если данные должны быть получены
	TimeOutWrite time.Duration `json:"write_timeout"`   //Таймаут на запись если данные должны быть переданы
	KeepAlive    time.Duration `json:"time_keep_alive"` //Интервал времени в течении которого должен прийти keepalive от устройства

}

//Pudge настройки подсистемы хранения состояние контроллеров
type Pudge struct {
	StepSave  int    `json:"step"` //Интервал времени в секундах для сохранения состояния контроллеров
	TableLog  string `json:"log"`  //Имя таблицы куда пишем лог
	TableSave string `json:"save"` //Имя таблицы где храним текущее состояние
}

//DataBase настройки базы данных postresql
type DataBase struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	DBname   string `json:"dbname"`
}

//Controller настройки имитатора
type Controller struct {
	IP        string        `json:"ip"`
	GuiPort   int           `json:"guiport"`
	Port      int           `json:"port"`
	Step      int           `json:"step"`            //Интервал времени в секундах для расчетов
	KeepAlive time.Duration `json:"time_keep_alive"` //Интервал времени в течении которого должен прийти keepalive от устройства
}
