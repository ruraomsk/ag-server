package xcontrol

import (
	"encoding/json"
	"fmt"

	// "github.com/ruraomsk/ag-server/comm"
	"github.com/ruraomsk/ag-server/logger"
	"github.com/ruraomsk/ag-server/pudge"
)

// type TabCtrl struct {
// 	Mutex sync.Mutex
// }
func workerXTCommand() {
	for {
		cmd := <-pudge.XTCommand
		changeState(cmd.Region, cmd.Command)
	}
}
func calculate() {
	// if !FirstCalculate {
	// 	ts, _ := time.ParseDuration(setup.Set.XCtrl.ShiftCtrl)
	// 	time.Sleep(ts)
	// }
	viewer = true
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
func changeState(region pudge.Region, cmd int) {
	for _, s := range stats {
		if s.State.Region == region.Region && s.State.Area == region.Area && s.State.SubArea == region.ID {
			mainTable.Mutex.Lock()
			defer mainTable.Mutex.Unlock()
			s.State.PKNow = 0
			s.State.PKCalc = 0
			oldRelease := s.State.Release
			oldSwitch := s.State.Switch
			switch cmd {
			case 0:
				s.State.Release = false
				s.State.Switch = false
			case 1:
				s.State.Release = true
				s.State.Switch = true
			case 2:
				s.State.Switch = false
			case 3:
				s.State.Switch = true
			}
			if !(oldRelease == s.State.Release && oldSwitch == s.State.Switch) {
				if s.State.Switch {
					for _, dev := range s.State.Devices {
						commARM <- pudge.CommandARM{ID: dev, User: UserName, Command: 5, Params: 0}
					}
				}
				s.State.PKNow = 0
				js, _ := json.Marshal(s.State)
				w := fmt.Sprintf("UPDATE public.xctrl SET state='%s' WHERE region=%d and  area=%d and subarea=%d;",
					string(js), s.State.Region, s.State.Area, s.State.SubArea)
				_, err := dbb.Exec(w)
				if err != nil {
					logger.Error.Printf("%s %s", w, err.Error())
				}
			}
			//logger.Info.Printf("ХТ %v %d",region,cmd)
			return
		}
	}
	logger.Error.Printf("Нет такого XT %v", region)
}
