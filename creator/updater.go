package creator

import (
	"bufio"
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"

	"github.com/ruraomsk/ag-server/logger"
	"github.com/ruraomsk/ag-server/pudge"
	"github.com/ruraomsk/ag-server/setup"
)

func Update(reg, nfile string) error {
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

	file, err := ioutil.ReadFile(nfile)
	if err != nil {
		logger.Error.Printf("Error reading file %s! %s\n", nfile, err.Error())
		return err
	}
	region, _ := strconv.Atoi(reg)
	logger.Info.Printf("Updater:Обрабатываем файл %s", nfile)
	scanner := bufio.NewScanner(bytes.NewReader(file))
	isempty := true
	state := pudge.NewCross()
	var dgis string
	var i148 []int
	var short148 bool
	for scanner.Scan() {
		str := scanner.Text()
		if len(str) == 0 {
			continue
		}
		if strings.HasPrefix(str, "@u,") {
			//Начало нового перекрестка
			if !isempty {
				_ = updateState(state, dgis)
			}
			short148 = false
			i148 = make([]int, 32)
			i148[0] = 148
			i148[2] = 23
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
			state.ConType = ss[3][0:2]
			state.IDevice, _ = strconv.Atoi(ss[3][2:])
			continue
		}
		if strings.HasPrefix(str, "@P,") {
			//Тип устройствва
			ss := strings.Split(str, ",")
			if len(ss) != 6 {
				logger.Error.Printf("в строке %s неверное число параметров", str)
				return err
			}
			state.NumDev, _ = strconv.Atoi(ss[5])
			state.Arrays.TypeDevice = state.NumDev
			state.Model.VPCPDL, _ = strconv.Atoi(ss[1])
			state.Model.VPCPDR, _ = strconv.Atoi(ss[2])
			state.Model.VPBSL, _ = strconv.Atoi(ss[3])
			state.Model.VPBSR, _ = strconv.Atoi(ss[4])
			switch state.NumDev {
			case 1:
				state.Model.C12 = true
			case 2:
			case 4:
				state.Model.DKA = true
			case 8:
				state.Model.DTA = true
			}
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
			state.Phone = str[3:]
			continue
		}
		if strings.HasPrefix(str, "@k1,") {
			//Массив
			str = strings.ReplaceAll(str, " ", "")
			ss := strings.Split(str, ",")
			sint := make([]int, 0)
			for i := 1; i < len(ss); i++ {
				ii, _ := strconv.Atoi(ss[i])
				sint = append(sint, ii)
			}
			if sint[0] == 14 {
				err = state.Arrays.StatDefine.FromBuffer(sint)
				if err != nil {
					logger.Error.Printf("в строке %s %s", str, err.Error())
					return err
				}
				continue
			}
			if sint[0] == 15 {
				err = state.Arrays.PointSet.FromBuffer(sint)
				if err != nil {
					logger.Error.Printf("в строке %s %s", str, err.Error())
					return err
				}
				continue
			}
			if sint[0] == 16 {
				err = state.Arrays.UseInput.FromBuffer(sint)
				if err != nil {
					logger.Error.Printf("в строке %s %s", str, err.Error())
					return err
				}
				continue
			}
			if sint[0] == 21 {
				err = state.Arrays.TimeDivice.FromBuffer(sint)
				if err != nil {
					logger.Error.Printf("в строке %s %s", str, err.Error())
					return err
				}
				continue
			}
			if sint[0] == 40 {
				err = state.Arrays.SetupDK.FromBuffer(sint)
				if err != nil {
					logger.Error.Printf("в строке %s %s", str, err.Error())
					return err
				}
				continue
			}
			if sint[0] >= 45 && sint[0] <= 56 {
				//Недельные массивы
				err = state.Arrays.WeekSets.FromBuffer(sint)
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
					// return err

				}
				continue
			}
			if sint[0] >= 140 && sint[0] <= 147 {
				if sint[2] == 23 && sint[3] == 4 {
					short148 = true
					p := 5 + ((sint[4] - 1) * 3)
					i148[p] = sint[5]
					i148[p+1] = sint[6]
					i148[p+2] = sint[7]
					continue
				}
			}
			if sint[0] == 148 {
				if short148 || sint[3] == 4 {
					if sint[2] == 23 && sint[3] == 4 {
						p := 5 + ((sint[4] - 1) * 3)
						i148[p] = sint[5]
						i148[p+1] = sint[6]
						i148[p+2] = sint[7]
					}
					err = state.Arrays.SetTimeUse.FromBuffer(i148)
				} else {
					err = state.Arrays.SetTimeUse.FromBuffer(sint)
				}
				// Массив настройки времен внешних входов
				if err != nil {
					logger.Error.Printf("в строке %s %s", str, err.Error())
					// return err
				}
				continue
			}
			if sint[0] == 157 {
				// Массив настройки времен внешних входов
				err = state.Arrays.SetTimeUse.FromBuffer(sint)
				if err != nil {
					logger.Error.Printf("в строке %s %s", str, err.Error())
					return err
				}
				continue
			}
			if sint[0] == 149 {
				//Массив контроля входов
				err = state.Arrays.SetCtrl.FromBuffer(sint)
				if err != nil {
					logger.Error.Printf("в строке %s %s", str, err.Error())
					return err
				}
				continue
			}
			logger.Info.Printf("в строке %s нет такого обработчика", str)

		}

	}
	if !isempty {
		_ = updateState(state, dgis)
	}
	return nil
}
func updateState(state *pudge.Cross, dgis string) error {

	w := fmt.Sprintf("select state from public.\"cross\" where region=%d and id=%d;",
		state.Region, state.ID)
	find, err := con.Query(w)
	if err != nil {
		logger.Error.Printf("%s %s\n", w, err.Error())
		return err
	}
	found := false
	var c []byte
	var oldState pudge.Cross
	for find.Next() {
		found = true
		err = find.Scan(&c)
		if err != nil {
			logger.Error.Printf("%s\n", err.Error())
			return err
		}
		err = json.Unmarshal(c, &oldState)
		if err != nil {
			logger.Error.Printf("%s\n", err.Error())
			return err
		}
	}
	if !found {
		logger.Info.Printf("Пропущен %d %d %d", state.Region, state.Area, state.ID)
		return nil
	}

	if state.NumDev == 0 {
		state.NumDev = 2
		state.Arrays.TypeDevice = state.NumDev
	}
	state.Dgis = oldState.Dgis
	state.Name = oldState.Name
	state.Area = oldState.Area
	state.Name = oldState.Name
	//w = fmt.Sprintf("delete from public.\"cross\" where region=%d and area=%d and id=%d);",
	//	state.Region, state.Area,  state.ID)
	//_, err = con.Exec(w)
	//if err != nil {
	//	logger.Error.Printf("Error %s  %s\n", w, err.Error())
	//	return err
	//}

	b, err := json.Marshal(&state)
	if err != nil {
		logger.Error.Printf("%s\n", err.Error())
		return err
	}

	w = fmt.Sprintf("update public.\"cross\" set idevice=%d, state='%s' "+
		"where region=%d and area=%d and id=%d;",
		state.IDevice, string(b), state.Region, state.Area, state.ID)
	_, err = con.Exec(w)

	if err != nil {
		logger.Error.Printf("Error %s  %s\n", w, err.Error())
		return err
	}
	return nil
}
