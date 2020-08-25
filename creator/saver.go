package creator

import (
	"database/sql"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	"github.com/JanFant/TLServer/logger"
	"github.com/ruraomsk/ag-server/pudge"
	"github.com/ruraomsk/ag-server/setup"
)

var st Creator
var con *sql.DB
var iregion int

func createPath(path string) error {
	err := os.Chdir(path)
	if err != nil {
		logger.Info.Printf("Каталог %s не существует. Создаем...", path)
		err = os.MkdirAll(path, os.ModePerm)
		if err != nil {
			logger.Error.Printf("Ошибка создания каталога %s %s", path, err.Error())
			return err
		}
	}
	return nil
}
func sqlCopy(pathSrc string, pathDest string, ext string) error {
	dirs, err := ioutil.ReadDir(pathSrc)
	if err != nil {
		logger.Error.Printf("Ошибка чтения содержимого кaталога %s %s", pathSrc, err.Error())
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
		cmd, err := ioutil.ReadFile(pathSrc + "/" + dir.Name())
		if err != nil {
			logger.Error.Printf("Error reading file %s! %s\n", pathSrc+"/"+dir.Name(), err.Error())
			return err
		}
		logger.Info.Printf("Обрабатываем файл %s", pathSrc+"/"+dir.Name())
		err = ioutil.WriteFile(pathDest+"/"+dir.Name(), cmd, 0777)
		if err != nil {
			logger.Error.Printf("Error writing file %s! %s\n", pathDest+"/"+dir.Name(), err.Error())
			return err
		}
	}
	return nil
}

