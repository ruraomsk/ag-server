package creator

import (
	"database/sql"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/JanFant/TLServer/logger"
	"github.com/ruraomsk/ag-server/pudge"
	"github.com/ruraomsk/ag-server/setup"
)

//SaveAll сохраняет всю БД для правки в символьном виде
func SaveAll(path string) error {
	var st Creator
	var con *sql.DB
	logger.Info.Println("Start Save DB...")
	fmt.Println("Start Save DB...")
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
	for _, reg := range st.Regions.Regs {
		file, err := os.Create(path + "/" + st.Regions.Path + "/" + reg.File)
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
			file.WriteString(str)
			file.WriteString(fmt.Sprintf("@C,%s\n", state.Dgis))
			file.WriteString(fmt.Sprintf("@S,%s\n", state.Name))
			file.WriteString(fmt.Sprintf("@N,%s\n", state.Phone))
			//Теперь начинаем выгружать массивы привязки
			if !state.Arrays.StatDefine.IsEmpty() {
				file.WriteString(fmt.Sprintf("@k1,%s\n", toLine(state.Arrays.StatDefine.ToBuffer())))
			}
			if !state.Arrays.PointSet.IsEmpty() {
				file.WriteString(fmt.Sprintf("@k1,%s\n", toLine(state.Arrays.PointSet.ToBuffer())))
			}
			if !state.Arrays.UseInput.IsEmpty() {
				file.WriteString(fmt.Sprintf("@k1,%s\n", toLine(state.Arrays.UseInput.ToBuffer())))
			}
			if !state.Arrays.TimeDivice.IsEmpty() {
				file.WriteString(fmt.Sprintf("@k1,%s\n", toLine(state.Arrays.TimeDivice.ToBuffer())))
			}
			if !state.Arrays.SetupDK.IsEmpty() {
				file.WriteString(fmt.Sprintf("@k1,%s\n", toLine(state.Arrays.SetupDK.ToBuffer())))
			}
			for _, ws := range state.Arrays.WeekSets.WeekSets {
				if !ws.IsEmpty() {
					file.WriteString(fmt.Sprintf("@k1,%s\n", toLine(ws.ToBuffer())))
				}
			}
			for _, ds := range state.Arrays.DaySets.DaySets {
				if !ds.IsEmpty() {
					file.WriteString(fmt.Sprintf("@k1,%s\n", toLine(ds.ToBuffer())))
				}
			}
			for _, ms := range state.Arrays.MonthSets.MonthSets {
				if !ms.IsEmpty() {
					file.WriteString(fmt.Sprintf("@k1,%s\n", toLine(ms.ToBuffer())))
				}
			}
			for i := 1; i < 13; i++ {
				if !state.Arrays.SetDK.IsEmpty(1, i) {
					file.WriteString(fmt.Sprintf("@k1,%s\n", toLine(state.Arrays.SetDK.DK[i-1].ToBuffer())))
				}
			}
			if !state.Arrays.SetTimeUse.IsEmpty() {
				file.WriteString(fmt.Sprintf("@k1,%s\n", toLine(state.Arrays.SetTimeUse.ToBuffer(148))))
			}
			if !state.Arrays.SetCtrl.IsEmpty() {
				file.WriteString(fmt.Sprintf("@k1,%s\n", toLine(state.Arrays.SetCtrl.ToBuffer())))
			}
			if !state.Arrays.SetTimeUse.IsEmpty() {
				file.WriteString(fmt.Sprintf("@k1,%s\n", toLine(state.Arrays.SetTimeUse.ToBuffer(157))))
			}
			file.WriteString("\n")
		}
		file.Close()
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
