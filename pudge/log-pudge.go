package pudge

import (
	"fmt"
	"time"

	"github.com/JanFant/TLServer/logger"
	"github.com/lib/pq"
)

// Ведем простое логирование
func writeLog() {
	for {
		ch := <-ChanLog
		w := fmt.Sprintf("insert into public.logdevice (id,tm,txt) values(%d,'%s','%s');",
			ch.ID, string(pq.FormatTimestamp(time.Now())), ch.LogString)
		_, err = conLog.Exec(w)
		if err != nil {
			logger.Error.Printf("Ошибка записи в БД логгирования %s \n%s", err.Error(), ch.LogString)
		}
	}
}
func (c *Controller) setLogString() {

}
