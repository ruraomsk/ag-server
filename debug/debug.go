package debug

import (
	"encoding/hex"
	"fmt"
	"os"
	"time"

	"github.com/ruraomsk/ag-server/extcon"
	"github.com/ruraomsk/ag-server/logger"
	"github.com/ruraomsk/ag-server/setup"
)

type DebugMessage struct {
	ID int
	//true от сервера к устройству false от устройства к серверу
	FromTo bool
	Info   bool
	Time   time.Time
	Buffer []byte
}

var (
	DebugChan chan DebugMessage
	working   bool
	files     map[int]*os.File
	err       error
)

func Run() {

	DebugChan = make(chan DebugMessage, 1000)
	files = make(map[int]*os.File)
	oneSec := time.NewTicker(time.Second)
	cont, _ := extcon.NewContext("debug")

	for {
		select {
		case <-oneSec.C:
			working = false
			if len(setup.Set.Debug) != 0 {
				if stat, err := os.Stat(setup.Set.Debug); err == nil && stat.IsDir() {
					working = true
				}
			}
		case message := <-DebugChan:
			if !working {
				continue
			}
			f, is := files[message.ID]
			if !is {
				f, err = os.Create(fmt.Sprintf("%s/device%d", setup.Set.Debug, message.ID))
				if err != nil {
					logger.Error.Printf("При создании лога для устройства %d %s", message.ID, err.Error())
					continue
				}
				files[message.ID] = f
			}
			if !message.Info {
				_, err = f.WriteString(messageRaw(message) + DecodeMessage(message))
				if err != nil {
					logger.Error.Printf("При записи лога для устройства %d %s", message.ID, err.Error())
					continue
				}
			} else {
				_, err = f.WriteString(errorMessageRaw(message))
				if err != nil {
					logger.Error.Printf("При записи лога для устройства %d %s", message.ID, err.Error())
					continue
				}
			}
		case <-cont.Done():
			for _, v := range files {
				v.Close()
			}
		}
	}
}
func errorMessageRaw(message DebugMessage) string {
	flag := "<-"
	if message.FromTo {
		flag = "->"
	}
	return fmt.Sprintf("%s %s %s %s\n", message.Time.Format(time.RFC3339)[0:10], message.Time.Format(time.RFC3339)[11:19], flag, string(message.Buffer))
}
func messageRaw(message DebugMessage) string {
	flag := "<-"
	if message.FromTo {
		flag = "->"
	}
	b := hex.EncodeToString(message.Buffer)
	r := ""
	for i := 0; i < len(b); i += 2 {
		r += b[i : i+2]
		r += " "
	}
	return fmt.Sprintf("%s %s %s %s {%d}", message.Time.Format(time.RFC3339)[0:10], message.Time.Format(time.RFC3339)[11:19], flag, r, len(message.Buffer))
}
func DecodeMessage(message DebugMessage) string {
	r := "["
	if message.FromTo {
		mess := make([]uint8, message.Buffer[12]-2)
		for i := 0; i < len(mess); i++ {
			mess[i] = message.Buffer[13+i]
		}
		pos := 0
		for pos < len(mess) {
			pos++
			l := int(mess[pos])
			pos++
			r += fmt.Sprintf("0x%02X ", mess[pos])
			pos += l
		}
		r += "]\n"
		return r
	}
	mess := make([]uint8, message.Buffer[18]-2)
	for i := 0; i < len(mess); i++ {
		mess[i] = message.Buffer[19+i]
	}
	pos := 0
	for pos < len(mess) {
		switch mess[pos] {
		case 0x00:
			pos++
			r += "0x00 "
		case 0x01:
			pos += 6
			r += "0x01 "
		case 0x04:
			pos += 5
			r += "0x04 "
		case 0x07:
			pos += 5
			r += "0x07 "
		case 0x08:
			pos += 14
			r += "0x08 "
		case 0x09:
			r += "0x09 "
			pos += 6
		case 0x0a:
			l := int(mess[pos+2]) + 3
			if l == 0 {
				l = 3
			}
			r += "0x0A "
			pos += l
		case 0x0b:
			r += "0x0B "
			pos += 16
		case 0x0f:
			r += "0x0F "
			pos += 22
		case 0x10:
			r += "0x10 "
			pos += 6
		case 0x11:
			r += "0x11 "
			pos += 4
		case 0x12:
			r += "0x12 "
			pos += 23
		case 0x13:
			r += "0x13 "
			pos += int(mess[pos+2]) + 3
		case 0x1b:
			r += "0x1B "
			pos += 9
		case 0x1d:
			r += "0x1D "
			pos += 13
		case 0x1c:
			r += "0x1C "
			pos += 6
		default:
			r += "0xXX "
			pos++
		}
	}
	r += "]\n"
	return r
}
