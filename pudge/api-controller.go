package pudge

import (
	"fmt"
	"reflect"
	"strings"
)

//IsConnected возвращает на связи ли устройство
func (c *Controller) IsConnected() bool {
	return c.StatusConnection == Connected
}

//IsRegistred проверяет зарегистрирован ли Id на перекрестке
func IsRegistred(id int) string {
	// mutex.Lock()
	// defer mutex.Unlock()
	for _, c := range crosses {
		if c.IDevice == id {
			reg := Region{Region: c.Region, Area: c.Area, ID: c.ID}
			return reg.ToKey()
		}
	}
	return ""
}

var cCount int

func setStatusCross() {
	// mutex.Lock()
	// defer mutex.Unlock()
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
			mes := fmt.Sprintf("Режим %s ПК=%d СК=%d НК=%d", statuses[statusDevice], cc.PK, cc.CK, cc.NK)
			cc.LastLogString = mes
			ChanLog <- RecLogCtrl{ID: cc.ID, LogString: mes}
			SetController(cc)
		} else {
			mes := fmt.Sprintf("Режим %s ПК=%d СК=%d НК=%d", statuses[statusDevice], cc.PK, cc.CK, cc.NK)
			if strings.Compare(cc.LastLogString, mes) != 0 {
				cc.LastLogString = mes
				ChanLog <- RecLogCtrl{ID: cc.ID, LogString: mes}
				SetController(cc)
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
		if !reflect.DeepEqual(&cr.Statistics, &cc.Statistics) {
			cr.Statistics = make([]Statistic, 0)
			for _, s := range cc.Statistics {
				cr.Statistics = append(cr.Statistics, s)
			}
			cr.WriteToDB = true
		}
		if cr.WriteToDB {
			reg := Region{cr.Region, cr.Area, cr.ID}
			crosses[reg.ToKey()] = cr
		}
	}
}
