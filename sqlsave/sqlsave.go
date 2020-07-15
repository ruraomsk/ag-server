package sqlsave

import (
	"database/sql"
	"fmt"
	"github.com/JanFant/TLServer/logger"
	_ "github.com/lib/pq"
	"github.com/ruraomsk/ag-server/setup"
	"os"
	"reflect"
	"strconv"
	"strings"
)

type name struct {
	ColumnName string
	DataType   string
	IsKey      bool
}

var tableFields map[string][]string
var tables map[string][]name

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

func Start() {
	dbinfo := fmt.Sprintf("host=%s user=%s password=%s dbname=%s sslmode=disable",
		setup.Set.DataBase.Host, setup.Set.DataBase.User,
		setup.Set.DataBase.Password, setup.Set.DataBase.DBname)

	dbb, err := sql.Open("postgres", dbinfo)
	if err != nil {
		logger.Error.Printf("Error open conn %s", err.Error())
		return
	}
	defer dbb.Close()
	if err = dbb.Ping(); err != nil {
		logger.Error.Printf("Ping %s", err.Error())
		return
	}
	tableFields = make(map[string][]string)
	tables = make(map[string][]name)
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
		tabs.Scan(&nameTable)
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
		cols.Close()
	}
	file, err := os.Create("result.sql")
	if err != nil {
		logger.Error.Printf("Error open file %s", err.Error())
		return
	}
	defer file.Close()
	for nameTab, names := range tables {
		fmt.Printf("Table %s\n", nameTab)
		w := fmt.Sprintf("select * from public.%s;", nameTab)
		rows, err := dbb.Query(w)
		if err != nil {
			logger.Error.Printf("Error %s", err.Error())
			return
		}
		for rows.Next() {
			//var err error
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
			default:
				logger.Error.Printf("Len >9! Is = %d", len(names))
				return
			}
			d := fmt.Sprintf("delete from public.%s where ", nameTab)
			w = fmt.Sprintf("insert into public.%s (", nameTab)
			for i, n := range names {
				if i > 0 {
					w += "," + n.ColumnName
				} else {
					w += n.ColumnName
				}
			}
			w += ") values("
			var str string
			for i, n := range names {
				v := reflect.ValueOf(sl[i])
				switch n.DataType {
				case "integer":
					str = strconv.Itoa(int(v.Int()))
				case "jsonb":
					str = "'" + string(v.Bytes()) + "'"
				case "boolean":
					str = strconv.FormatBool(v.Bool())
				case "point":
					str = "'" + string(v.Bytes()) + "'"
				case "text":
					str = "'" + v.String() + "'"
				case "timestamp with time zone":
					str = fmt.Sprintf("%v", v)
					str = "'" + strings.TrimLeft(str, "time:") + "'"
					//fmt.Println(str)
				default:
					logger.Error.Printf("DataType %s", n.DataType)
					str = "0"
				}
				if i > 0 {
					w += "," + str
					if n.IsKey {
						d += " and " + n.ColumnName + "=" + str
					}
				} else {
					w += str
					if n.IsKey {
						d += n.ColumnName + "=" + str
					}
				}
			}
			d += ";"
			w += ");"
			file.WriteString(d + "\n")
			file.WriteString(w + "\n")
		}
	}
	file.Close()

}
