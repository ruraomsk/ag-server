package xcontrol

import (
	"encoding/json"
	"fmt"
	"github.com/ruraomsk/ag-server/setup"
	"time"

	"github.com/ruraomsk/TLServer/logger"
	//Инициализатор постргресса
	_ "github.com/lib/pq"
)

type ExtState struct {
	State   State
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
	time := step
	r := make([]LineResult, 0)
	for time <= 60*24 {
		l := new(LineResult)
		l.init(time, size)
		r = append(r, *l)
		time += step
	}
	return r
}
func (e *ExtState) init() {
	e.Results = make(map[string][]LineResult)
	e.Devices = make([]int, 0)
	for _, s := range e.State.Xctrls {
		e.Results[s.Name] = initLineResult(e.State.Step, 3)
	}
	e.Results["result"] = initLineResult(e.State.Step, 2+len(e.State.Xctrls))
}

func (e *ExtState) calculate() {
	logger.Info.Printf(" Управление %d %d %d", e.State.Region, e.State.Area, e.State.SubArea)
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
	logger.Info.Printf(" Смотрим %d:%d для  %d %d %d", e.Time/60, e.Time%60, e.State.Region, e.State.Area, e.State.SubArea)

	for _, r := range e.Results["result"] {
		if r.Time == e.Time {
			for _, x := range e.State.Xctrls {
				x.calculate(e)
			}

		}
	}

}

func ReaderStates() error {
	logger.Info.Printf("Загружаем и настраиваем XT....")
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
		extState.State = state
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
