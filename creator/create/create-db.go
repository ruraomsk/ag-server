package create

import (
	"bufio"
	"bytes"
	"database/sql"
	"fmt"
	"io/ioutil"
	"rura/ag-server/logger"
	"rura/ag-server/setup"
	"strconv"
	"strings"
	"unicode/utf16"
	"unicode/utf8"

	_ "github.com/lib/pq"
)

//SQLCreate просмотр каталога и исполнить все запросы с расширением create
func SQLCreate(path string) error {
	dbinfo := fmt.Sprintf("host=%s user=%s password=%s dbname=%s sslmode=disable",
		setup.Set.DataBase.Host, setup.Set.DataBase.User,
		setup.Set.DataBase.Password, setup.Set.DataBase.DBname)
	con, err := sql.Open("postgres", dbinfo)
	if err != nil {
		logger.Error.Printf("Запрос на открытие %s %s", dbinfo, err.Error())
		return err
	}
	defer con.Close()
	if err = con.Ping(); err != nil {
		logger.Error.Printf("Ping %s", err.Error())
		return err
	}
	dirs, err := ioutil.ReadDir(path)
	if err != nil {
		logger.Error.Printf("Ошибка чтения содержимого кaталога %s %s", path, err.Error())
		return err
	}
	//Создаем все таблицы
	for _, dir := range dirs {
		if dir.IsDir() {
			continue
		}
		if !strings.HasSuffix(dir.Name(), ".sql") {
			continue
		}
		nfile := path + "/" + dir.Name()
		cmd, err := ioutil.ReadFile(nfile)
		if err != nil {
			logger.Error.Printf("Error reading file %s! %s\n", path, err.Error())
			return err
		}
		logger.Info.Printf("Обрабатываем файл %s", nfile)
		_, err = con.Exec(string(cmd))

		if err != nil {
			logger.Error.Printf("Error create  %s\n", err.Error())
			return err
		}

	}
	dirs, err = ioutil.ReadDir(path)
	//Загружаем координаты устройств
	for _, dir := range dirs {
		if dir.IsDir() {
			continue
		}
		if !strings.HasSuffix(dir.Name(), ".mrk") {
			// fmt.Println(dir.Name())
			continue
		}
		nfile := path + "/" + dir.Name()
		file, err := ioutil.ReadFile(nfile)
		if err != nil {
			logger.Error.Printf("Error reading file %s! %s\n", path, err.Error())
			return err
		}
		logger.Info.Printf("Обрабатываем файл %s", nfile)
		region := "0"
		if strings.Contains(nfile, "Мосавтодор") {
			region = "1"
		}
		strFile, err := decodeUTF16(file)
		strFile = strings.ReplaceAll(strFile, "\ufeff", "")
		scanner := bufio.NewScanner(strings.NewReader(strFile))
		for scanner.Scan() {
			str := scanner.Text()
			if len(str) == 0 {
				continue
			}

			ss := strings.Split(str, "#")
			if len(ss) != 3 {
				continue
			}
			w := "insert into dev_gis (region,id,dgis,describ) values(" + region + "," + ss[0] + ",point(" + ss[1] + "),'" + ss[2] + "');"
			_, err = con.Exec(w)

			if err != nil {
				logger.Error.Printf("Error %s  %s\n", w, err.Error())
				return err
			}

		}

	}
	//Теперь создаем таблицу привязки cross
	rows, err := con.Query("select region,id from dev_gis;")
	var region int
	var id int
	for rows.Next() {
		err = rows.Scan(&region, &id)
		if err != nil {
			logger.Error.Printf("Error   %s\n", err.Error())
			return err
		}
		w := "insert into public.\"cross\" (region,id,idevice) values(" + strconv.Itoa(region) + "," + strconv.Itoa(id) + "," + strconv.Itoa(region*10000+id) + ");"
		_, err = con.Exec(w)
		if err != nil {
			logger.Error.Printf("Error   %s\n", err.Error())
			return err
		}
	}

	return nil
}
func decodeUTF16(b []byte) (string, error) {

	if len(b)%2 != 0 {
		return "", fmt.Errorf("Must have even length byte slice")
	}

	u16s := make([]uint16, 1)

	ret := &bytes.Buffer{}

	b8buf := make([]byte, 4)

	lb := len(b)
	for i := 0; i < lb; i += 2 {
		u16s[0] = uint16(b[i]) + (uint16(b[i+1]) << 8)
		r := utf16.Decode(u16s)
		n := utf8.EncodeRune(b8buf, r[0])
		ret.Write(b8buf[:n])
	}

	return ret.String(), nil
}
