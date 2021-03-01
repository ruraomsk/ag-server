package xcontrol

import (
	"time"

	"github.com/ruraomsk/TLServer/logger"
	"github.com/ruraomsk/ag-server/setup"
)

// type TabCtrl struct {
// 	Mutex sync.Mutex
// }
func calculate() {
	ts, _ := time.ParseDuration(setup.Set.XCtrl.ShiftCtrl)
	time.Sleep(ts)
	logger.Info.Println("calculate")
	m := time.Now().Minute()
	for m%setup.Set.XCtrl.StepDev != 0 {
		m--
	}
	for _, st := range stats {
		st.calculate()
	}
}
