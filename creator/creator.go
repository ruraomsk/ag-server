package creator

import (
	"bufio"
	"bytes"
	"database/sql"
	"encoding/json"
	"encoding/xml"
	"io/ioutil"
	"rura/ag-server/pudge"
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
	File  string `xml:"file,attr" json:"file"`
	Areas []Area `xml:"area" json:"area"`
}

//Area def area
type Area struct {
	ID   int    `xml:"id,attr" json:"id"`
	Name string `xml:"name,attr" json:"name"`
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

		}
		err = loadCross(reg.ID, path+"/"+reg.File)
		if err != nil {
			logger.Error.Printf("Error loadCross  %s\n", err.Error())
			return err
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
func loadCross(region int, nfile string) error {
	file, err := ioutil.ReadFile(nfile)
	if err != nil {
		logger.Error.Printf("Error reading file %s! %s\n", nfile, err.Error())
		return err
	}
	logger.Info.Printf("Обрабатываем файл %s", nfile)
	var scanner *bufio.Scanner
	scanner = bufio.NewScanner(bytes.NewReader(file))
	isempty := true
	state := pudge.NewCross()
	var dgis string
	for scanner.Scan() {
		str := scanner.Text()
		if len(str) == 0 {
			continue
		}
		if strings.HasPrefix(str, "@u,") {
			//Начало нового перекрестка
			if !isempty {
				saveState(state, dgis)
			}
			isempty = false
			state = pudge.NewCross()
			ss := strings.Split(str, ",")
			if len(ss) != 8 {
				logger.Error.Printf("в строке %s неверное число параметров", str)
				return err
			}
			state.ID, _ = strconv.Atoi(ss[1])
			state.Region = region
			state.Area, _ = strconv.Atoi(ss[4])
			state.SubArea, _ = strconv.Atoi(ss[5])
			state.NumDev, _ = strconv.Atoi(ss[2])
			state.ConType = ss[3][0:2]
			state.ID, _ = strconv.Atoi(ss[3][2:])
			continue
		}
		if strings.HasPrefix(str, "@C,") {
			//Координаты
			dgis = str[3:]
			continue
		}
		if strings.HasPrefix(str, "@S,") {
			//Наименование
			state.Name = str[3:]
			continue
		}
		if strings.HasPrefix(str, "@N,") {
			//Телефон
			state.Fone = str[3:]
			continue
		}
		if strings.HasPrefix(str, "@k1,") {
			//Массив
			ss := strings.Split(str, ",")
			sint := make([]int, 0)
			for i := 1; i < len(ss); i++ {
				ii, _ := strconv.Atoi(ss[i])
				sint = append(sint, ii)
			}
			if sint[0] == 14 {
				err = state.StatDefine.FromBuffer(sint)
				if err != nil {
					logger.Error.Printf("в строке %s %s", str, err.Error())
					return err
				}
				continue
			}
			if sint[0] == 15 {
				err = state.PointSet.FromBuffer(sint)
				if err != nil {
					logger.Error.Printf("в строке %s %s", str, err.Error())
					return err
				}
				continue
			}
			if sint[0] == 16 {
				err = state.UseInput.FromBuffer(sint)
				if err != nil {
					logger.Error.Printf("в строке %s %s", str, err.Error())
					return err
				}
				continue
			}
			if sint[0] == 21 {
				err = state.TimeDivice.FromBuffer(sint)
				if err != nil {
					logger.Error.Printf("в строке %s %s", str, err.Error())
					return err
				}
				continue
			}
			if sint[0] == 40 {
				err = state.Arrays.SetupDK1.FromBuffer(sint)
				if err != nil {
					logger.Error.Printf("в строке %s %s", str, err.Error())
					return err
				}
				continue
			}
			if sint[0] == 41 {
				err = state.Arrays.SetupDK2.FromBuffer(sint)
				if err != nil {
					logger.Error.Printf("в строке %s %s", str, err.Error())
					return err
				}
				continue
			}
			if sint[0] >= 45 && sint[0] <= 56 {
				//Недельные массивы
				err = state.Arrays.NedelSets.FromBuffer(sint)
				if err != nil {
					logger.Error.Printf("в строке %s %s", str, err.Error())
					return err
				}
				continue
			}
			if sint[0] >= 65 && sint[0] <= 76 {
				//Суточные массивы
				err = state.Arrays.DaySets.FromBuffer(sint)
				if err != nil {
					logger.Error.Printf("в строке %s %s", str, err.Error())
					return err
				}
				continue
			}
			if sint[0] >= 85 && sint[0] <= 96 {
				//Годовые массивы
				err = state.Arrays.MonthSets.FromBuffer(sint)
				if err != nil {
					logger.Error.Printf("в строке %s %s", str, err.Error())
					return err
				}
				continue
			}
			if sint[0] >= 100 && sint[0] <= 131 {
				//Планы координации
				err = state.Arrays.SetDK.FromBuffer(sint)
				if err != nil {
					logger.Error.Printf("в строке %s %s", str, err.Error())
					return err
				}
				continue
			}
			logger.Info.Printf("в строке %s нет такого обработчика", str)

		}

	}
	return nil
}
func saveState(state *pudge.Cross, dgis string) error {
	b, err := json.Marshal(&state)
	if err != nil {
		logger.Error.Printf("%s\n", err.Error())
		return err
	}

	w := fmt.Sprintf("insert into public.\"cross\" (region,area,subarea,id,dgis,describ,idevice,state) values(%d,%d,%d,%d,point(%s),'%s',%d,'%s');",
		state.Region, state.Area, state.SubArea, state.ID, dgis, state.Name, state.IDevice, string(b))
	_, err = con.Exec(w)

	if err != nil {
		logger.Error.Printf("Error %s  %s\n", w, err.Error())
		return err
	}
	return nil
}
