package cameras

import (
	"bufio"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/ruraomsk/ag-server/logger"
	"github.com/ruraomsk/ag-server/pudge"
)

func (c *Connection) workCamera(ext Exchange) {
	if c.Step < 2 {
		c.Step = 2
	}
	tickConnect := time.NewTicker(time.Duration(c.Step/2) * time.Second)
	tickRead := time.NewTicker(time.Duration(c.Step) * time.Second)
	var err error
	var con net.Conn = nil
	var reader *bufio.Reader
	var writer *bufio.Writer
	var data = CameraData{ID: c.ID, Datas: make([]pudge.DataStat, 0), Date: time.Unix(0, 0)}
loop:
	for {
		select {
		case <-ext.to:
			//Отдаем данные
			//logger.Info.Printf("Отправили %v", data)
			ext.from <- data
		case <-tickConnect.C:
			if con == nil {
				con, err = net.Dial("tcp", c.IP)
				if err != nil {
					logger.Error.Print(err.Error())
					con = nil
				}
				reader = bufio.NewReader(con)
				writer = bufio.NewWriter(con)
				_, _ = writer.WriteString(c.Login + "\r\n")
				err = writer.Flush()
				if err != nil {
					logger.Error.Print(err.Error())
					con.Close()
					con = nil
					goto loop
				}
				_, _ = writer.WriteString(c.Password + "\r\n")
				err = writer.Flush()
				if err != nil {
					logger.Error.Print(err.Error())
					con.Close()
					con = nil
					goto loop
				}
				rep, err := reader.ReadString('\n')
				if err != nil {
					logger.Error.Print(err.Error())
					con.Close()
					con = nil
					goto loop
				}
				rep = strings.Replace(rep, "\r", "", 1)
				rep = strings.Replace(rep, "\n", "", 1)
				if !strings.Contains(rep, "Gamotron") {
					logger.Error.Print(rep)
					con.Close()
					con = nil
					goto loop
				}
				//Установим время
				//fmt.Printf("st t=%s\r\n", time.Now().Format("2006-01-02 15:04:05"))
				_, _ = writer.WriteString(fmt.Sprintf("st t=%s\r\n", time.Now().Format("2006-01-02 15:04:05")))
				err = writer.Flush()
				if err != nil {
					logger.Error.Print(err.Error())
					con.Close()
					con = nil
					goto loop
				}
				_, err = reader.ReadString('\n')
				if err != nil {
					logger.Error.Print(err.Error())
					con.Close()
					con = nil
					goto loop
				}
			}
			// else {
			// 	//logger.Info.Printf("Камера %s на связи",c.IP)
			// }
		case <-tickRead.C:
			if con == nil {
				goto loop
			}
			_, _ = writer.WriteString("gd last\r\n")
			err = writer.Flush()
			if err != nil {
				logger.Error.Print(err.Error())
				con.Close()
				con = nil
				goto loop
			}
			data.Datas = make([]pudge.DataStat, 0)
			for {
				rep, err := reader.ReadString('\n')
				if err != nil {
					logger.Error.Print(err.Error())
					con.Close()
					con = nil
					goto loop
				}
				rep = strings.Replace(rep, "\r", "", 1)
				rep = strings.Replace(rep, "\n", "", 1)
				ds := strings.Split(rep, ",")
				if len(ds) < 14 {
					//logger.Error.Printf("Приняли [%s]",rep)
					continue
				}
				d := pudge.DataStat{}
				d.Chanel, _ = strconv.Atoi(ds[0])
				d.Status = 0
				d.Intensiv, _ = strconv.Atoi(ds[3])
				data.Date, _ = time.Parse("2006-01-02 15:04:05", ds[1])
				d.Speed, _ = strconv.Atoi(ds[4])
				d.Density, _ = strconv.Atoi(ds[5])
				d.Occupant, _ = strconv.Atoi(ds[6])
				d.GP, _ = strconv.Atoi(ds[len(ds)-4])
				data.Datas = append(data.Datas, d)
				if d.Chanel == c.Zones {
					break
				}
			}
			//logger.Info.Printf("Перевели ввод с камеры %v", data)
		}
	}
}
