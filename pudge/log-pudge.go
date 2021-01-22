package pudge

import (
	"encoding/json"
	"fmt"
	"github.com/JanFant/TLServer/logger"
	"github.com/lib/pq"
	"strconv"
	"strings"
)

//JSONLog структура для хранения адреса
type JSONLog struct {
	ID          int    `json:"ID"`
	Area        string `json:"area"`
	Region      string `json:"region"`
	Type        int    `json:"type"`
	Description string `json:"description"`
}

var mapMessages map[string]map[int]string

func getCross(idevice int) *Cross {
	// logger.Debug.Printf("region om %d", idevice)
	mutexCtrl.Lock()
	defer mutexCtrl.Unlock()
	for _, c := range crosses {
		if c.IDevice == idevice {
			//mutexCross.Unlock()
			//mutexCtrl.Unlock()
			return c
		}
	}
	//mutexCross.Unlock()
	//mutexCtrl.Unlock()
	return nil
}

// Ведем простое логирование
func writeLog() {
	mapMessages = make(map[string]map[int]string)

	for {
		ch := <-ChanLog
		cr := getCross(ch.ID)
		if cr == nil {
			logger.Error.Printf("error %v", ch)
			continue
		}
		//if ch.Type != 2 {
		//	//logger.Info.Printf("log message %v", ch)
		//}
		reg := Region{cr.Region, cr.Area, cr.ID}
		crm, is := mapMessages[reg.ToKey()]
		if !is {
			news := make(map[int]string)
			news[0] = ""
			news[1] = ""
			news[2] = ""
			mapMessages[reg.ToKey()] = news
			crm = mapMessages[reg.ToKey()]
		}
		switch ch.Type {
		case 0:
			if strings.Compare(crm[0], ch.LogString) != 0 {
				writeLogDB(cr, ch, 0)
				crm[0] = ch.LogString
			}
		case 1:
			if strings.Compare(crm[1], ch.LogString) != 0 {
				writeLogDB(cr, ch, 1)
				crm[1] = ch.LogString
			}
		case 2:
			if strings.Compare(crm[2], ch.LogString) != 0 {
				writeLogDB(cr, ch, 2)
				crm[2] = ch.LogString
			}
		case -1:
			if strings.Contains(ch.LogString, "Координированное управление ПК=") {
				if strings.Compare(crm[1], "Координированное управление") != 0 {
					temp := ch.LogString
					ch.LogString = "Координированное управление"
					writeLogDB(cr, ch, 1)
					crm[1] = ch.LogString
					ch.LogString = temp
				}
				if strings.Compare(crm[0], ch.LogString) != 0 {
					writeLogDB(cr, ch, 0)
					crm[0] = ch.LogString
				}
				break
			}
			if strings.Compare(crm[0], ch.LogString) != 0 {
				writeLogDB(cr, ch, 0)
				crm[0] = ch.LogString
			}
			if strings.Compare(crm[1], ch.LogString) != 0 {
				writeLogDB(cr, ch, 1)
				crm[1] = ch.LogString
			}

		}
		mapMessages[reg.ToKey()] = crm
		//time.Sleep(100 * time.Millisecond)
	}
}
func writeLogDB(cr *Cross, ch RecLogCtrl, tup int) {
	j := JSONLog{Region: strconv.Itoa(cr.Region), Area: strconv.Itoa(cr.Area), ID: cr.ID, Description: cr.Name, Type: tup}
	result, _ := json.Marshal(j)
	w := fmt.Sprintf("insert into public.logdevice (id,tm,crossinfo,txt) values(%d,'%s','%s','%s');",
		ch.ID, string(pq.FormatTimestamp(ch.Time)), result, ch.LogString)
	_, err = conLog.Exec(w)
	if err != nil {
		logger.Error.Printf("Ошибка записи в БД логгирования %s \n%s", err.Error(), w)
	}
}
