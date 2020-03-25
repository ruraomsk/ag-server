package pudge

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/JanFant/TLServer/logger"
	"github.com/lib/pq"
)

//JSONLog структура для хранения адреса
type JSONLog struct {
	ID          int    `json:"ID"`
	Area        string `json:"area"`
	Region      string `json:"region"`
	Description string `json:"description"`
}

//getJsonLog возвращает json для лога
func getJSONLog(idevice int) []byte {
	// logger.Debug.Printf("region om %d", idevice)
	result := make([]byte, 0)
	mutexCross.Lock()
	defer mutexCross.Unlock()
	for _, c := range crosses {
		if c.IDevice == idevice {
			j := JSONLog{ID: c.ID, Area: strconv.Itoa(c.Area), Region: strconv.Itoa(c.Region), Description: c.Name}
			result, _ := json.Marshal(j)
			return result
		}
	}
	return result
}

// Ведем простое логирование
func writeLog() {
	for {
		ch := <-ChanLog
		w := fmt.Sprintf("insert into public.logdevice (id,tm,crossinfo,txt) values(%d,'%s','%s','%s');",
			ch.ID, string(pq.FormatTimestamp(time.Now())), getJSONLog(ch.ID), ch.LogString)
		_, err = conLog.Exec(w)
		if err != nil {
			logger.Error.Printf("Ошибка записи в БД логгирования %s \n%s", err.Error(), ch.LogString)
		}
		time.Sleep(10 * time.Millisecond)
	}
}
