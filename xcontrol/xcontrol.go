package xcontrol

import (
	"github.com/ruraomsk/TLServer/logger"
	//Инициализатор постргресса
	_ "github.com/lib/pq"
	"github.com/ruraomsk/ag-server/extcon"
	"github.com/ruraomsk/ag-server/setup"
	"time"
)

//Данный пакет производит управление по характерным точкам
// Разбит на два раздела
// 	в первом разделе производится расчет характерной точки и выбор стратегии
// 	во втором разделе производится выполнение выбранной стратегии для каждого района и подрайона

//StateSubArea описание выбранной стратегии для одного подрайона
type State struct {
	Region      int       `json:"region"`
	Area        int       `json:"area"`
	SubArea     int       `json:"subarea"`
	Switch      bool      `json:"switch"`  //true призводим расчет нового плана
	Release     bool      `json:"release"` //true выполняем план
	UseStrategy bool      `json:"use"`     //true выполняем стратегию А иначе стратегия B
	Step        int       `json:"step"`    //Время цикла для данного подрайона
	Remain      int       `json:"rem"`     //Остаток времени для нового расчета
	LastTime    time.Time `json:"ltime"`   //Последний расчет характерной точки
	PKCalc      int       `json:"pkcalc"`  //Расчитанный ПК
	PKNow       int       `json:"pknow"`   //Текущий ПК
	PKLast      int       `json:"pklast"`  //Предыдущий ПК
	Status      []string  `json:"status"`  //Состояние расчетов и итоги проверки

	Left  int `json:"left"`  //Максимум для прямого направления
	Right int `json:"right"` //Максимум для обратного направления

	StrategysA []StrategyA //Правила перехода по схеме А (области)
	StrategysB []StrategyB //Правила перехода по схеме B (лучи)
	Calculates []Calc      //Правила расчета характерной точки
	Results    []Result    //Промежуточные результаты
}
type Result struct {
	Ileft  int `json:"il"` //Интенсивность прямого направления
	Iright int `json:"ir"` //Интенсивность обратного направления
}

//StrategyB описание стратегии
type StrategyB struct {
	XLeft       int     `json:"xleft"`  //Интенсивность в прямом направлении
	XRight      int     `json:"xright"` //Интенсивность в обратном направлении
	VLeft       float32 `json:"vleft"`  //Луч левый
	VRight      float32 `json:"vright"` //Луч правый
	PKL         int     `json:"pkl"`    // Назначенный план прямой
	PKS         int     `json:"pks"`    // Назначенный план средний
	PKR         int     `json:"pkr"`    // Назначенный план обратный
	Description string  `json:"desc"`   //Описание
}

//StrategyA описание стратегии
type StrategyA struct {
	XLeft       int    `json:"xleft"`  //Некое число для центра области
	XRight      int    `json:"xright"` //Некое число для центра области
	PK          int    `json:"pk"`     // Назначенный план
	Description string `json:"desc"`   //Описание
}

//Calc расчет одной позиции точки
type Calc struct {
	Region int   `json:"region"`
	Area   int   `json:"area"`
	ID     int   `json:"id"`    //Перекресток по которому принимается решение
	ChanL  []int `json:"chanL"` //Номера каналов по статистике прямой направление
	ChanR  []int `json:"chanR"` //Номера каналов по статистике обратное направление
}

type key struct {
	Region  int `json:"region"`
	Area    int `json:"area"`
	SubArea int `json:"subarea"`
}

//Start главный модуль инспектора
func Start(context *extcon.ExtContext, stop chan int) {

	if !setup.Set.XCtrl.Switch {
		//Не нужен модель управления по характерным точкам
		logger.Info.Print("Модуль управления по характерным точкам отключен... ")
		return
	}
	err := Corrector()
	if err != nil {
		logger.Error.Printf("Контроль управленя  %s", err.Error())
		//logger.Info.Print("Модуль управления по характерным точкам будет отключен!")
		//return
	}
	logger.Info.Print("Модуль расчета характерных точек запущен... ")
	go Calculator()
	logger.Info.Print("Модуль управления по характерным точкам запущен... ")
	go Sender()
	select {
	case <-context.Done():
		return
	case <-stop:
		return
	}

}
