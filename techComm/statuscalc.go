package techComm

import (
	"fmt"
	"github.com/ruraomsk/ag-server/memDB"
	"github.com/ruraomsk/ag-server/pudge"
	"strings"
	"time"
)

var nowstatus = make(map[string]string)

func setStatusCross() {
	crosses := memDB.CrossesTable.GetAllKeys()
	memDB.TableDevices.Lock()
	memDB.CrossesTable.Lock()
	defer func() {
		memDB.CrossesTable.Unlock()
		memDB.TableDevices.Unlock()
	}()
	for _, key := range crosses {
		cr, _ := memDB.GetCross(key)
		cc, err := memDB.GetController(cr.IDevice)
		if err != nil {
			cr.StatusDevice = 18
			memDB.SetCross(cr)
			continue
		}

		//Вычисляем новый статус

		statusDevice := calcStatus(cc)
		if statusDevice != cr.StatusDevice {
			cr.StatusDevice = statusDevice
		}
		if cr.PK != cc.PK {
			cr.PK = cc.PK
		}
		if cr.CK != cc.CK {
			cr.CK = cc.CK
		}
		if cr.NK != cc.NK {
			cr.NK = cc.NK
		}
		memDB.SetController(cc)
		memDB.SetCross(cr)
		t := 0
		t, statusDevice = calcJournal(cc)
		reg := pudge.Region{Region: cr.Region, Area: cr.Area, ID: cr.ID}
		status, is := nowstatus[reg.ToKey()]
		if !is {
			status = ""
		}
		if strings.Compare(makeMessage(cc, statusDevice), status) != 0 && cc.DK.FDK != 9 {
			if cc.DK.EDK != 1 {
				ChanLog <- pudge.RecLogCtrl{ID: cc.ID, Type: t, Time: time.Now(), LogString: makeMessage(cc, statusDevice)}
				nowstatus[reg.ToKey()] = makeMessage(cc, statusDevice)
			} else {
				nowstatus[reg.ToKey()] = ""
			}
		}
		if memDB.GetControls(statusDevice) {
			w := "Лампы "
			if cc.DK.LDK == 0 {
				w += "исправны "
			} else {
				w += "не исправны "
			}
			w += " Двери "
			if !cc.DK.ODK {
				w += "закрыты "
			} else {
				w += "открыты "
			}
			ChanLog <- pudge.RecLogCtrl{ID: cc.ID, Type: 2, Time: time.Now(), LogString: w}
		}
	}
}
func makeMessage(cc pudge.Controller, statusDevice int) string {
	status := memDB.GetStatus(statusDevice)
	switch statusDevice {
	case 1:
		return fmt.Sprintf("%s ПК=%d СК=%d НК=%d", status, cc.PK, cc.CK, cc.NK)
	case 2:
		return fmt.Sprintf("%s Фаза=%d", status, cc.DK.FDK)
	case 5:
		return fmt.Sprintf("%s Фаза=%d", status, cc.DK.FDK)
	case 27:
		return fmt.Sprintf("%s ПК=%d СК=%d НК=%d", status, cc.PK, cc.CK, cc.NK)
	}
	return fmt.Sprintf("%s", status)
}
