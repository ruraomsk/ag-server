package sqlsave

import (
	"crypto/md5"
	"fmt"
	"io"
	"reflect"

	"github.com/ruraomsk/ag-server/logger"
	"github.com/ruraomsk/ag-server/setup"

	"os"
)

func updater() bool {

	file, err := os.Create(setup.Set.Saver.File)
	if err != nil {
		logger.Error.Printf("Error open file %s", err.Error())
		return false
	}
	defer file.Close()
	for _, dTable := range useTables {
		for _, rec := range dTable.Records {
			rec.Used = false
		}
	}
	for nameTab, dTable := range useTables {
		//fmt.Printf("Table %s %v",nameTab,dTable.Records)
		rows, err := dbb.Query(fmt.Sprintf("select * from public.\"%s\";", nameTab))
		if err != nil {
			logger.Error.Printf("Error %s", err.Error())
			return false
		}
		names := tables[nameTab]
		for rows.Next() {
			_, insert, update, key := readOneRecord(rows, nameTab, names)
			rec, is := dTable.Records[key]
			if !is {
				//Новая запись
				r := new(Record)
				r.Hash = md5.New()
				r.Used = true
				_, _ = io.WriteString(r.Hash, update)
				useTables[nameTab].Records[key] = r
				_, _ = file.WriteString(insert + "\n")
			} else {
				//запись есть
				hnew := md5.New()
				_, _ = io.WriteString(hnew, update)
				rec.Used = true
				if !reflect.DeepEqual(&rec.Hash, &hnew) {
					_, _ = file.WriteString(update + "\n")
					rec.Hash = hnew
				}
			}
		}
		_ = rows.Close()

		for k, rec := range dTable.Records {
			if !rec.Used {
				_, _ = file.WriteString(fmt.Sprintf("delete from public.\"%s\" where %s;\n", nameTab, k))
				delete(useTables[nameTab].Records, k)
			}
		}
	}
	file.Close()
	//Coda
	return true
}
