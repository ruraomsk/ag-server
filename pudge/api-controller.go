package pudge

import (
	"fmt"
	"strings"
)

//IsConnected возвращает на связи ли устройство
func (c *Controller) IsConnected() bool {
	return c.StatusConnection
}

//IsRegistred проверяет зарегистрирован ли Id на перекрестке
func IsRegistred(id int) *Region {
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
	mutexCross.Lock()
	defer mutexCross.Unlock()
	// logger.Debug.Print("setStatusCross start")
	result := make(map[string]*Cross, 0)
	for _, cr := range crosses {
		cc, is := GetController(cr.IDevice)

		if !is {
			continue
		}
		//Вычисляем новый статус

		statusDevice := cc.calcStatus()
		if statusDevice != cr.StatusDevice {
			cr.StatusDevice = statusDevice
			cr.WriteToDB = true
			mes := fmt.Sprintf("Режим %s ПК=%d СК=%d НК=%d", statuses[statusDevice], cc.PK, cc.CK, cc.NK)
			cc.LastLogString = mes
			SetController(cc)
			ChanLog <- RecLogCtrl{ID: cc.ID, LogString: mes}
		} else {
			mes := fmt.Sprintf("Режим %s ПК=%d СК=%d НК=%d", statuses[statusDevice], cc.PK, cc.CK, cc.NK)
			if strings.Compare(cc.LastLogString, mes) != 0 {
				cc.LastLogString = mes
				SetController(cc)
				ChanLog <- RecLogCtrl{ID: cc.ID, LogString: mes}
			}
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
		//if !reflect.DeepEqual(&cr.Statistics, &cc.Statistics) {
		//	//logger.Info.Printf("region %d area %d cross %d device %d измениась статистика",cr.Region,cr.Area,cr.ID,cr.IDevice)
		//	cr.Statistics = make([]Statistic, 0)
		//	for _, s := range cc.Statistics {
		//		cr.Statistics = append(cr.Statistics, s)
		//	}
		//	cr.WriteToDB = true
		//}
		if cr.WriteToDB {
			reg := Region{cr.Region, cr.Area, cr.ID}
			result[reg.ToKey()] = cr
		}
	}
	for key, cr := range result {
		crosses[key] = cr
	}
	// logger.Debug.Print("setStatusCross end")
}
