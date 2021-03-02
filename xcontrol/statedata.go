package xcontrol

import (
	"encoding/json"
	"github.com/ruraomsk/TLServer/logger"
	"github.com/ruraomsk/ag-server/pudge"
	"sort"
)

//State описание xctrs
type State struct {
	Region      int        `json:"region"`
	Area        int        `json:"area"`
	SubArea     int        `json:"subarea"`
	Switch      bool       `json:"switch"`  //true призводим расчет нового плана
	Release     bool       `json:"release"` //true выполняем план
	UseStrategy bool       `json:"use"`     //true выполняем стратегию А иначе стратегия B
	Step        int        `json:"step"`    //Время цикла для данного подрайона
	LastTime    int        `json:"ltime"`   //Последний расчет характерной точки
	PKCalc      int        `json:"pkcalc"`  //Посчитанный ПК
	PKNow       int        `json:"pknow"`   //Текущий ПК
	PKLast      int        `json:"pklast"`  //Предыдущий ПК
	Status      []string   `json:"status"`  //Состояние расчетов и итоги проверки
	Xctrls      []Xctrl    `json:"xctrls"`
	External    [12][2]int `json:"ext"`
	Prioryty    [4][3]int
}
type Xctrl struct {
	Name       string       `json:"name"`
	Left       int          `json:"left"`  //Максимум для прямого направления
	Right      int          `json:"right"` //Максимум для обратного направления
	Status     []string     `json:"status"`
	StrategyA  []StrategyA  //Правила перехода по схеме А (области)
	StrategyB  []StrategyB  //Правила перехода по схеме B (лучи)
	Calculates []Calculates //Правила расчета характерной точки

}

func (x *Xctrl) calculate(estate *ExtState) {
	logger.Info.Printf("Расчитываем %d %d %d для %d:%d", estate.State.Region, estate.State.Area, estate.State.SubArea, estate.Time/60, estate.Time%60)
	result := estate.Results[x.Name]
	for i, r := range result {
		start := 0
		if i != 0 {
			start = result[i-1].Time
		}
		for _, c := range x.Calculates {
			good := false
			left := 0
			right := 0
			reg := pudge.Region{Region: c.Region, Area: c.Area, ID: c.ID}
			for _, l := range c.ChanL {
				ll, g := mainTable.getInfo(reg, l, start, r.Time)
				if !g {
					good = false
				}
				left += ll
			}
			for _, rt := range c.ChanR {
				rr, g := mainTable.getInfo(reg, rt, start, r.Time)
				if !g {
					good = false
				}
				right += rr
			}
			r.Good = good
			r.Value[0] = left
			r.Value[1] = right
			result[i] = r
			if r.Time == estate.Time {
				return
			}

		}
	}

	for i, r := range result {
		if r.Time == estate.Time {

			r.Good = true
			result[i] = r
			return
		}
	}
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

//Calculates расчет одной позиции точки
type Calculates struct {
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

func listStates() string {
	res := new(ListTables)
	res.List = make([]pudge.Region, 0)
	for _, x := range stats {
		r := pudge.Region{Area: x.State.Area, Region: x.State.Region, ID: x.State.SubArea}
		res.List = append(res.List, r)
	}
	sort.Slice(res.List, func(i, j int) bool {
		if res.List[i].Region != res.List[j].Region {
			return res.List[i].Region < res.List[j].Region
		}
		if res.List[i].Area != res.List[j].Area {
			return res.List[i].Area < res.List[j].Area
		}
		return res.List[i].ID < res.List[j].ID
	})
	result, err := json.Marshal(res)
	if err != nil {
		logger.Error.Println(err.Error())
	}
	return string(result)
}
func getState(region pudge.Region) string {
	for _, s := range stats {
		if s.State.Region == region.Region && s.State.Area == region.Area && s.State.SubArea == region.ID {
			result, err := json.Marshal(s.State)
			if err != nil {
				logger.Error.Println(err.Error())
				return "{}"
			}
			return string(result)
		}
	}
	return "{}"
}
func getData(region pudge.Region, name string) string {
	for _, s := range stats {
		if s.State.Region == region.Region && s.State.Area == region.Area && s.State.SubArea == region.ID {
			r, is := s.Results[name]
			if !is {
				logger.Error.Printf("Нет такого %v %s", region, name)
				return "{}"
			}
			result, err := json.Marshal(r)
			if err != nil {
				logger.Error.Println(err.Error())
				return "{}"
			}
			//logger.Info.Println(string(result))
			return "{\"datas\":" + string(result) + "}"
		}
	}
	return "{}"

}