//SaveAll сохраняет всю БД для правки в символьном виде
func SaveAll(path string, sreg string) error {
	if strings.Contains(sreg, "all") {
		iregion = 0
	} else {
		iregion, _ = strconv.Atoi(sreg)
	}
	logger.Info.Println("Start Save DB...")
	fmt.Println("Start Save DB...")
	buf, err := ioutil.ReadFile(setup.Set.Home + "/setup/creator.xml")
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
	err = createPath(path)
	if err != nil {
		return err
	}
	err = sqlCopy(setup.Set.Home+"/"+st.SQL.Path, path, st.SQL.Ext)
	if err != nil {
		return err
	}
	for _, reg := range st.Regions.Regs {
		if iregion > 0 && iregion != reg.ID {
			continue
		}
		file, err := os.Create(path + "/" + reg.File)
		defer file.Close()
		// fmt.Println(path + "/" + st.Regions.Path + "/" + reg.File)
		w := fmt.Sprintf("select state from public.\"cross\" where region=%d order by region,area,id;", reg.ID)
		// fmt.Println(w)
		rows, err := con.Query(w)
		if err != nil {
			logger.Error.Printf("Error %s  %s\n", w, err.Error())
			return err
		}
		var c []byte
		for rows.Next() {
			state := new(pudge.Cross)
			err = rows.Scan(&c)
			if err != nil {
				logger.Error.Printf("%s\n", err.Error())
				return err
			}
			err = json.Unmarshal(c, &state)
			if err != nil {
				logger.Error.Printf("%s\n", err.Error())
				return err
			}
			str := fmt.Sprintf("@u,%d,1,%s%08d,%d,%d,%d,0\n", state.ID, state.ConType, state.IDevice, state.Area, state.SubArea, state.ID)
			_, _ = file.WriteString(str)
			_, _ = file.WriteString(fmt.Sprintf("@C,%s\n", state.Dgis))
			_, _ = file.WriteString(fmt.Sprintf("@S,%s\n", state.Name))
			str = fmt.Sprintf("@P,%d,%d,%d,%d,%d\n", state.Model.VPCPDL, state.Model.VPCPDR, state.Model.VPBSL, state.Model.VPBSR, state.NumDev)
			_, _ = file.WriteString(str)

			_, _ = file.WriteString(fmt.Sprintf("@N,%s\n", state.Phone))
			//Теперь начинаем выгружать массивы привязки
			if !state.Arrays.StatDefine.IsEmpty() {
				_, _ = file.WriteString(fmt.Sprintf("@k1,%s\n", toLine(state.Arrays.StatDefine.ToBuffer())))
			}
			if !state.Arrays.PointSet.IsEmpty() {
				_, _ = file.WriteString(fmt.Sprintf("@k1,%s\n", toLine(state.Arrays.PointSet.ToBuffer())))
			}
			if !state.Arrays.UseInput.IsEmpty() {
				_, _ = file.WriteString(fmt.Sprintf("@k1,%s\n", toLine(state.Arrays.UseInput.ToBuffer())))
			}
			if !state.Arrays.TimeDivice.IsEmpty() {
				_, _ = file.WriteString(fmt.Sprintf("@k1,%s\n", toLine(state.Arrays.TimeDivice.ToBuffer())))
			}
			if !state.Arrays.SetupDK.IsEmpty() {
				_, _ = file.WriteString(fmt.Sprintf("@k1,%s\n", toLine(state.Arrays.SetupDK.ToBuffer())))
			}
			for _, ws := range state.Arrays.WeekSets.WeekSets {
				if !ws.IsEmpty() {
					_, _ = file.WriteString(fmt.Sprintf("@k1,%s\n", toLine(ws.ToBuffer())))
				}
			}
			for _, ds := range state.Arrays.DaySets.DaySets {
				if !ds.IsEmpty() {
					_, _ = file.WriteString(fmt.Sprintf("@k1,%s\n", toLine(ds.ToBuffer())))
				}
			}
			for _, ms := range state.Arrays.MonthSets.MonthSets {
				if !ms.IsEmpty() {
					_, _ = file.WriteString(fmt.Sprintf("@k1,%s\n", toLine(ms.ToBuffer())))
				}
			}
			for i := 1; i < 13; i++ {
				if !state.Arrays.SetDK.IsEmpty(1, i) {
					_, _ = file.WriteString(fmt.Sprintf("@k1,%s\n", toLine(state.Arrays.SetDK.DK[i-1].ToBuffer())))
				}
			}
			if !state.Arrays.SetTimeUse.IsEmpty() {
				_, _ = file.WriteString(fmt.Sprintf("@k1,%s\n", toLine(state.Arrays.SetTimeUse.ToBuffer(148))))
			}
			if !state.Arrays.SetCtrl.IsEmpty() {
				_, _ = file.WriteString(fmt.Sprintf("@k1,%s\n", toLine(state.Arrays.SetCtrl.ToBuffer())))
			}
			if !state.Arrays.SetTimeUse.IsEmpty() {
				_, _ = file.WriteString(fmt.Sprintf("@k1,%s\n", toLine(state.Arrays.SetTimeUse.ToBuffer(157))))
			}
			_, _ = file.WriteString("\n")
		}
		file.Close()
	}
	err = saveDataRouters(path)
	if err != nil {
		return err
	}
	err = saveDataXctrl(path)
	if err != nil {
		return err
	}
	logger.Info.Println("Exit Save DB ...")
	fmt.Println("Exit Save DB...")
	return nil
}
func toLine(in []int) string {
	s := fmt.Sprintf("%v", in)
	s = strings.ReplaceAll(s, " ", ",")
	s = strings.ReplaceAll(s, "[", "")
	s = strings.ReplaceAll(s, "]", "")
	return s
}
func saveDataRouters(path string) error {
	file, err := os.OpenFile(path+"/routes."+st.SQL.Ext, os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		logger.Error.Printf("Error appending file %s! %s", path+"/routes."+st.SQL.Ext, err.Error())
		return err
	}
	defer file.Close()
	file.WriteString("\n")
	tabs, err := con.Query("select region,id,description,box,listtl from public.routes order by region; ")
	if err != nil {
		logger.Error.Printf("Error %s", err.Error())
		return err
	}
	for tabs.Next() {
		var region, id int
		var desc string
		var box, listtl []byte
		_ = tabs.Scan(&region, &id, &desc, &box, &listtl)
		if iregion > 0 && iregion != region {
			continue
		}
		w := fmt.Sprintf("insert into public.routes ( region, id, description, box, listtl) VALUES (%d,%d,'%s','%s','%s');\n",
			region, id, desc, string(box), string(listtl))
		file.WriteString(w)
	}
	return nil
}
func saveDataXctrl(path string) error {
	file, err := os.OpenFile(path+"/xctrl."+st.SQL.Ext, os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		logger.Error.Printf("Error appending file %s! %s", path+"/xctrl."+st.SQL.Ext, err.Error())
		return err
	}
	defer file.Close()
	file.WriteString("\n")
	tabs, err := con.Query("select region,area,subarea,state from public.xctrl order by region,area,subarea; ")
	if err != nil {
		logger.Error.Printf("Error %s", err.Error())
		return err
	}
	for tabs.Next() {
		var region, area, subarea int
		var state []byte
		_ = tabs.Scan(&region, &area, &subarea, &state)
		if iregion > 0 && iregion != region {
			continue
		}
		w := fmt.Sprintf("insert into public.xctrl ( region, area, subarea, state) VALUES (%d,%d,%d,'%s');\n",
			region, area, subarea, string(state))
		file.WriteString(w)
	}
	return nil
}
