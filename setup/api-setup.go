package setup

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"github.com/ruraomsk/ag-server/logger"
)

//Set переменная для хранения текущих настроек
var Set Setup

//Setuppath путь к файлу
var Setuppath string

//LoadSetUp загружает настройки из json
func LoadSetUp(path string) error {
	buf, err := ioutil.ReadFile(path)
	if err != nil {
		logger.Error.Printf("Error reading file %s! %s\n", path, err.Error())
		return err
	}
	err = json.Unmarshal(buf, &Set)
	if err != nil {
		logger.Error.Printf("Error reading json! %s\n", err.Error())
		return err
	}
	Setuppath = path
	return nil
}

//WriteSetUp записывает в json настройки
func WriteSetUp() error {
	if len(Setuppath) == 0 {
		return fmt.Errorf("настройки не загружались")
	}
	result, err := json.MarshalIndent(&Set, "", "\t")
	if err != nil {
		return fmt.Errorf("error creating json %s", err.Error())
	}
	file, err := os.Create(Setuppath)
	if err != nil {
		return fmt.Errorf("error creating file %s %s", Setuppath, err.Error())
	}
	defer file.Close()
	_, err = file.Write(result)

	return err
}

//ToJSON переводит в Json
func ToJSON() (result []byte, err error) {
	result, err = json.Marshal(&Set)
	return result, err
}

//FromJSON приниает json
func FromJSON(result []byte) error {
	err := json.Unmarshal(result, &Set)
	return err

}
