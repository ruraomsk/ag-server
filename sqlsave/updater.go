package sqlsave

import (
	"crypto/md5"
	"fmt"
	"github.com/JanFant/TLServer/logger"
	"github.com/ruraomsk/ag-server/setup"
	"io"
	"reflect"

	"os"
)

func updater(path string) bool {

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
		w := fmt.Sprintf("select * from public.%s;", nameTab)
		rows, err := dbb.Query(w)
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
				io.WriteString(r.Hash, update)
				useTables[nameTab].Records[key] = r
				file.WriteString(insert + "\n")
			} else {
				//запись есть
				hnew := md5.New()
				io.WriteString(hnew, update)
				rec.Used = true
				if reflect.DeepEqual(&rec.Hash, &hnew) {
					file.WriteString(update + "\n")
					rec.Hash = hnew
				}
			}
		}
		rows.Close()

		for k, rec := range dTable.Records {
			if !rec.Used {
				file.WriteString(fmt.Sprintf("delete from public.%s where %s;\n", nameTab, k))
				delete(useTables[nameTab].Records, k)
			}
		}
	}
	file.Close()
	//Coda
	return true
}
