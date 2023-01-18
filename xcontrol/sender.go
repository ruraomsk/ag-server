package xcontrol

import (
	"encoding/json"
	"net"
	"time"

	"github.com/ruraomsk/ag-server/logger"
	"github.com/ruraomsk/ag-server/setup"
)

func sender() {
	for {
		soc, err := net.Dial("tcp", setup.Set.XCtrl.FullHost)
		if err != nil {
			logger.Error.Printf("Sender Соединение с сервером команд %s", err.Error())
			continue
		}
		logger.Info.Printf("Sender started...")
		_, err = soc.Write([]byte("0\n"))
		if err != nil {
			logger.Error.Printf("Передача keep alive на сервер команд %s", err.Error())
			soc.Close()
			continue
		}
		alive := time.NewTicker(1 * time.Minute)
		for {
			select {
			case <-alive.C:
				_, err = soc.Write([]byte("0\n"))
				if err != nil {
					logger.Error.Printf("Передача keep alive на сервер команд %s", err.Error())
					soc.Close()
					break
				}
			case cmd := <-commARM:
				//logger.Debug.Printf("Команда %v",cmd)
				c, err := json.Marshal(cmd)
				if err != nil {
					logger.Error.Printf("Конвертация команды %v %s", cmd, err.Error())
					continue
				}
				c = append(c, '\n')
				_, err = soc.Write(c)
				if err != nil {
					logger.Error.Printf("Передача %s на сервер команд %s", string(c), err.Error())
					soc.Close()
					break
				}
			}
		}

	}
}
