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
