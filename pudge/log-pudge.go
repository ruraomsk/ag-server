package pudge

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/lib/pq"
	"github.com/ruraomsk/ag-server/logger"
)

type killRecord struct {
	Time   time.Time
	Region int
	ID     int
}

var lastWrite = time.Now()

func GetRegion(idevice int) Region {
	mutexCtrl.Lock()
	defer mutexCtrl.Unlock()
	for _, c := range crosses {
		if c.IDevice == idevice {
			return Region{Region: c.Region, Area: c.Area, ID: c.ID}

		}
	}
	return Region{Region: 0, Area: 0, ID: 0}
}

type varLog struct {
	stop   bool
	tJ     Journal
	dJ     Journal
	txt    string
	wtJ    bool
	wdJ    Journal
	kills  []killRecord
	isKill bool
}

var mapLogs map[string]*varLog

func workTechJournal(ch LogRecord, cr Region) {
	t := mapLogs[cr.ToKey()]
	if CompareTech(&t.tJ, &ch.Journal) {
		return
	}
	if strings.HasSuffix(ch.Journal.Rezim, " ДУ") {
		return
	}
	if !t.wtJ && ch.Journal.Status == "ПЕРЕХОД" {
		return
	}
	t.tJ = ch.Journal
	mapLogs[cr.ToKey()] = t
	writeLogDB(ch, 0)
	if t.wtJ && ch.Journal.Status == "ПЕРЕХОД" {
		t.kills = append(t.kills, killRecord{Time: ch.Time, Region: cr.Region, ID: ch.ID})
		return
	}
	if t.wtJ && ch.Journal.Status == "НОРМ" {
		clearKillRecords(t.kills)
		t.wtJ = false
		return
	}
	if len(ch.Journal.Arm) != 0 {
		t.wtJ = true
		t.kills = make([]killRecord, 0)
	}
}
func workDeviceJournal(ch LogRecord, cr Region) {
	d := mapLogs[cr.ToKey()]
	if CompareDevice(&d.dJ, &ch.Journal) {
		return
	}
	d.dJ = ch.Journal
	mapLogs[cr.ToKey()] = d
	writeLogDB(ch, 1)
	if ch.Journal.Status == "НОРМ" && d.isKill {
		if d.dJ.Rezim == d.wdJ.Rezim && d.dJ.Phase == d.wdJ.Phase {
			clearKillRecords(d.kills)
			d.isKill = false
			cross, is := GetCross(cr)
			if is {
				// logger.Debug.Printf("Правим арм %v %s", cr, d.wdJ.Arm)
				if d.wdJ.Rezim == "КУ" {
					cross.Arm = ""
				} else {
					cross.Arm = d.wdJ.Arm
				}
				SetCross(cross)
			}
		} else {
			d.kills = append(d.kills, killRecord{Time: ch.Time, Region: cr.Region, ID: ch.ID})
		}
	} else {
		if d.isKill {
			d.kills = append(d.kills, killRecord{Time: ch.Time, Region: cr.Region, ID: ch.ID})
		}
	}
	if len(ch.Journal.Arm) != 0 && ch.Journal.Arm != "Сервер" {
		d.isKill = true
		d.wdJ = Journal{Rezim: ch.Journal.Rezim, Phase: ch.Journal.Phase, Arm: ch.Journal.Arm}
		if d.wdJ.Phase == "ЛР" {
			d.wdJ.Rezim = "ЛУ"
		}
		d.kills = make([]killRecord, 0)
	}
}

// Ведем простое логирование
func writeLog() {
	mapLogs = make(map[string]*varLog)
	var cr Region
	for {
		ch := <-ChanLog
		cr = ch.Region
		if cr.Region == 0 {
			cr = GetRegion(ch.ID)
		}
		for cr.Region == 0 {
			logger.Error.Printf("Устройство %d не привязано", ch.ID)
			Reload <- 1
			time.Sleep(2 * time.Second)
			cr = GetRegion(ch.ID)
			continue
		}
		ch.Region = cr
		// logger.Debug.Printf("%v %v", cr, ch)
		_, is := mapLogs[cr.ToKey()]
		if !is {
			l := new(varLog)
			l.stop = false
			l.kills = make([]killRecord, 0)
			l.isKill = false
			mapLogs[cr.ToKey()] = l
		}
		if mapLogs[cr.ToKey()].stop {
			continue
		}
		switch ch.Type {
		case 0:
			workTechJournal(ch, cr)
		case 1:
			workDeviceJournal(ch, cr)
		case 2:
			if strings.Compare(mapLogs[cr.ToKey()].txt, ch.LogString) != 0 {
				writeLogDB(ch, 2)
				mapLogs[cr.ToKey()].txt = ch.LogString
			}
		case 3:
			writeLogDB(ch, 1)
		}
	}
}
func clearKillRecords(recs []killRecord) {
	for _, r := range recs {
		w := fmt.Sprintf("delete from public.logdevice where id=%d and region=%d and tm='%s';", r.ID, r.Region, string(pq.FormatTimestamp(r.Time)))
		clearLog.Exec(w)
		// logger.Debug.Print(w)
	}
}
func CompareDevice(d1 *Journal, d2 *Journal) bool {
	if d1.Status == d2.Status && d1.Arm == d2.Arm && d1.Rezim == d2.Rezim && d1.Device == d2.Device {
		if d1.Rezim == "ДУ" || d1.Rezim == "РУ" {
			return d1.Phase == d2.Phase
		}
	}
	return d1.Status == d2.Status && d1.Arm == d2.Arm && d1.Rezim == d2.Rezim && d1.Device == d2.Device
}
func CompareTech(d1 *Journal, d2 *Journal) bool {
	return d1.Rezim == d2.Rezim && d1.Arm == d2.Arm && d1.PK == d2.PK && d1.CK == d2.CK && d1.NK == d2.NK
}

type crossInfo struct {
	Region      string `json:"region"`
	Area        string `json:"area"`
	ID          int    `json:"ID"`
	Type        int    `json:"type"`
	Description string `json:"description"`
}

func writeLogDB(ch LogRecord, tup int) {
	if ch.Time == lastWrite {
		ch.Time = ch.Time.Add(1 * time.Microsecond)
	}
	lastWrite = ch.Time
	reg := ch.Region
	cross, is := GetCross(reg)
	if !is {
		logger.Error.Printf("Нет такого %v", reg)
		return
	}
	ci := crossInfo{Region: strconv.Itoa(reg.Region), Area: strconv.Itoa(reg.Area), ID: reg.ID, Description: cross.Name, Type: tup}
	cit, _ := json.Marshal(ci)
	if tup == 2 {
		w := fmt.Sprintf("insert into public.logdevice (region,id,tm,crossinfo,txt,journal) values(%d,%d,'%s','%s','%s','%s');",
			reg.Region, ch.ID, string(pq.FormatTimestamp(ch.Time)), string(cit), ch.LogString, "{}")
		_, err = conLog.Exec(w)
		if err != nil {
			logger.Error.Printf("Ошибка записи в БД логирования %s \n%s", err.Error(), w)
		}
		return
	}
	result, _ := json.Marshal(ch.Journal)
	w := fmt.Sprintf("insert into public.logdevice (region,id,tm,crossinfo,txt,journal) values(%d,%d,'%s','%s','%s','%s');",
		reg.Region, ch.ID, string(pq.FormatTimestamp(ch.Time)), string(cit), "", string(result))
	// logger.Debug.Println(w)
	_, err = conLog.Exec(w)
	if err != nil {
		logger.Error.Printf("Ошибка записи в БД логирования %s \n%s", err.Error(), w)
	}
}
