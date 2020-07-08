package xcontrol

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/JanFant/TLServer/logger"
	"github.com/ruraomsk/ag-server/comm"
	"github.com/ruraomsk/ag-server/setup"

	//Инициализатор постргресса
	_ "github.com/lib/pq"
)

//Данный пакет производит управление по характерным точкам
// Разбит на два раздела
// 	в первом разделе производится расчет характерной точки и выбор стратегии
// 	во втором разделе производится выполнение выбранной стратегии для каждого района и подрайона

//CommonState Общая структура управления расчетами
type CommonState struct {
	StateSubAreas []StateSubArea
}

//StateSubArea описание выбранной стратегии для одного подрайона
type StateSubArea struct {
	Region     int        `json:"region"`
	Area       int        `json:"area"`
	SubArea    int        `json:"subarea"`
	Switch     bool       `json:"switch"` //true призводим расчет нового плана
	LastTime   time.Time  `json:"ltime"`  //Последний расчет характерной точки
	PKNow      int        `json:"pknow"`  //Текущий ПК
	PKLast     int        `json:"pklast"` //Предыдущий ПК
	XNumber    int        `json:"xnum"`   //Характерное число текущее
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

//Sender Посылает новые планы координации на устройства
func Sender() error {
	logger.Info.Printf("Управление по характерным точка стадия 2....")
	dbinfo := fmt.Sprintf("host=%s user=%s password=%s dbname=%s sslmode=disable",
		setup.Set.DataBase.Host, setup.Set.DataBase.User,
		setup.Set.DataBase.Password, setup.Set.DataBase.DBname)
	conDB, err := sql.Open("postgres", dbinfo)
	if err != nil {
		logger.Error.Printf("Запрос на открытие %s %s", dbinfo, err.Error())
		return err
	}
	defer conDB.Close()
	if err = conDB.Ping(); err != nil {
		logger.Error.Printf("Ping %s", err.Error())
		return err
	}
	for true {
		time.Sleep(time.Duration(setup.Set.XCtrl.StepSend) * time.Second)
		_, err = conDB.Exec("begin;")
		if err != nil {
			logger.Error.Printf("Запрос begin %s", err.Error())
			return err
		}
		defer conDB.Exec("rollback;")
		comms := make([]comm.CommandARM, 0)
		w := "select region,area,subarea,switch,pknow,pklast,xnum,strat,calc from public.xctrl;"
		rows, err := conDB.Query(w)
		if err != nil {
			logger.Error.Printf("Запрос  %s %s", w, err.Error())
			return err
		}
		for rows.Next() {
			var v StateSubArea
			err = rows.Scan(&v.Region, &v.Area, &v.SubArea, &v.Switch, &v.PKNow, &v.PKLast)
			if err != nil {
				logger.Error.Printf("Запрос scan %s", err.Error())
				return err
			}

			if v.PKNow != v.PKLast {
				w = fmt.Sprintf("select idevice from public.cross where region = %d and area=%d and subarea = %d;", v.Region, v.Area, v.SubArea)
				cross, err := conDB.Query(w)
				if err != nil {
					logger.Error.Printf("Запрос  %s %s", w, err.Error())
					return err
				}
				for cross.Next() {
					var idevice int
					err = cross.Scan(&idevice)
					c := comm.CommandARM{ID: idevice, User: "XCtrl", Command: 5, Params: v.PKNow}
					comms = append(comms, c)
				}
				w = fmt.Sprintf("update public.xctrl set pklast=%d where region=%d and area=%d and subarea=%d;", v.PKNow, v.Region, v.Area, v.SubArea)
			}
		}

		_, err = conDB.Exec("commit;")
		if err != nil {
			logger.Error.Printf("Запрос commit %s", err.Error())
			return err
		}

	}
	return nil
}
