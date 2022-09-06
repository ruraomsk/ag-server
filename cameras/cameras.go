package cameras

import (
	"fmt"
	"github.com/jasonlvhit/gocron"
	"github.com/ruraomsk/ag-server/pudge"
	"sync"
	"time"
)

var setcams SetupCameras
var writer chan ExtStatistic
var cams map[int]Exchange
var mutex sync.Mutex

func (dc *DefCamera) loadData() {
	result := ExtStatistic{Region: dc.Region, Area: dc.Area, ID: dc.ID, Date: time.Now()}
	result.Statistic = pudge.Statistic{Hour: result.Date.Hour(), Min: result.Date.Minute(), TLen: dc.Step, Type: 3, Datas: make([]pudge.DataStat, 0)}
	max := 0
	for _, c := range dc.Connections {
		if c.Step > max {
			max = c.Step
		}
	}
	time.Sleep(time.Duration(max) * time.Second)
	for _, c := range dc.Connections {
		ex := cams[c.ID]
		ex.to <- 1
		d := <-ex.from
		td := d.Date.Hour()*60 + d.Date.Minute()
		tt := result.Statistic.Hour*60 + result.Statistic.Min

		if td > tt-dc.Step && td < tt+dc.Step {
			for _, r := range dc.Relations {
				if r.ID == c.ID {
					for _, z := range d.Datas {
						if z.Chanel == r.Zone {
							z.Chanel = r.Channel
							result.Statistic.Datas = append(result.Statistic.Datas, z)
						}
					}
				}
			}
		}
	}
	//logger.Info.Printf("Шлем на запись %v",result)
	writer <- result
}
func (dc *DefCamera) startwork() {
	//logger.Info.Printf("Настраиваемся на %d:%d:%d", dc.Region, dc.Area, dc.ID)
	for time.Now().Minute()%dc.Step != 0 {
		time.Sleep(1 * time.Second)
	}
	//logger.Info.Printf("Запускаем камеры на %d:%d:%d", dc.Region, dc.Area, dc.ID)
	for _, con := range dc.Connections {
		ex := Exchange{to: make(chan interface{}), from: make(chan CameraData)}
		mutex.Lock()
		cams[con.ID] = ex
		mutex.Unlock()
		go con.workCamera(ex)
	}
	_ = gocron.Every(uint64(dc.Step)).Minutes().Do(dc.loadData)
	//logger.Info.Printf("Камеры на %d:%d:%d запущены", dc.Region, dc.Area, dc.ID)

	<-gocron.Start()
}
func CamerasStart(path string) {
	fmt.Println("Модуль приема статистики от видеокамер")
	setcams = SetupCameras{DefCameras: make([]DefCamera, 0)}
	df := DefCamera{ID: 300, Step: 5, Region: 1, Area: 1, Relations: make([]Relation, 0), Connections: make([]Connection, 0)}
	df.Connections = append(df.Connections, Connection{ID: 1, Step: 10, Zones: 4, Password: "admin", Login: "admin", IP: "192.168.115.168:8441"})
	df.Relations = append(df.Relations, Relation{ID: 1, Channel: 1, Zone: 1})
	df.Relations = append(df.Relations, Relation{ID: 1, Channel: 2, Zone: 2})
	df.Relations = append(df.Relations, Relation{ID: 1, Channel: 4, Zone: 3})
	df.Relations = append(df.Relations, Relation{ID: 1, Channel: 5, Zone: 4})
	setcams.DefCameras = append(setcams.DefCameras, df)
	df.ID = 500
	setcams.DefCameras = append(setcams.DefCameras, df)

	//buf, err := ioutil.ReadFile(path)
	//if err != nil {
	//	logger.Error.Println(err.Error())
	//	return
	//}
	//err = xml.Unmarshal(buf, &setcams)
	//if err != nil {
	//	logger.Error.Println(err.Error())
	//	return
	//}
	writer = make(chan ExtStatistic)
	cams = make(map[int]Exchange)
	go writedata()
	for _, cam := range setcams.DefCameras {
		go cam.startwork()
	}
}
