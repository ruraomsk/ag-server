package cameras

import (
	"github.com/ruraomsk/ag-server/pudge"
	"time"
)

type SetupCameras struct {
	DefCameras []DefCamera `json:"cameras"`
}
type DefCamera struct {
	Region      int          `json:"region"`
	Area        int          `json:"area"`
	ID          int          `json:"id"`
	Step        int          `json:"step"`
	Connections []Connection `json:"connections"`
	Relations   []Relation   `json:"relations"`
}
type Connection struct {
	ID       int    `json:"id"`
	IP       string `json:"ip"`
	Login    string `json:"login"`
	Password string `json:"password"`
	Zones    int    `json:"zones"` // Число зон которые снимает камера
	Step     int    `json:"step"`  //Интервал в секундах когда опрашивать
}
type Relation struct {
	Channel int `json:"channel"`
	ID      int `json:"id"`
	Zone    int `json:"zone"`
}
type ExtStatistic struct {
	Region    int       `json:"region"`
	Area      int       `json:"area"`
	ID        int       `json:"id"`
	Date      time.Time `json:"date"`
	Statistic pudge.Statistic
}
type CameraData struct {
	ID    int `json:"id"`
	Date  time.Time
	Datas []pudge.DataStat
}
type Exchange struct {
	to   chan interface{}
	from chan CameraData
}
