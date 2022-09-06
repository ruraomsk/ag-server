package loader

import (
	"fmt"
	"net"
	"os"
	"strconv"

	_ "github.com/lib/pq"
	"github.com/ruraomsk/ag-server/logger"
	"github.com/ruraomsk/ag-server/setup"
)

var files map[byte]string

func createPath(path string) error {
	err := os.Chdir(path)
	if err != nil {
		//logger.Info.Printf("Каталог %s не существует. Создаем...", path)
		err = os.MkdirAll(path, os.ModePerm)
		if err != nil {
			logger.Error.Printf("Ошибка создания каталога %s %s", path, err.Error())
			return err
		}
	}
	return nil
}
func StartSVG(stop chan int) {
	//Проверим и если нет то создадим каталог для сохранения рисунков
	err := createPath(setup.Set.Loader.Path)
	if err != nil {
		stop <- 1
		return
	}
	files = make(map[byte]string)
	for i := 0; i < len(setup.Set.Loader.Files); i++ {
		s := setup.Set.Loader.Files[i][0]
		n, _ := strconv.Atoi(setup.Set.Loader.Files[i][1])
		files[byte(n)] = s
	}
	//fmt.Printf("files %v\n",files)
	ln, err := net.Listen("tcp", ":"+strconv.Itoa(setup.Set.Loader.SVGPort))
	if err != nil {
		logger.Error.Printf("Ошибка открытия порта %s", err.Error())
		stop <- 1
		return
	}
	defer ln.Close()
	for {
		socket, err := ln.Accept()
		if err != nil {
			logger.Error.Printf("Ошибка accept %s", err.Error())
			continue
		}
		go workerSVG(socket)
	}
}
func isKeepAlive(buf []byte) bool {
	for _, b := range buf {
		if b != 0 {
			return false
		}
	}
	return true
}
func isKillFile(buf []byte) bool {
	if buf[5] == 0 && buf[6] == 0 && buf[7] == 0 && buf[8] == 0 {
		return true
	}
	return false
}
func fullPath(buf []byte) string {
	region := int(buf[0])
	area := int(buf[1])
	cross := (int(buf[2]) << 8) | int(buf[3])
	return fmt.Sprintf("%s/%d/%d/%d", setup.Set.Loader.Path, region, area, cross)
}
func fullName(buf []byte) string {
	return fmt.Sprintf("%s/%s", fullPath(buf), files[buf[4]])
}
func makeFile(buf, data []byte) {
	err := createPath(fullPath(buf))
	if err != nil {
		return
	}
	file, err := os.Create(fullName(buf))
	if err != nil {
		logger.Error.Printf("При создании файла %s ошибка %s", fullName(buf), err.Error())
		return
	}
	defer file.Close()
	n, err := file.Write(data)
	if err != nil {
		logger.Error.Printf("при записи файда  %s ошибка %s", fullName(buf), err.Error())
		return
	}
	if n != len(data) {
		logger.Error.Printf("при записи файда  %s записано %d байт нужно %d", fullName(buf), n, len(data))
		return
	}
}
func killFile(buf []byte) {
	_ = os.Remove(fullName(buf))
}
func workerSVG(soc net.Conn) {
	defer soc.Close()
	logger.Info.Printf("Новый клиент сервера SVG %s", soc.RemoteAddr().String())
	for {
		buf := make([]byte, 9)
		n, err := soc.Read(buf)
		if err != nil {
			logger.Error.Printf("при чтении заголовка от устройства %s ошибка %s", soc.RemoteAddr().String(), err.Error())
			return
		}
		if n != len(buf) {
			logger.Error.Printf("при чтении заголовка от устройства %s прочитано %d байт нужно %d", soc.RemoteAddr().String(), n, len(buf))
			return
		}
		if isKeepAlive(buf) {
			continue
		}
		if isKillFile(buf) {
			killFile(buf)
			continue
		}

		l := int(buf[5]) << 24
		l = l | (int(buf[6]) << 16)
		l = l | (int(buf[7]) << 8)
		l = l | int(buf[8])
		file := make([]byte, 0)
		//logger.Info.Printf("%v len=%d",buf,l)
		for l > 0 {
			data := make([]byte, l)
			n, err = soc.Read(data)
			if err != nil {
				logger.Error.Printf("при чтении данных от устройства %s ошибка %s", soc.RemoteAddr().String(), err.Error())
				return
			}
			if n > len(data) {
				logger.Error.Printf("при чтении данных от устройства %s прочитано %d байт нужно %d", soc.RemoteAddr().String(), n, len(data))
				return
			}
			for i := 0; i < n; i++ {
				file = append(file, data[i])
			}
			l -= n
		}
		makeFile(buf, file)
		recv := make([]byte, 1)
		n, err = soc.Write(recv)
		if err != nil {
			logger.Error.Printf("при передаче квитанции на устройство %s ошибка %s", soc.RemoteAddr().String(), err.Error())
			return
		}

	}
}
