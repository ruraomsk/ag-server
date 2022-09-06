package svgsave

import (
	"crypto/md5"
	"database/sql"
	"fmt"
	"io/ioutil"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/ruraomsk/ag-server/logger"
	"github.com/ruraomsk/ag-server/setup"
)

type record struct {
	region  int
	area    int
	cross   int
	svg     byte     //0 svg 1 map 2 - template.tmpl
	hash    [16]byte //Hash на весь файл
	used    bool     //True если он существует
	receive bool     //Истина если нужно отправить
}

var files map[byte]string
var mapFiles map[string]*record
var socket net.Conn
var err error
var connected bool

func newCross(region, area, cross int, svg byte) *record {
	rec := new(record)
	rec.region = region
	rec.area = area
	rec.cross = cross
	rec.svg = svg
	rec.used = true
	rec.receive = true
	return rec
}

func getKeys(region, area, cross int) []string {
	result := make([]string, 0)
	for d, _ := range files {
		result = append(result, fmt.Sprintf("%d/%d/%d/%d", region, area, cross, d))
	}
	return result
}
func getSVG(key string) byte {
	ss := strings.Split(key, "/")
	n, _ := strconv.Atoi(ss[3])
	return byte(n)
}
func isCross(region, area, cross int) {
	keys := getKeys(region, area, cross)
	var rec *record
	var is bool
	for _, key := range keys {
		//fmt.Printf("key %s\n",key)
		_, is = mapFiles[key]
		if !is {
			// Создаем запись в карте
			rec = newCross(region, area, cross, getSVG(key))
			rec.hash = rec.newHash()
			mapFiles[key] = rec
			//fmt.Printf("insert %v\n", rec)
		} else {
			rec = mapFiles[key]
			rec.used = true
			newHash := rec.newHash()
			if !rec.compareHash(newHash) {
				rec.receive = true
				rec.hash = newHash
				//fmt.Printf("update %v\n", rec)
			}

		}
	}
}
func (r *record) compareHash(hash [16]byte) bool {
	for i, _ := range r.hash {
		if r.hash[i] != hash[i] {
			return false
		}
	}
	return true
}
func (r *record) fullPath() string {
	return fmt.Sprintf("%s/%d/%d/%d", setup.Set.Saver.Path, r.region, r.area, r.cross)
}
func (r *record) fullName() string {
	return fmt.Sprintf("%s/%s", r.fullPath(), files[r.svg])
}
func (r *record) bufferSendFile() []byte {
	//Структура буфера
	// 0	-region
	// 1	-area
	// 2-3 	-cross
	// 4 	-type 0 cross.svg 1-map.png 2 - template.tmpl
	// 5-8  -длина файла
	// 9+длина файла - данные файла
	// Если длина файла == 0 то удалить такой файл
	// Если все поля равны 0 то это Keep Alive
	temp, err := ioutil.ReadFile(r.fullName())
	if err != nil {
		// logger.Error.Printf("Error open file %s %s", r.fullName(), err.Error())
		return make([]byte, 0)
	}
	buf := make([]byte, 9)
	buf[0] = byte(r.region)
	buf[1] = byte(r.area)
	buf[2] = byte((r.cross >> 8) & 0xff)
	buf[3] = byte(r.cross & 0xff)
	buf[4] = r.svg
	buf[5] = byte((len(temp) >> 24) & 0xff)
	buf[6] = byte((len(temp) >> 16) & 0xff)
	buf[7] = byte((len(temp) >> 8) & 0xff)
	buf[8] = byte(len(temp) & 0xff)
	buf = append(buf, temp...)
	return buf
}
func (r *record) bufferDeleteFile() []byte {
	//Структура буфера
	// 0	-region
	// 1	-area
	// 2-3 	-cross
	// 4 	-type 0 cross.svg 1-map.png 2 - template.tmpl
	// 5-8  -длина файла
	// 9+длина файла - данные файла
	// Если длина файла == 0 то удалить такой файл
	// Если все поля равны 0 то это Keep Alive
	buf := make([]byte, 9)
	buf[0] = byte(r.region)
	buf[1] = byte(r.area)
	buf[2] = byte((r.cross >> 8) & 0xff)
	buf[3] = byte(r.cross & 0xff)
	buf[4] = r.svg
	buf[5] = 0
	buf[6] = 0
	buf[7] = 0
	buf[8] = 0
	return buf
}
func (r *record) newHash() [16]byte {
	var result [16]byte
	file, err := ioutil.ReadFile(r.fullName())
	if err != nil {
		// logger.Error.Printf("При чтении файла %s ошибка %s", r.fullName(), err.Error())
		return result
	}
	//fmt.Printf("md5 %v\n",md5.Sum(file))
	return md5.Sum(file)
}
func bufferKeepAlive() []byte {
	//Структура буфера
	// 0	-region
	// 1	-area
	// 2-3 	-cross
	// 4 	-type 0 cross.svg 1-map.png 2 template.tmpl
	// 5-8  -длина файла
	// 9+длина файла - данные файла
	// Если длина файла == 0 то удалить такой файл
	// Если все поля равны 0 то это Keep Alive
	buf := make([]byte, 9)
	for i, _ := range buf {
		buf[i] = 0
	}
	return buf
}
func send(buf []byte) bool {
	if !connected {
		return false
	}
	if len(buf) < 9 {
		return true
	}
	l := int(buf[5]) << 24
	l = l | (int(buf[6]) << 16)
	l = l | (int(buf[7]) << 8)
	l = l | int(buf[8])
	//logger.Info.Printf("%v len=%d", buf[0:9],l)
	//socket.SetWriteDeadline(time.Now().Add(30 * time.Second))
	n, err := socket.Write(buf)
	if err != nil {
		logger.Error.Printf("При передаче на %s ошибка %s", socket.RemoteAddr().String(), err.Error())
		socket.Close()
		connected = false
		return false
	}
	if n != len(buf) {
		logger.Error.Printf("При передаче на %s передано %d вместо %d", socket.RemoteAddr().String(), n, len(buf))
		socket.Close()
		connected = false
		return false
	}
	recv := make([]byte, 1)

	_, err = socket.Read(recv)
	if err != nil {
		logger.Error.Printf("При приеме квитанции на %s ошибка %s", socket.RemoteAddr().String(), err.Error())
		socket.Close()
		connected = false
		return false
	}

	return true
}
func (r *record) sendKillFile() {
	send(r.bufferDeleteFile())
}
func (r *record) sendFile() {
	send(r.bufferSendFile())
}
func sendKeepAlive() {
	send(bufferKeepAlive())
}
func Start() {
	files = make(map[byte]string)
	for i := 0; i < len(setup.Set.Saver.Files); i++ {
		s := setup.Set.Saver.Files[i][0]
		n, _ := strconv.Atoi(setup.Set.Saver.Files[i][1])
		files[byte(n)] = s
	}
	dbinfo := fmt.Sprintf("host=%s user=%s password=%s dbname=%s sslmode=disable",
		setup.Set.DataBase.Host, setup.Set.DataBase.User,
		setup.Set.DataBase.Password, setup.Set.DataBase.DBname)
	dbb, err := sql.Open("postgres", dbinfo)
	if err != nil {
		logger.Error.Printf("Запрос на открытие %s %s", dbinfo, err.Error())
		return
	}
	defer dbb.Close()
	if err = dbb.Ping(); err != nil {
		logger.Error.Printf("Ping %s", err.Error())
		return
	}
	mapFiles = make(map[string]*record)
	connected = false
	for true {
		time.Sleep(time.Duration(setup.Set.Saver.StepSVG) * time.Second)
		if !connected {
			socket, err = net.Dial("tcp", setup.Set.Saver.Svg)
			if err != nil {
				logger.Error.Printf("Error dial %s %s", setup.Set.Saver.Svg, err.Error())
				continue
			}
			connected = true
			mapFiles = make(map[string]*record)
		}
		rows, err := dbb.Query("select region,area,id from public.cross; ")
		if err != nil {
			logger.Error.Printf("Error read public.cross %s", err.Error())
			continue
		}
		for rows.Next() {
			var region, area, cross int
			rows.Scan(&region, &area, &cross)
			//fmt.Printf("cross %d %d %d\n", region, area, cross)
			isCross(region, area, cross)
		}
		rows.Close()
		needKeep := true
		for key, rec := range mapFiles {
			if !connected {
				break
			}
			//time.Sleep(1*time.Second)
			if !rec.used {
				rec.sendKillFile()
				delete(mapFiles, key)
				needKeep = false
				continue
			}
			if rec.receive {
				rec.sendFile()
				needKeep = false
			}
		}
		if needKeep {
			sendKeepAlive()
		}
		for _, rec := range mapFiles {
			rec.used = false
			rec.receive = false
		}
	}
}
