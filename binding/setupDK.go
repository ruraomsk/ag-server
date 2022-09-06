package binding

import (
	"fmt"
	"reflect"
)

//SetupDK настройка ДК
type SetupDK struct {
	DKNum    int  `json:"dkn"`     //Номер ДК
	TMaxF    int  `json:"tmaxf"`   //Максимальное время ожидания смены фаз
	TMinF    int  `json:"tminf"`   //Минимальное время ожидания смены фаз
	TmaxTmin int  `json:"tminmax"` // Максимальное время ожидания включения фазы
	DKType   int  `json:"dktype"`  //Тип ДК
	ExtNum   int  `json:"extn"`    //Внешний номер ДК
	Tprom    int  `json:"tprom"`   //Максимальное время промежуточного такта
	IsPreset bool `json:"preset"`  // Прищнак наличия контроллера на линии
}

//Compare сравнивание истина если равны
func (sdk *SetupDK) Compare(tdk *SetupDK) bool {
	return reflect.DeepEqual(sdk, tdk)
}

//NewSetupDK создает новый набор настройки ДК
func NewSetupDK() *SetupDK {
	r := new(SetupDK)
	return r
}

//FromBuffer переводит из массива кодов в структуру1
func (sdk *SetupDK) FromBuffer(buffer []int) error {
	if len(buffer) != 14 {
		return fmt.Errorf("неверная длина массива")
	}
	if buffer[0] != 40 && buffer[0] != 41 {
		return fmt.Errorf("несовпал номер массива на сервере и номер массива")
	}
	if buffer[2] != 7 {
		return fmt.Errorf("неверный номер массива")
	}
	sdk.DKNum = 1
	sdk.TMaxF = buffer[5]
	sdk.TMinF = buffer[6]
	// if buffer[7] != buffer[4]-1 {
	// 	return fmt.Errorf("неверный смещение массива ДК")
	// }
	sdk.TmaxTmin = buffer[8]
	// if buffer[9] != buffer[4] {
	// 	return fmt.Errorf("неверный номер ДК массива ")
	// }
	sdk.DKType = buffer[10]
	sdk.ExtNum = buffer[11]
	sdk.Tprom = buffer[12]
	sdk.IsPreset = false
	if buffer[13] != 0 {
		sdk.IsPreset = true
	}

	return nil
}

//IsEmpty возвращает истину если данный массив пустой
func (sdk *SetupDK) IsEmpty() bool {
	return false
}

//ToBuffer переводит в буфер кодов
func (sdk *SetupDK) ToBuffer() []int {
	r := make([]int, 14)
	r[0] = 40
	if sdk.DKNum == 2 {
		r[0] = 41
	}
	r[1] = 0
	r[2] = 7
	r[3] = 10
	r[4] = 1
	//sdk.ExtNum
	r[5] = sdk.TMaxF
	r[6] = sdk.TMinF
	r[7] = 0
	r[8] = sdk.TmaxTmin
	r[9] = 1
	r[10] = sdk.DKType
	r[11] = sdk.ExtNum
	r[12] = sdk.Tprom
	if sdk.IsPreset {
		r[13] = 1
	}
	return r
}
