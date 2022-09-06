package xcontrol

import (
	"encoding/json"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/ruraomsk/ag-server/logger"
)

type messError struct {
	Time    time.Time
	Message string
}
type Messages struct {
	Messages []string
}

var listDevice []messError
var errMutex sync.Mutex

func clearError() {
	errMutex.Lock()
	listDevice = make([]messError, 0)
	errMutex.Unlock()
}
func addMessage(message string) {
	errMutex.Lock()
	defer errMutex.Unlock()
	for i, mes := range listDevice {
		if strings.Compare(message, mes.Message) == 0 {
			listDevice[i].Time = time.Now()
			return
		}
	}
	listDevice = append(listDevice, messError{Time: time.Now(), Message: message})
}
func getMessages() string {
	ms := new(Messages)
	ms.Messages = make([]string, 0)
	for _, m := range listDevice {
		ms.Messages = append(ms.Messages, m.Time.Format("15:04:05")+";"+m.Message)
	}
	sort.Slice(ms.Messages, func(i, j int) bool {
		return strings.Compare(ms.Messages[i], ms.Messages[j]) < 0
	})
	js, err := json.Marshal(ms)
	if err != nil {
		logger.Error.Println(err.Error())
		return "{}"
	}
	return string(js)
}
