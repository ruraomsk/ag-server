package main

import (
	"reflect"
	"testing"

	"github.com/ruraomsk/ag-server/binding"
)

func Test_SetTimeUseOld(t *testing.T) {
	buffer148 := []int{148, 0, 23, 34, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 30, 0, 0}
	buffer157 := []int{157, 0, 20, 9, 0, 15, 12, 8, 0, 0, 0, 0, 0}
	td := binding.NewSetTimeUse()
	err := td.FromBuffer(buffer148)
	if err != nil {
		t.Error(err.Error())
		return
	}
	err = td.FromBuffer(buffer157)
	if err != nil {
		t.Error(err.Error())
		return
	}
	out148 := td.ToBuffer(148)
	out157 := td.ToBuffer(157)
	if !reflect.DeepEqual(&buffer148, &out148) {
		t.Errorf("No equal\n%v\n%v", buffer148, out148)
		return
	}
	if !reflect.DeepEqual(&buffer157, &out157) {
		t.Errorf("No equal\n%v\n%v", buffer157, out157)
		return
	}
}
func Test_SetTimeUseNew(t *testing.T) {
	buffer148 := []int{148, 0, 23, 58, 0, 138, 2, 6, 0, 0, 0, 9, 1, 40, 9, 1, 40, 9, 1, 40, 9, 1, 40, 9, 2, 40, 9, 2, 40, 9, 2, 40, 9, 4, 35, 9, 4, 30, 9, 8, 40, 9, 8, 30, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0}
	buffer157 := []int{157, 0, 20, 9, 0, 36, 24, 13, 8, 0, 0, 0, 0}
	td := binding.NewSetTimeUse()
	err := td.FromBuffer(buffer148)
	if err != nil {
		t.Error(err.Error())
		return
	}
	err = td.FromBuffer(buffer157)
	if err != nil {
		t.Error(err.Error())
		return
	}
	out148 := td.ToBuffer(148)
	out157 := td.ToBuffer(157)
	if !reflect.DeepEqual(&buffer148, &out148) {
		t.Errorf("No equal\n%v\n%v", buffer148, out148)
		return
	}
	if !reflect.DeepEqual(&buffer157, &out157) {
		t.Errorf("No equal\n%v\n%v", buffer157, out157)
		return
	}
}

func Test_Ctrl(t *testing.T) {
	buffer := []int{149, 0, 24, 35, 1, 0, 0, 5, 30, 0, 30, 7, 0, 40, 15, 9, 0, 15, 5, 16, 0, 40, 5, 19, 0, 0, 15, 21, 0, 0, 0, 24, 0, 0, 0, 0, 0, 255, 255}

	td := binding.NewSetCtrl()
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

	td := binding.NewWeekSets()
	err := td.FromBuffer(buffer)
	if err != nil {
		t.Error(err.Error())
		return
	}
	out := td.WeekSets[0].ToBuffer()
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
	buffer := []int{14, 0, 14, 9, 0, 1, 5, 3, 0, 0, 0, 0, 0}

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
	buffer := []int{15, 0, 15, 17, 0, 7, 1, 8, 1, 5, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}

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

	td := binding.NewSetupDK()
	err := td.FromBuffer(buffer)
	if err != nil {
		t.Error(err.Error())
		return
	}
	out := td.ToBuffer()
	if !reflect.DeepEqual(&buffer, &out) {
		t.Errorf("No equal buffer\n%v\n%v", buffer, out)
		return
	}
}
func setPk(t *testing.T, buffer []int) bool {
	st := binding.NewSetPk(1)
	err := st.FromBuffer(buffer)
	if err != nil {
		t.Error(err.Error())
		return false
	}
	out := st.ToBuffer()
	if !reflect.DeepEqual(&buffer, &out) {
		t.Errorf("No equal\n%v\n%v", buffer, out)
		return false
	}
	return true
}
func Test_SetDK(t *testing.T) {
	setPk(t, []int{100, 0, 133, 30, 1, 75, 31, 61, 195, 128, 1, 30, 2, 50, 3, 75, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0})
	setPk(t, []int{101, 0, 133, 30, 2, 75, 31, 61, 195, 128, 1, 30, 3, 55, 2, 75, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0})
	setPk(t, []int{102, 0, 133, 30, 3, 75, 31, 61, 195, 0, 1, 30, 2, 50, 3, 75, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0})
	setPk(t, []int{103, 0, 133, 30, 4, 75, 31, 61, 196, 16, 3, 10, 1, 40, 2, 60, 3, 75, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0})
	setPk(t, []int{104, 0, 133, 30, 5, 75, 31, 61, 196, 16, 1, 20, 2, 40, 3, 65, 1, 75, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0})
	setPk(t, []int{105, 0, 133, 30, 6, 75, 31, 61, 195, 0, 3, 25, 1, 55, 2, 75, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0})
	setPk(t, []int{106, 0, 133, 30, 7, 180, 76, 16, 201, 16, 5, 12, 1, 42, 2, 62, 3, 82, 4, 102, 5, 122, 1, 152, 3, 172, 5, 180, 0, 0, 0, 0, 0, 0})
	setPk(t, []int{107, 0, 133, 30, 8, 120, 16, 16, 196, 192, 1, 30, 0, 70, 2, 92, 0, 120, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0})
	setPk(t, []int{108, 0, 133, 30, 9, 120, 16, 16, 196, 64, 1, 30, 0, 70, 2, 92, 0, 120, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0})
	setPk(t, []int{109, 0, 133, 30, 10, 120, 16, 16, 197, 80, 0, 8, 1, 38, 0, 78, 2, 100, 0, 120, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0})
	setPk(t, []int{110, 0, 133, 30, 11, 120, 16, 16, 197, 80, 2, 20, 0, 48, 1, 78, 0, 118, 2, 120, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0})
	setPk(t, []int{111, 0, 133, 30, 12, 120, 16, 16, 196, 64, 0, 40, 2, 62, 0, 90, 1, 120, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0})
	setPk(t, []int{104, 0, 133, 30, 5, 180, 76, 16, 196, 2, 1, 90, 2, 134, 3, 165, 4, 180, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0})
	setPk(t, []int{105, 0, 133, 30, 6, 62, 8, 2, 195, 0, 1, 38, 2, 54, 0, 62, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0})
	setPk(t, []int{106, 0, 133, 30, 7, 56, 32, 16, 196, 32, 1, 29, 0, 32, 47, 56, 17, 56, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0})
	setPk(t, []int{107, 0, 133, 30, 8, 54, 40, 34, 198, 56, 1, 16, 228, 40, 130, 36, 67, 34, 21, 30, 1, 54, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0})
}
