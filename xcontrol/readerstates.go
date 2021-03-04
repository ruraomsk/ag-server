package xcontrol

import (
	"encoding/json"
	"fmt"
	"github.com/ruraomsk/ag-server/comm"
	"github.com/ruraomsk/ag-server/setup"
	"time"

	"github.com/ruraomsk/TLServer/logger"
	//Инициализатор постргресса
	_ "github.com/lib/pq"
)

type ExtState struct {
	State   *State
	Time    int //Внутреннее время
	Results map[string][]LineResult
	Devices []int
}
type LineResult struct {
	Time  int
	Value []int
	Good  bool
}

func (l *LineResult) init(time int, size int) {
	l.Time = time
	l.Good = false
	l.Value = make([]int, size)
}
func initLineResult(step int, size int) []LineResult {
	ltime := step
	r := make([]LineResult, 0)
	for ltime <= 60*24 {
		l := new(LineResult)
		l.init(ltime, size)
		r = append(r, *l)
		ltime += step
	}
	return r
}
func (e *ExtState) init() {
	e.Results = make(map[string][]LineResult)
	e.Devices = make([]int, 0)
	for _, s := range e.State.Xctrls {
		e.Results[s.Name] = initLineResult(e.State.Step, 3)
	}
	e.Results["result"] = initLineResult(e.State.Step, 2)
}

func (e *ExtState) calculate() {
	//logger.Info.Printf(" Управление %d %d %d", e.State.Region, e.State.Area, e.State.SubArea)
	if !e.State.Switch {
		return
	}
	s := 0
	for _, r := range setup.Set.XCtrl.Regions {
		if r[0] == e.State.Region {
			s = r[1]
			break
		}
	}
	t := time.Now().Add(time.Duration(s) * time.Hour)
	e.Time = t.Hour()*60 + t.Minute()
	//logger.Info.Printf(" Смотрим %d:%d для  %d %d %d", e.Time/60, e.Time%60, e.State.Region, e.State.Area, e.State.SubArea)
	result := e.Results["result"]
	mf := false
	for _, r := range result {
		if r.Time == e.Time {
			e.State.LastTime = e.Time
			for _, x := range e.State.Xctrls {
				x.calculate(e)
			}
			mf = true
		}
	}
	if !mf {
		return
	}
	//Сливаем результаты
	temp := initLineResult(e.State.Step, len(e.State.Xctrls))
	goods := initLineResult(e.State.Step, len(e.State.Xctrls))
	c := 0
	for _, x := range e.State.Xctrls {
		for i, r := range e.Results[x.Name] {
			temp[i].Value[c] = r.Value[2]
			if r.Good {
				goods[i].Value[c] = 1
			}
		}
		c++
	}
	for i, r := range result {
		good := false
		for _, g := range goods[i].Value {
			if g != 0 {
				good = true
			}
		}
		if !good {
			r.Value[0] = 0
			r.Good = false
			result[i] = r
			continue
		}
		r.Good = true
		ir := make([]int, 13)
		for _, v := range temp[i].Value {
			ir[v]++
		}
		if ir[0] == len(temp[i].Value) {
			r.Value[0] = 0
			result[i] = r
			continue
		}
		r.Value[0] = e.getKC(ir)
	}
	e.Results["result"] = result
	for _, r := range result {
		if r.Time == e.Time {
			pk := 0
			for _, p := range e.State.External {
				if p[0] == r.Value[0] {
					pk = p[1]
				}
			}
			e.State.PKCalc = pk
			if e.State.Release {
				//Выслать всем устройствам новый ПК
				for _, dev := range e.Devices {
					commARM <- comm.CommandARM{ID: dev, User: UserName, Command: 5, Params: e.State.PKCalc}
				}
				e.State.PKNow = e.State.PKCalc
			} else {
				if e.State.PKNow != 0 {
					//Выслать всем устройствам команду 0
					for _, dev := range e.Devices {
						commARM <- comm.CommandARM{ID: dev, User: UserName, Command: 5, Params: 0}
					}
					e.State.PKNow = 0
				}
			}
			return
		}
	}

}
func (e *ExtState) getKC(ir []int) int {
	for i := 0; i < 4; i++ {
		t := make([]int, 3)
		t[0] = ir[e.State.Prioryty[i][0]]
		t[1] = ir[e.State.Prioryty[i][1]]
		t[2] = ir[e.State.Prioryty[i][2]]
		if (t[0] + t[1] + t[2]) == 0 {
			continue
		}
		if t[0] >= t[1] && t[0] >= t[2] {
			return e.State.Prioryty[i][0]
		}
		if t[1] >= t[2] {
			return e.State.Prioryty[i][1]
		}
		return e.State.Prioryty[i][2]
	}
	return 0
}

func ReaderStates() error {
	logger.Info.Printf("Загружаем и настраиваем XT....")
	mainTable.Mutex.Lock()
	defer mainTable.Mutex.Unlock()
	stats = make([]ExtState, 0)
	w := "select state from public.xctrl;"
	rows, err := dbb.Query(w)
	if err != nil {
		logger.Error.Printf("Запрос  %s %s", w, err.Error())
		return err
	}
	for rows.Next() {
		var state State
		var stat []byte
		err = rows.Scan(&stat)
		if err != nil {
			logger.Error.Printf("Запрос scan %s", err.Error())
			return err
		}
		err = json.Unmarshal(stat, &state)
		if err != nil {
			logger.Error.Printf("json %s %s", string(stat), err.Error())
			return err
		}
		extState := new(ExtState)
		extState.State = &state
		s := 0
		for _, r := range setup.Set.XCtrl.Regions {
			if r[0] == extState.State.Region {
				s = r[1]
				break
			}
		}
		h := (time.Now().Hour() + s) % 24
		extState.Time = h*60 + time.Now().Minute()
		extState.init()
		w = fmt.Sprintf("select idevice from public.\"cross\" where region=%d and area=%d and subarea=%d;", state.Region, state.Area, state.SubArea)
		devs, err := dbb.Query(w)
		if err != nil {
			logger.Error.Printf("Запрос  %s %s", w, err.Error())
			return err
		}
		for devs.Next() {
			var id int
			err = devs.Scan(&id)
			if err != nil {
				logger.Error.Printf("Запрос scan %s", err.Error())
				return err
			}
			extState.Devices = append(extState.Devices, id)
		}
		_ = devs.Close()
		stats = append(stats, *extState)
	}
	_ = rows.Close()
	return nil
}
