package pudge

import (
	"math/rand"
	"reflect"
)

//IsConnected возвращает на связи ли устройство
func (c *Controller) IsConnected() bool {
	return c.StatusConnection == Connected
}

func isRegistred(id int) string {
	mutex.Lock()
	defer mutex.Unlock()
	for _, c := range crosses {
		if c.IDevice == id {
			return c.Name
		}
	}
	return ""
}

var cCount int

func setStatusCross() {
	mutex.Lock()
	defer mutex.Unlock()
	for _, cr := range crosses {
		cc, is := controllers[cr.IDevice]
		if !is {
			continue
		}
		//Вычисляем новый статус
		// 1 2 3 4 5 6 7 8 9 10
		// 11 12 13 14 15 16 17 18 19 20
		// 21 22 23 24 25 26 27 28 29 30
		// 31 32 33
		// statusDevice := 0

		// for statusDevice == 0 {
		// if cc.DK1.ODK || cc.DK2.ODK {
		// 	//16 - Открыта дверь
		// 	statusDevice = 16
		// 	continue
		// }
		// if cc.Error.V220DK1 || cc.Error.V220DK2 {
		// 	//18 - Авария 220
		// 	statusDevice = 18
		// 	continue
		// }
		// if cc.LastOperation == time.Unix(0, 0) {
		// 	//17 - Нет информации
		// 	statusDevice = 17
		// 	continue
		// }
		// if !cc.GPS.Ok {
		// 	//32 - Неисправность часов
		// 	statusDevice = 32
		// 	continue
		// }
		// if cc.Base {
		// 	//30 - Базовая привязка
		// 	statusDevice = 30
		// 	continue
		// }
		// if !cc.IsConnected() {
		// 	//19 Нет связи с УСДК
		// 	statusDevice = 19
		// 	continue
		// }
		cCount++
		if cCount%3 == 0 {
			statusDevice := rand.Intn(34)
			if statusDevice == 0 || statusDevice == 34 {
				statusDevice = 1
			}
			if statusDevice != cr.StatusDevice {
				cr.StatusDevice = statusDevice
				cr.WriteToDB = true
			}
		}
		// }
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
		if !reflect.DeepEqual(&cr.Statistics, cc.Statistics) {
			cr.Statistics = make([]Statistic, 0)
			for _, s := range cc.Statistics {
				cr.Statistics = append(cr.Statistics, s)
			}
			cr.WriteToDB = true
		}
		if cr.WriteToDB {
			reg := Region{cr.Region, cr.ID}
			crosses[reg.ToKey()] = cr
		}
	}
}
