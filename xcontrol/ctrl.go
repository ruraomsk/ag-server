package xcontrol

import (
	"github.com/ruraomsk/ag-server/setup"
	"time"
)

// type TabCtrl struct {
// 	Mutex sync.Mutex
// }
func calculate() {
	if !FirstCalculate {
		ts, _ := time.ParseDuration(setup.Set.XCtrl.ShiftCtrl)
		time.Sleep(ts)
	}
	//logger.Info.Println("calculate")
	//m := time.Now().Minute()
	//for m%setup.Set.XCtrl.StepDev != 0 {
	//	m--
	//}
	mainTable.Mutex.Lock()
	defer mainTable.Mutex.Unlock()
	for _, st := range stats {
		st.calculate()
	}
}
