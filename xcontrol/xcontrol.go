package xcontrol

import (
	"github.com/JanFant/TLServer/logger"
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
	Region     int        `json:"region"`
	Area       int        `json:"area"`
	SubArea    int        `json:"subarea"`
	Switch     bool       `json:"switch"`  //true призводим расчет нового плана
	Release    bool       `json:"release"` //true выполняем план
	LastTime   time.Time  `json:"ltime"`   //Последний расчет характерной точки
	PKNow      int        `json:"pknow"`   //Текущий ПК
	PKLast     int        `json:"pklast"`  //Предыдущий ПК
	XNumber    int        `json:"xnum"`    //Характерное число текущее
	Status     []string   `json:"status"`  //Состояние расчетов и итоги проверки
	Strategys  []Strategy //Правила перехода
	Calculates []Calc     //Правила расчета характерной точки

}

//Strategy описание стратегии
type Strategy struct {
	XLeft  int `json:"xleft"`  //Некое число для смены плана >=
	XRight int `json:"xright"` //Некое число для смены плана <
	PK     int `json:"pk"`     // Назначенный план
}

//Calc расчет одной позиции точки
type Calc struct {
	Region int     `json:"region"`
	Area   int     `json:"area"`
	ID     int     `json:"id"`   //Перекресток по которому принимается решение
	Chanal int     `json:"chan"` //Номер канала по статистике
	Mult   float32 `json:"mult"` //Коэффицент приведения
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
	if !setup.Set.XCtrl.Calculate {
		//Не нужен расчет управления по характерным точкам
		logger.Info.Print("Модуль расчета характерных точек отключен... ")
		return
	} else {
		logger.Info.Print("Модуль расчета характерных точек запущен... ")
		go Calculator()
	}
	logger.Info.Print("Модуль управления по характерным точкам запущен... ")
	go Sender()
	select {
	case <-context.Done():
		return
	case <-stop:
		return
	}

}
