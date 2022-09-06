package sqlsave

import (
	"crypto/md5"
	"database/sql"
	"fmt"
	"hash"
	"io"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"

	_ "github.com/lib/pq"
	"github.com/ruraomsk/ag-server/logger"
	"github.com/ruraomsk/ag-server/setup"
)

type name struct {
	ColumnName string
	DataType   string
	IsKey      bool
}

type defTable struct {
	Records map[string]*Record
}
type Record struct {
	Hash hash.Hash
	Used bool
}

var tableFields map[string][]string
var tables map[string][]name
var useTables map[string]*defTable

func isTable(table string) bool {
	_, is := tableFields[table]
	return is
}
func addTable(table string) {
	_, is := tables[table]
	if is {
		return
	}
	tables[table] = make([]name, 0)
	t := new(defTable)
	t.Records = make(map[string]*Record)
	useTables[table] = t
}
func addName(table string, fieldname string, typeColumn string) {
	n := new(name)
	n.ColumnName = fieldname
	n.DataType = typeColumn
	m, is := tableFields[table]
	if is {
		for _, nn := range m {
			if strings.Compare(nn, n.ColumnName) == 0 {
				n.IsKey = true
				break
			}
		}
	}
	na := tables[table]
	na = append(na, *n)
	tables[table] = na
}

var dbb *sql.DB
var err error

