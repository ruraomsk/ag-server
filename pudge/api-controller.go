package pudge

import (
	"fmt"
	"strings"
	"time"
)

//IsConnected возвращает на связи ли устройство
func (cc *Controller) IsConnected() bool {
	return cc.StatusConnection
}

//IsRegistred проверяет зарегистрирован ли Id на перекрестке
func IsRegistred(id int) *Region {
	mutexCtrl.Lock()
	defer mutexCtrl.Unlock()
	for _, c := range crosses {
		if c.IDevice == id {
			reg := Region{Region: c.Region, Area: c.Area, ID: c.ID}
			return &reg
		}
	}
	return nil
}

var cCount int

func setStatusCross() {
	mutexCtrl.Lock()
	defer mutexCtrl.Unlock()
	// logger.Debug.Print("setStatusCross start")
	//result := make(map[string]*Cross, 0)
	for _, cr := range crosses {
		cc, is := controllers[cr.IDevice]

		if !is {
			continue
		}
		//Вычисляем новый статус

		statusDevice := cc.calcStatus()
		if statusDevice != cr.StatusDevice {
			cr.StatusDevice = statusDevice
			cr.WriteToDB = true
		}
		if cr.PK != cc.PK {
			cr.PK = cc.PK
			cr.WriteToDB = true
		}
		if cr.CK != cc.CK {
			cr.CK = cc.CK
			cr.WriteToDB = true
		}
		if cr.NK != cc.NK {
			cr.NK = cc.NK
			cr.WriteToDB = true
		}
		//if cr.WriteToDB {
		//	reg := Region{cr.Region, cr.Area, cr.ID}
		//	result[reg.ToKey()] = cr
		//}
		t := 0
		t, statusDevice = cc.calcJournal()
		//statusDevice=cc.getSource()+statusDevice
		reg := Region{cr.Region, cr.Area, cr.ID}
		status, is := nowstatus[reg.ToKey()]
		if !is {
			status = ""
		}
		if strings.Compare(makeMessage(cc, statusDevice), status) != 0 && cc.DK.FDK != 9 {
			if cc.coderr() != 1 {
				ChanLog <- RecLogCtrl{ID: cc.ID, Type: t, Time: time.Now(), LogString: makeMessage(cc, statusDevice)}
				nowstatus[reg.ToKey()] = makeMessage(cc, statusDevice)
			} else {
				nowstatus[reg.ToKey()] = ""
			}
		}
		if controls[statusDevice] {
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
			ChanLog <- RecLogCtrl{ID: cc.ID, Type: 2, Time: time.Now(), LogString: w}
		}
	}
	//crosses=make(map[string]*Cross)
	//for key, cr := range result {
	//	crosses[key] = cr
	//}
}
func makeMessage(cc *Controller, statusDevice int) string {
	switch statusDevice {
	case 1:
		return fmt.Sprintf("%s ПК=%d СК=%d НК=%d", statuses[statusDevice], cc.PK, cc.CK, cc.NK)
	case 2:
		return fmt.Sprintf("%s Фаза=%d", statuses[statusDevice], cc.DK.FDK)
	//case 5:
	//return fmt.Sprintf("%s Фаза=%d", statuses[statusDevice], cc.DK.FDK)
	case 27:
		return fmt.Sprintf("%s ПК=%d СК=%d НК=%d", statuses[statusDevice], cc.PK, cc.CK, cc.NK)
	}
	return fmt.Sprintf("%s%s", cc.getSource(), statuses[statusDevice])
}
