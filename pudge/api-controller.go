package pudge

import (
	"strconv"
	"time"
)

// IsConnected возвращает на связи ли устройство
func (cc *Controller) IsConnected() bool {
	return cc.StatusConnection
}

// var cCount int
func setStatusCross() {
	mutexCtrl.Lock()
	defer mutexCtrl.Unlock()
	// logger.Debug.Print("setStatusCross start")
	//result := make(map[string]*Cross, 0)
	for _, cr := range crosses {
		cc, is := controllers[cr.IDevice]

		if !is {
			if cr.StatusDevice != 18 {
				cr.StatusDevice = 18
				cr.PK = 0
				cr.CK = 0
				cr.NK = 0
				cr.WriteToDB = true
			}
			continue
		}
		//Вычисляем новый статус

		statusDevice := cc.CalcStatus()
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
		//		ChanLog <- LogRecord{ID: cc.ID, Type: 0, Time: time.Now(), Journal: SetTechStatus(cc.ID)}

		if controls[statusDevice] {
			w := "Лампы "
			if cc.DK.LDK == 0 {
				w += "исправны "
			} else {
				p := "фаза " + strconv.Itoa(cc.DK.LDK)
				switch cc.DK.LDK {
				case 9:
					p = "ПРОМТАКТ"
				case 10:
					p = "ЖМ"
				case 11:
					p = "ОС"
				case 12:
					p = "КК"
				}
				w += "не исправны " + p + " "
			}
			w += " Двери "
			if !cc.DK.ODK {
				w += "закрыты "
			} else {
				w += "открыты "
			}
			ChanLog <- LogRecord{ID: cc.ID, Region: Region{Region: cr.Region, Area: cr.Area, ID: cr.ID}, Type: 2, Time: time.Now(), LogString: w}
		}
	}
}