//Start сохранение БД
func Start() {
	dbinfo := fmt.Sprintf("host=%s user=%s password=%s dbname=%s sslmode=disable",
		setup.Set.DataBase.Host, setup.Set.DataBase.User,
		setup.Set.DataBase.Password, setup.Set.DataBase.DBname)

	dbb, err = sql.Open("postgres", dbinfo)
	if err != nil {
		logger.Error.Printf("Error open conn %s", err.Error())
		return
	}
	defer dbb.Close()
	if err = dbb.Ping(); err != nil {
		logger.Error.Printf("Ping %s", err.Error())
		return
	}
	for {
		time.Sleep(time.Duration(setup.Set.Saver.Step) * time.Second)
		tableFields = make(map[string][]string)
		tables = make(map[string][]name)
		useTables = make(map[string]*defTable)
		//fmt.Printf("%v\n", setup.Set.Saver.Keys)
		for i := 0; i < len(setup.Set.Saver.Keys); i++ {
			nt := setup.Set.Saver.Keys[i][0]
			ns := make([]string, 0)
			for j := 1; j < len(setup.Set.Saver.Keys[i]); j++ {
				ns = append(ns, setup.Set.Saver.Keys[i][j])
			}
			tableFields[nt] = ns
		}
		//fmt.Printf("%v\n", tableFields)
		tabs, err := dbb.Query("SELECT table_name FROM information_schema.tables  WHERE table_schema='public' ORDER BY table_name;")
		if err != nil {
			logger.Error.Printf("Error %s", err.Error())
			return
		}
		for tabs.Next() {
			var nameTable string
			_ = tabs.Scan(&nameTable)
			//fmt.Printf("Table :%s", nameTable)

			if !isTable(nameTable) {
				//fmt.Printf(" ignored\n")
				continue
			}
			addTable(nameTable)

			w := fmt.Sprintf("select column_name,data_type from information_schema.columns WHERE table_schema='public' and table_name='%s' order by ordinal_position;", nameTable)
			cols, err := dbb.Query(w)
			if err != nil {
				logger.Error.Printf("Error %s", err.Error())
				return
			}
			for cols.Next() {
				n := new(name)
				cols.Scan(&n.ColumnName, &n.DataType)
				addName(nameTable, n.ColumnName, n.DataType)
			}
			_ = cols.Close()
		}
		file, err := os.Create(setup.Set.Saver.File)
		if err != nil {
			logger.Error.Printf("Error open file %s", err.Error())
			return
		}
		defer file.Close()
		for _, s := range setup.Set.Saver.PreSQL {
			_, _ = file.WriteString(s + "\n")
		}
		for nameTab, names := range tables {
			//fmt.Printf("Table %s\n", nameTab)
			w := fmt.Sprintf("select * from public.\"%s\";", nameTab)
			rows, err := dbb.Query(w)
			if err != nil {
				logger.Error.Printf("Error %s", err.Error())
				return
			}
			for rows.Next() {
				del, insert, update, key := readOneRecord(rows, nameTab, names)
				_, _ = file.WriteString(del + "\n")
				_, _ = file.WriteString(insert + "\n")
				//file.WriteString(update + "\n")
				r := new(Record)
				r.Hash = md5.New()
				_, _ = io.WriteString(r.Hash, update)
				useTables[nameTab].Records[key] = r
			}
		}
		_ = file.Close()

		if !sender() {
			continue
		}
		for {
			time.Sleep(time.Duration(setup.Set.Saver.Step) * time.Second)

			if !updater() {
				break
			}
			if !sender() {
				break
			}

		}
	}

}
func readOneRecord(rows *sql.Rows, nameTab string, names []name) (del string, insert string, update string, key string) {
	var err error
	sl := make([]interface{}, len(names))
	switch len(names) {
	case 1:
		err = rows.Scan(&sl[0])
	case 2:
		err = rows.Scan(&sl[0], &sl[1])
	case 3:
		err = rows.Scan(&sl[0], &sl[1], &sl[2])
	case 4:
		err = rows.Scan(&sl[0], &sl[1], &sl[2], &sl[3])
	case 5:
		err = rows.Scan(&sl[0], &sl[1], &sl[2], &sl[3], &sl[4])
	case 6:
		err = rows.Scan(&sl[0], &sl[1], &sl[2], &sl[3], &sl[4], &sl[5])
	case 7:
		err = rows.Scan(&sl[0], &sl[1], &sl[2], &sl[3], &sl[4], &sl[5], &sl[6])
	case 8:
		err = rows.Scan(&sl[0], &sl[1], &sl[2], &sl[3], &sl[4], &sl[5], &sl[6], &sl[7])
	case 9:
		err = rows.Scan(&sl[0], &sl[1], &sl[2], &sl[3], &sl[4], &sl[5], &sl[6], &sl[7], &sl[8])
	case 10:
		err = rows.Scan(&sl[0], &sl[1], &sl[2], &sl[3], &sl[4], &sl[5], &sl[6], &sl[7], &sl[8], &sl[9])
	case 11:
		err = rows.Scan(&sl[0], &sl[1], &sl[2], &sl[3], &sl[4], &sl[5], &sl[6], &sl[7], &sl[8], &sl[9], &sl[10])
	default:
		logger.Error.Printf("Добавьте полей! Нужно %d", len(names))
		return
	}
	if err != nil {
		logger.Error.Printf("Error of scan %s", err.Error())
		return

	}

	del = fmt.Sprintf("delete from public.\"%s\" where ", nameTab)
	insert = fmt.Sprintf("insert into public.\"%s\" (", nameTab)
	update = fmt.Sprintf("update public.\"%s\" set ", nameTab)
	for i, n := range names {
		if i > 0 {
			insert += "," + n.ColumnName
		} else {
			insert += n.ColumnName
		}
	}
	insert += ") values("
	var str string
	for i, n := range names {
		v := reflect.ValueOf(sl[i])
		switch n.DataType {
		case "integer":
			str = strconv.Itoa(int(v.Int()))
		case "jsonb":
			if v.IsValid() {
				str = "'" + string(v.Bytes()) + "'"
			} else {
				str = "'{}'"
			}
		case "boolean":
			str = strconv.FormatBool(v.Bool())
		case "point":
			str = "'" + string(v.Bytes()) + "'"
		case "text":
			str = "'" + v.String() + "'"
		case "timestamp with time zone":
			str = fmt.Sprintf("%v", v)
			str = strings.TrimLeft(str, "time:")
			str = str[0 : len(str)-6]
			str = "'" + str + "'"
			//fmt.Println(str)
		case "date":
			//fmt.Println(str)
			str = fmt.Sprintf("%v", v)
			str = "'" + str[0:10] + "'"
		default:
			logger.Error.Printf("DataType %s", n.DataType)
			str = "0"
		}
		if i > 0 {
			update += " ," + n.ColumnName + "=" + str
			insert += "," + str
			if n.IsKey {
				key += " and " + n.ColumnName + "=" + str
			}
		} else {
			update += n.ColumnName + "=" + str
			insert += str
			if n.IsKey {
				key += n.ColumnName + "=" + str
			}
		}
	}
	del += key + ";"
	insert += ");"
	update += " where " + key + ";"

	return
}
