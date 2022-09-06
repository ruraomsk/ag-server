package creator

import (
	"bufio"
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/ruraomsk/ag-server/logger"
	"github.com/ruraomsk/ag-server/pudge"
	"github.com/ruraomsk/ag-server/setup"
)

type code struct {
	old_region pudge.Region
	new_region pudge.Region
	subarea    int
	name       string
}

var (
	message = `
	Вы начинаете перенос нумерации перекрестков 
	для региона %d и файла задания %s
	Вы уверены?
	`
	killCopyCross = `DROP TABLE IF EXISTS crosscopy;`
	copyCross     = `CREATE TABLE crosscopy AS
	TABLE public."cross";`
	crosses    map[pudge.Region]code
	newcrosses map[pudge.Region]pudge.Region
	dontmove   map[pudge.Region]bool
	db         *sql.DB
)

func Mover(region int, path string) {
	fmt.Printf(message, region, path)
	var repl string
	fmt.Scan(&repl)
	if strings.Compare(strings.ToUpper(repl), "ДА") != 0 {
		fmt.Println("Спасибо что одумались!")
		os.Exit(-1)
	}
	// Загружаем таблицу переходов
	buffer, err := ioutil.ReadFile(path)
	if err != nil {
		fmt.Printf("Error reading file %s! %s\n", path, err.Error())
		os.Exit(-1)
	}
	crosses = make(map[pudge.Region]code)
	newcrosses = make(map[pudge.Region]pudge.Region)
	dontmove = make(map[pudge.Region]bool)
	fmt.Println("Читаем управляющий файл...")
	var area, id int
	scanner := bufio.NewScanner(bytes.NewReader(buffer))

	for scanner.Scan() {
		str := scanner.Text()
		if len(str) == 0 {
			continue
		}
		if strings.HasPrefix(str, "@") {
			fmt.Printf("Пропускаем %s\n", str)
			continue
		}
		strs := strings.Split(str, "\t")
		if len(strs) != 5 {
			fmt.Printf("Не верная строка %s\n", str)
			os.Exit(-1)
		}
		area, _ = strconv.Atoi(strs[0])
		subarea, _ := strconv.Atoi(strs[1])
		newid, _ := strconv.Atoi(strs[2])
		oldid := 0
		if len(strs[3]) != 0 {
			oldid, _ = strconv.Atoi(strs[3])
		}
		name := strings.ReplaceAll(strs[4], "\"", "")
		if oldid == 0 {
			fmt.Printf("Пропускаем %s\n", str)
			continue
		}
		val := code{old_region: pudge.Region{Region: region, Area: area, ID: oldid},
			new_region: pudge.Region{Region: region, Area: area, ID: newid},
			subarea:    subarea, name: name}
		crosses[val.old_region] = val
		newcrosses[val.new_region] = val.old_region
	}
	dbinfo := fmt.Sprintf("host=%s user=%s password=%s dbname=%s sslmode=disable",
		setup.Set.DataBase.Host, setup.Set.DataBase.User,
		setup.Set.DataBase.Password, setup.Set.DataBase.DBname)
	for {
		db, err = sql.Open("postgres", dbinfo)
		if err != nil {
			logger.Error.Print(err.Error())
			time.Sleep(10 * time.Second)
			continue
		}
		err = db.Ping()
		if err != nil {
			logger.Error.Print(err.Error())
			time.Sleep(10 * time.Second)
			continue
		}
		break
	}
	fmt.Println("Соединение с базой данных установлено...")
	db.Exec(killCopyCross)
	db.Exec(copyCross)
	fmt.Println("Создана копия таблицы перекрестков...")
	//чистим ее от чужих регионов
	db.Exec("delete from public.crosscopy where region<>$1;", region)
	fmt.Println("Проверяем наличие перекрестков...")
	count := 0
	for _, v := range crosses {
		if !isCross(v.old_region) {
			fmt.Printf("Нет такого перекрестка %v\n", v.old_region)
			count++
		}
	}
	if count != 0 {
		fmt.Println("Работа прекращена!")
		os.Exit(-1)
	}
	fmt.Println("Проверяем какие перекрестки не вошли в перестановки и возможные конфликты...")
	rows, _ := db.Query("select area,id from public.crosscopy;")
	for rows.Next() {
		rows.Scan(&area, &id)
		reg := pudge.Region{Region: region, Area: area, ID: id}
		if _, is := crosses[reg]; !is {

			if _, is := newcrosses[reg]; is {
				fmt.Printf("кофликт остающегося %v с переводом [%v -> %v]\n", reg, newcrosses[reg], reg)
				count++
			} else {
				dontmove[reg] = true
			}
		}
	}
	rows.Close()
	if count != 0 {
		fmt.Println("Работа прекращена!")
		os.Exit(-1)
	}
	//Делаем копию рисунков
	os.RemoveAll("svg")
	err = copyDir(fmt.Sprintf("%s/%d", setup.Set.Dumper.PathSVG, region), "svg")
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(-1)
	}
	//Удаляем в tlserver
	os.RemoveAll(fmt.Sprintf("%s/%d", setup.Set.Dumper.PathSVG, region))
	fmt.Println("Переносим карты перекрестков ")
	for _, v := range crosses {
		dst := fmt.Sprintf("%s/%d/%d/%d", setup.Set.Dumper.PathSVG, v.new_region.Region, v.new_region.Area, v.new_region.ID)
		src := fmt.Sprintf("%s/%d/%d", "svg", v.old_region.Area, v.old_region.ID)
		fmt.Printf("%v -> %v\n", v.old_region, v.new_region)
		err = copyDir(src, dst)
		if err != nil {
			fmt.Println(err.Error())
		}
	}
	fmt.Println("Оставляем без изменения карты перекрестков ")
	for reg := range dontmove {
		dst := fmt.Sprintf("%s/%d/%d/%d", setup.Set.Dumper.PathSVG, reg.Region, reg.Area, reg.ID)
		src := fmt.Sprintf("%s/%d/%d", "svg", reg.Area, reg.ID)
		fmt.Println(reg)
		err = copyDir(src, dst)
		if err != nil {
			fmt.Println(err.Error())
		}
	}
	//Начинаем переносить данные БД
	//Вначале удаляем в cross весь регион
	db.Exec("delete from public.\"cross\" where region=$1", region)
	fmt.Println("Переносим данные новых перекрестков ")
	for _, v := range crosses {
		fmt.Printf("%v -> %v\n", v.old_region, v.new_region)
		cross, err := getCrossFromCopy(v.old_region)
		if err != nil {
			fmt.Println(err.Error())
			continue
		}
		cross.Area = v.new_region.Area
		cross.ID = v.new_region.ID
		cross.SubArea = v.subarea
		cross.Name = fmt.Sprintf("ДК %d %s", v.new_region.ID, v.name)
		err = setCross(cross)
		if err != nil {
			fmt.Println(err.Error())
		}
	}
	fmt.Println("Добавляем данные без изменения перекрестков ")
	for reg := range dontmove {
		fmt.Println(reg)

		cross, err := getCrossFromCopy(reg)
		if err != nil {
			fmt.Println(err.Error())
			continue
		}
		cross.Name = strings.ReplaceAll(cross.Name, ",", " ")
		for strings.Contains(cross.Name, "  ") {
			cross.Name = strings.ReplaceAll(cross.Name, "  ", " ")
		}
		names := strings.Split(cross.Name, " ")
		if len(names) > 3 {
			name := ""
			for i := 3; i < len(names); i++ {
				name += names[i] + " "
			}
			cross.Name = fmt.Sprintf("ДК %d %s", reg.ID, name)
		}
		err = setCross(cross)
		if err != nil {
			fmt.Println(err.Error())
		}
	}

}
func isCross(region pudge.Region) bool {
	rows, err := db.Query("select count(*) from crosscopy where region=$1 and area=$2 and id=$3", region.Region, region.Area, region.ID)
	if err != nil {
		panic(err.Error())
	}
	count := 0
	for rows.Next() {
		rows.Scan(&count)
	}
	return count == 1
}
func getCrossFromCopy(reg pudge.Region) (pudge.Cross, error) {
	rows, err := db.Query("select state from crosscopy where region=$1 and area=$2 and id=$3", reg.Region, reg.Area, reg.ID)
	if err != nil {
		return pudge.Cross{}, err
	}
	var buff []byte
	var state pudge.Cross
	found := false
	for rows.Next() {
		rows.Scan(&buff)
		err = json.Unmarshal(buff, &state)
		if err != nil {
			return pudge.Cross{}, err
		}
		found = true
	}
	if found {
		return state, nil
	}
	return pudge.Cross{}, fmt.Errorf("not found %v", reg)
}
func setCross(c pudge.Cross) error {
	js, _ := json.Marshal(c)
	w := fmt.Sprintf("insert into public.\"cross\" (region,area,subarea,id,dgis,describ,idevice,status,state) values(%d,%d,%d,%d,point(%s),'%s',%d,%d,'%s');",
		c.Region, c.Area, c.SubArea, c.ID, c.Dgis, c.Name, c.IDevice, c.StatusDevice, string(js))
	_, err = db.Exec(w)
	return err
}
func copyDir(src string, dst string) error {
	var err error
	var fds []os.FileInfo
	var srcinfo os.FileInfo
	if srcinfo, err = os.Stat(src); err != nil {
		return err
	}

	if err = os.MkdirAll(dst, srcinfo.Mode()); err != nil {
		return err
	}

	if fds, err = ioutil.ReadDir(src); err != nil {
		return err
	}
	for _, fd := range fds {
		srcfp := path.Join(src, fd.Name())
		dstfp := path.Join(dst, fd.Name())

		if fd.IsDir() {
			if err = copyDir(srcfp, dstfp); err != nil {
				fmt.Println(err)
			}
		} else {
			if err = copyFile(srcfp, dstfp); err != nil {
				fmt.Println(err)
			}
		}
	}
	return nil
}
func copyFile(src, dst string) error {
	var err error
	var srcfd *os.File
	var dstfd *os.File
	var srcinfo os.FileInfo

	if srcfd, err = os.Open(src); err != nil {
		return err
	}
	defer srcfd.Close()

	if dstfd, err = os.Create(dst); err != nil {
		return err
	}
	defer dstfd.Close()

	if _, err = io.Copy(dstfd, srcfd); err != nil {
		return err
	}
	if srcinfo, err = os.Stat(src); err != nil {
		return err
	}
	return os.Chmod(dst, srcinfo.Mode())
}
