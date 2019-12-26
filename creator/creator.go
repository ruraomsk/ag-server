package creator

import (
	"bufio"
	"bytes"
	"database/sql"
	"encoding/xml"
	"io/ioutil"
	"rura/ag-server/setup"
	"strconv"
	"strings"

	"fmt"
	"rura/ag-server/logger"

	_ "github.com/lib/pq"
)

//Создает все таблицы в постгресс
//Дополнительно производит загрузку из внешних текстовых файлов
//Если таковые будут))
var err error

//Creator define Creator
type Creator struct {
	SQL     SQL     `xml:"sql" json:"sql"`
	Regions Regions `xml:"regions" json:"regions"`
}

//SQL def SQL
type SQL struct {
	Ext  string `xml:"ext,attr" json:"ext"`
	Path string `xml:"path,attr" json:"path"`
}

//Regions def Regions
type Regions struct {
	Path string   `xml:"path,attr" json:"path"`
	Regs []Region `xml:"reg" json:"reg"`
}

//Region def Region
type Region struct {
	ID    int    `xml:"id,attr" json:"id"`
	Name  string `xml:"name,attr" json:"name"`
	Areas []Area `xml:"area" json:"area"`
}

//Area def area
type Area struct {
	ID       int       `xml:"id,attr" json:"id"`
	Name     string    `xml:"name,attr" json:"name"`
	SubAreas []SubArea `xml:"subarea" json:"subarea"`
}

//SubArea def subarea
type SubArea struct {
	ID   int    `xml:"id,attr" json:"id"`
	File string `xml:"file,attr" json:"file"`
}

var st Creator
var con *sql.DB

//Start создание баз данных
func Start(path string) error {
	logger.Info.Println("Start creator...")
	fmt.Println("Start creator...")
	buf, err := ioutil.ReadFile(path + "/setup/creator.xml")
	if err != nil {
		logger.Error.Println(err.Error())
		return err
	}
	err = xml.Unmarshal(buf, &st)
	if err != nil {
		logger.Error.Println(err.Error())
		return err
	}
	dbinfo := fmt.Sprintf("host=%s user=%s password=%s dbname=%s sslmode=disable",
		setup.Set.DataBase.Host, setup.Set.DataBase.User,
		setup.Set.DataBase.Password, setup.Set.DataBase.DBname)
	con, err = sql.Open("postgres", dbinfo)
	if err != nil {
		logger.Error.Printf("Запрос на открытие %s %s", dbinfo, err.Error())
		return err
	}
	defer con.Close()
	if err = con.Ping(); err != nil {
		logger.Error.Printf("Ping %s", err.Error())
		return err
	}
	err = sqlCreate(path+"/"+st.SQL.Path, st.SQL.Ext)
	if err != nil {
		return fmt.Errorf("Найдены ошибки проверьте log file")
	}
	err = regionCreate(path)
	logger.Info.Println("Exit creator...")
	fmt.Println("Exit creator...")
	return nil
}
func regionCreate(path string) error {
	path += "/" + st.Regions.Path
	for _, reg := range st.Regions.Regs {
		for _, ar := range reg.Areas {
			w := "insert into region (region,area,nameregion,namearea) values(" + strconv.Itoa(reg.ID) +
				"," + strconv.Itoa(ar.ID) + ",'" + reg.Name + "','" + ar.Name + "');"
			_, err = con.Exec(w)
			if err != nil {
				logger.Error.Printf("Error %s  %s\n", w, err.Error())
				return err
			}
			for _, sub := range ar.SubAreas {
				err = loadCross(reg.ID, ar.ID, sub.ID, path+"/"+sub.File)
				if err != nil {
					logger.Error.Printf("Error loadCross  %s\n", err.Error())
					return err
				}

			}
		}
	}
	return nil
}

//sqlCreate просмотр каталога и исполнить все запросы с расширением create
func sqlCreate(path string, ext string) error {
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
		if !strings.HasSuffix(dir.Name(), ext) {
			continue
		}
		nfile := path + "/" + dir.Name()
		cmd, err := ioutil.ReadFile(nfile)
		if err != nil {
			logger.Error.Printf("Error reading file %s! %s\n", path, err.Error())
			return err
		}
		logger.Info.Printf("Обрабатываем файл %s", nfile)
		// fmt.Println(string(cmd))
		_, err = con.Exec(string(cmd))

		if err != nil {
			logger.Error.Printf("Error create  %s\n", err.Error())
			return err
		}

	}
	return nil
}
func loadCross(region int, area int, subarea int, nfile string) error {
	file, err := ioutil.ReadFile(nfile)
	if err != nil {
		logger.Error.Printf("Error reading file %s! %s\n", nfile, err.Error())
		return err
	}
	logger.Info.Printf("Обрабатываем файл %s", nfile)
	var scanner *bufio.Scanner
	scanner = bufio.NewScanner(bytes.NewReader(file))
	reg := strconv.Itoa(region) + "," + strconv.Itoa(area) + "," + strconv.Itoa(subarea)
	for scanner.Scan() {
		str := scanner.Text()
		if len(str) == 0 {
			continue
		}

		ss := strings.Split(str, "#")
		if len(ss) != 3 {
			continue
		}
		id, _ := strconv.Atoi(ss[0])
		w := "insert into public.\"cross\" (region,area,subarea,id,dgis,describ,idevice,state) values(" + reg + "," + ss[0] + ",point(" + ss[1] + "),'" + ss[2] + "'," +
			strconv.Itoa(region*10000+id) + ",'{}');"
		_, err = con.Exec(w)

		if err != nil {
			logger.Error.Printf("Error %s  %s\n", w, err.Error())
			return err
		}

	}
	return nil
}
