package setup

//Setup общая структура для настройки всей системы
type Setup struct {
	DataBase   DataBase
	Server     Server
	Pudge      Pudge
	CommServer CommServer
	Controller Controller
}

//CommServer настройки для сервера коммуникации
type CommServer struct {
	Port int `json:"port"` //Стартовый номер порта на прием
}

//Server настройки для сервера армов
type Server struct {
	Port int `json:"port"` //Стартовый номер порта на прием

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
	Step int `json:"step"` //Интервал времени в секундах для расчетов
}
