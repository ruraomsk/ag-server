package xcontrol

import (
	"encoding/json"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/ruraomsk/ag-server/logger"
	"github.com/ruraomsk/ag-server/pudge"
	"github.com/ruraomsk/ag-server/setup"
)

type Table struct {
	Table map[string]*Xcross
	Seek  map[int]string
	Mutex sync.Mutex
}
type Xcross struct {
	Region  pudge.Region
	IDevice int
	Step    int
	Count   int
	Values  map[int]*Value
}
type Value struct {
	Def     bool
	Chanels []int
	Good    []bool
}
type ListTables struct {
	List []pudge.Region `json:"ls"`
}

func (t *Table) listTables() string {
	//logger.Info.Println("listTables")
	t.Mutex.Lock()
	defer t.Mutex.Unlock()
	res := new(ListTables)
	res.List = make([]pudge.Region, 0)
	for _, x := range t.Table {
		res.List = append(res.List, x.Region)
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
	//logger.Info.Println(string(result))
	return string(result)
}

type LineCross struct {
	Region   pudge.Region
	DiffTime int
	Step     int
	Count    int
	Values   []LineValue
}
type LineValue struct {
	Time    int
	Def     bool
	Chanels []int
	Good    []bool
}

func (t *Table) getXCross(region pudge.Region) string {
	t.Mutex.Lock()
	defer t.Mutex.Unlock()
	res := new(LineCross)
	for _, r := range setup.Set.XCtrl.Regions {
		if r[0] == region.Region {
			res.DiffTime = r[1]
			break
		}
	}
	xcr, is := t.Table[region.ToKey()]
	if is {
		res.Region = xcr.Region
		res.Step = xcr.Step
		res.Count = xcr.Count
		res.Values = make([]LineValue, 0)
		for t, v := range xcr.Values {
			r := new(LineValue)
			r.Time = t
			r.Def = v.Def
			r.Chanels = v.Chanels
			r.Good = v.Good
			res.Values = append(res.Values, *r)
		}
		sort.Slice(res.Values, func(i, j int) bool { return res.Values[i].Time < res.Values[j].Time })
	}
	result, err := json.Marshal(res)
	if err != nil {
		logger.Error.Println(err.Error())
	}
	return string(result)

}
func (t *Table) getInfo(region pudge.Region, chanel int, start int, stop int) (value int, good bool) {
	//t.Mutex.Lock()
	//defer t.Mutex.Unlock()
	xcross, is := t.Table[region.ToKey()]
	if !is {
		return -1, false
	}
	if chanel > xcross.Count {
		return -2, false

	}
	sum := 0
	start += xcross.Step
	for start <= stop {
		v, is := xcross.Values[start]
		if !is {
			return 0, false
		}
		if !v.Def {
			return 0, false
		}
		if (chanel-1) < 0 || (chanel-1) >= len(v.Good) {
			return 0, false
		}
		if !v.Good[chanel-1] {
			return 0, false
		}
		sum += v.Chanels[chanel-1]
		start += xcross.Step
	}
	return sum, true
}
func getDefStat(cross *pudge.Cross) (is bool, period int, count int) {
	is = false
	for _, ds := range cross.Arrays.StatDefine.Levels {
		if ds.Period == 0 {
			return false, 0, 0
		}
		if ds.Count > 0 || ds.Ninput > 0 {
			//logger.Debug.Printf("Перекресток %d %d %d имеет статистику",cross.Region,cross.Area,cross.ID)

			if ds.Count > 16 || ds.Count == 0 {
				return true, ds.Period, 16
			}
			return true, ds.Period, ds.Count
		}
	}
	//logger.Debug.Printf("Перекресток %d %d %d не имеет статистику",cross.Region,cross.Area,cross.ID)
	return false, 0, 0
}
func newXcross(cross *pudge.Cross) *Xcross {
	var is bool
	x := new(Xcross)
	is, x.Step, x.Count = getDefStat(cross)
	if !is {
		return nil
	}
	x.Region = pudge.Region{
		Region: cross.Region,
		Area:   cross.Area,
		ID:     cross.ID,
	}
	x.IDevice = cross.IDevice
	x.Values = make(map[int]*Value)
	t := 0
	for t < 1440 {
		v := new(Value)
		v.Chanels = make([]int, x.Count)
		v.Good = make([]bool, x.Count)
		x.Values[t] = v
		t += x.Step
	}
	//logger.Debug.Printf("Заполнил %s",x.Region.ToKey())
	return x
}
func clearRegion(region int) {
	//logger.Info.Println("clearRegion ", region)
	mainTable.Mutex.Lock()
	defer mainTable.Mutex.Unlock()
	for _, t := range mainTable.Table {
		if t.Region.Region != region {
			continue
		}
		for _, v := range t.Values {
			v.Good = make([]bool, len(v.Good))
			v.Def = false
			v.Chanels = make([]int, len(v.Chanels))
		}
	}
	for i, e := range stats {
		if stats[i].State.Region != region {
			continue
		}
		stats[i].State.Time = 0
		for _, s := range stats[i].State.Xctrls {
			stats[i].State.Results[s.Name] = initLineResult(e.State.Step, 3)
		}
		e.State.Results["result"] = initLineResult(e.State.Step, 2)
	}

}
func (t *Table) setXCross(xcross *Xcross) {
	//t.Mutex.Lock()
	//defer t.Mutex.Unlock()

	_, is := t.Table[xcross.Region.ToKey()]
	if is {
		logger.Error.Printf("Дубликат %v", xcross.Region)
	} else {
		t.Table[xcross.Region.ToKey()] = xcross
	}
	_, is = t.Seek[xcross.IDevice]
	if is {
		logger.Error.Printf("Дубликат %d", xcross.IDevice)
	} else {
		t.Seek[xcross.IDevice] = xcross.Region.ToKey()
	}
	//logger.Debug.Printf("сохранил %s",xcross.Region.ToKey())
}
func (t *Table) setData(device *pudge.Controller) error {
	if !work {
		return fmt.Errorf("not work")
	}
	//logger.Info.Printf("Записываем статистику %d",device.ID)
	//t.Mutex.Lock()
	//defer t.Mutex.Unlock()
	r, is := t.Seek[device.ID]
	if !is {
		return fmt.Errorf("нет устройства %d", device.ID)
	}
	reg := pudge.FromKeyToRegion(r)
	stats := getStatistics(reg)
	//logger.Info.Print("Есть %s ",r )
	xcross := *t.Table[r]

	for _, s := range stats {
		t := s.Hour*60 + s.Min
		//if t == 0 {
		//	continue
		//}
		v, is := xcross.Values[t]
		if !is {
			continue
		}
		v.Def = true
		for _, d := range s.Datas {
			if d.Chanel <= xcross.Count {
				v.Chanels[d.Chanel-1] = d.Intensiv
				if d.Status == 0 {
					v.Good[d.Chanel-1] = true
				}
			}
		}
		//logger.Info.Print(" %v ",v )

	}
	t.Table[r] = &xcross
	return nil
}
func makeTable() error {
	//logger.Info.Println("makeTable start")
	mainTable.Mutex.Lock()
	defer mainTable.Mutex.Unlock()
	mainTable.Table = make(map[string]*Xcross)
	mainTable.Seek = make(map[int]string)
	//dbb.Exec("begin work ;")
	//dbb.Exec("lock table  public.\"cross\" in exclusive;")
	//defer func() {
	//	dbb.Exec("ccommit work;")
	//}()
	w := "select region,area,id,idevice,state from public.\"cross\";"
	crosses, err := dbb.Query(w)
	if err != nil {
		logger.Error.Printf("Запрос  %s %s", w, err.Error())
		return err
	}
	defer crosses.Close()
	for crosses.Next() {
		var cr pudge.Cross
		var region, area, id, idevice int
		var stat []byte
		err = crosses.Scan(&region, &area, &id, &idevice, &stat)
		if err != nil {
			logger.Error.Printf("Запрос scan %s", err.Error())
			return err
		}

		err = json.Unmarshal(stat, &cr)
		if err != nil {
			logger.Error.Printf("json %s %s", string(stat), err.Error())
			return err
		}
		if region != cr.Region || area != cr.Area || id != cr.ID || idevice != cr.IDevice {
			logger.Error.Printf("в БД %d %d %d %d != %d %d %d %d", region, area, id, idevice,
				cr.Region, cr.Area, cr.ID, cr.IDevice)
		}
		xcross := newXcross(&cr)
		if xcross != nil {
			mainTable.setXCross(xcross)
		}
	}
	mainTable.Mutex.Unlock()
	loadTable()
	err = ReaderStates()
	mainTable.Mutex.Lock()
	if err != nil {
		logger.Error.Printf("Контроль управления  %s", err.Error())
		return err
	}
	//logger.Info.Println("makeTable end")
	return nil
}
func loadTable() {
	//logger.Info.Println("loadTable start")
	// ts, _ := time.ParseDuration(setup.Set.XCtrl.ShiftDevice)
	// time.Sleep(ts)
	for time.Now().Second() < 45 {
		time.Sleep(time.Second)
	}
	mainTable.Mutex.Lock()
	defer mainTable.Mutex.Unlock()
	//logger.Info.Println("loadTable")
	w := "select device from public.\"devices\";"
	devs, err := dbb.Query(w)
	if err != nil {
		logger.Error.Printf("Запрос  %s %s", w, err.Error())
		return
	}
	defer devs.Close()
	for devs.Next() {
		var dev pudge.Controller
		var stat []byte
		err = devs.Scan(&stat)
		if err != nil {
			logger.Error.Printf("Запрос scan %s", err.Error())
			return
		}
		err = json.Unmarshal(stat, &dev)
		if err != nil {
			logger.Error.Printf("json %s %s", string(stat), err.Error())
			return
		}
		err = mainTable.setData(&dev)
		if err != nil {
			addMessage(fmt.Sprintf("загрузка %s", err.Error()))
		}
	}
	//logger.Info.Println("loadTable end")
}
func getStatistics(reg pudge.Region) []pudge.Statistic {
	w := fmt.Sprintf("select stat from public.statistics where date='%s' and region=%d and area=%d and id=%d;",
		reg.LocalTime().Format("2006-01-02"), reg.Region, reg.Area, reg.ID)
	var state pudge.ArchStat
	rows, _ := dbb.Query(w)
	for rows.Next() {
		var buf []byte
		rows.Scan(&buf)
		json.Unmarshal(buf, &state)
		rows.Close()
		return state.Statistics
	}
	rows.Close()
	return make([]pudge.Statistic, 0)
}
