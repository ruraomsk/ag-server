package main

import (
	"reflect"
	"rura/ag-server/binding"
	"testing"
)

func Test_Days(t *testing.T) {
	buffer := []int{65, 0, 137, 39, 1, 9, 8, 6, 0, 4, 7, 0, 9, 10, 0, 5, 11, 30, 1, 14, 0, 6, 16, 0, 2, 19, 45, 3, 22, 0, 4, 24, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 255}

	td := binding.NewDaySet()
	err := td.FromBuffer(buffer)
	if err != nil {
		t.Error(err.Error())
		return
	}
	out := td.DaySets[0].ToBuffer()
	if !reflect.DeepEqual(&buffer, &out) {
		t.Errorf("%v\n", td)
		t.Errorf("No equal\n%v\n%v", buffer, out)
		return
	}
}
func Test_Nedels(t *testing.T) {
	buffer := []int{45, 0, 8, 8, 1, 5, 1, 1, 1, 2, 3, 4}

	td := binding.NewNedelSets()
	err := td.FromBuffer(buffer)
	if err != nil {
		t.Error(err.Error())
		return
	}
	out := td.NedelSets[0].ToBuffer()
	if !reflect.DeepEqual(&buffer, &out) {
		t.Errorf("No equal\n%v\n%v", buffer, out)
		return
	}
}
func Test_Years(t *testing.T) {
	buffer := []int{85, 0, 22, 32, 1, 2, 2, 2, 2, 2, 2, 2, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1}

	td := binding.NewYearSets()
	err := td.FromBuffer(buffer)
	if err != nil {
		t.Error(err.Error())
		return
	}
	out := td.MonthSets[0].ToBuffer()
	if !reflect.DeepEqual(&buffer, &out) {
		t.Errorf("No equal\n%v\n%v", buffer, out)
		return
	}
}
func Test_StatDefine(t *testing.T) {
	buffer := []int{14, 0, 14, 9, 0, 1, 5, 6, 0, 0, 0, 0, 0}

	td := binding.NewStatDefine()
	err := td.FromBuffer(buffer)
	if err != nil {
		t.Error(err.Error())
		return
	}
	out := td.ToBuffer()
	if !reflect.DeepEqual(&buffer, &out) {
		t.Errorf("No equal\n%v\n%v", buffer, out)
		return
	}
}
func Test_PointSet(t *testing.T) {
	buffer := []int{15, 0, 15, 13, 0, 1, 1, 2, 1, 3, 1, 4, 1, 5, 1, 6, 1}

	td := binding.NewPointSet()
	err := td.FromBuffer(buffer)
	if err != nil {
		t.Error(err.Error())
		return
	}
	out := td.ToBuffer()
	if !reflect.DeepEqual(&buffer, &out) {
		t.Errorf("No equal\n%v\n%v", buffer, out)
		return
	}
}
func Test_UseInput(t *testing.T) {
	buffer := []int{16, 0, 16, 9, 0, 1, 1, 1, 1, 1, 1, 1, 1}

	td := binding.NewUseInput()
	err := td.FromBuffer(buffer)
	if err != nil {
		t.Error(err.Error())
		return
	}
	out := td.ToBuffer()
	if !reflect.DeepEqual(&buffer, &out) {
		t.Errorf("No equal\n%v\n%v", buffer, out)
		return
	}
}
func Test_TimeDevice(t *testing.T) {
	buffer := []int{21, 0, 21, 5, 1, 6, 0, 0, 0}

	td := binding.NewTimeDevice()
	err := td.FromBuffer(buffer)
	if err != nil {
		t.Error(err.Error())
		return
	}
	out := td.ToBuffer()
	if !reflect.DeepEqual(&buffer, &out) {
		t.Errorf("No equal\n%v\n%v", buffer, out)
		return
	}
}
func Test_SetupDK(t *testing.T) {
	buffer := []int{40, 0, 7, 10, 1, 6, 1, 0, 40, 1, 7, 1, 12, 1}
	buff := []int{41, 0, 7, 10, 2, 6, 1, 1, 40, 2, 7, 201, 12, 1}

	td := binding.NewSetupDK()
	err := td.FromBuffer(buffer)
	if err != nil {
		t.Error(err.Error())
		return
	}
	out := td.ToBuffer()
	if !reflect.DeepEqual(&buffer, &out) {
		t.Errorf("No equal\n%v\n%v", buffer, out)
		return
	}
	err = td.FromBuffer(buff)
	if err != nil {
		t.Error(err.Error())
		return
	}
	out = td.ToBuffer()
	if !reflect.DeepEqual(&buff, &out) {
		t.Errorf("No equal\n%v\n%v", buffer, out)
		return
	}
}
