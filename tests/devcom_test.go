package main

import (
	"testing"
	"time"

	"github.com/ruraomsk/ag-server/pudge"
	"github.com/ruraomsk/ag-server/transport"
)

var sb transport.SubMessage

func Test_0x00D(t *testing.T) {
	sb.Set0x00Device()
	if sb.Message[0] != 0 {
		t.Error("Ошибка ")
	}
}
func Test_0x01D(t *testing.T) {
	sb.Set0x01Device(1, 2, 3, 4, 5)
	a1, a2, a3, a4, a5 := sb.Get0x01Device()
	if a1 != 1 || a2 != 2 || a3 != 3 || a4 != 4 || a5 != 5 {
		t.Errorf("Ошибка %d %d %d %d %d", a1, a2, a3, a4, a5)

	}
}
func Test_0x04D(t *testing.T) {
	sb.Set0x04Device(10, 20, 30, 40)
	a1, a2, a3, a4 := sb.Get0x04Device()
	if a1 != 10 || a2 != 20 || a3 != 30 || a4 != 40 {
		t.Errorf("Ошибка %d %d %d %d ", a1, a2, a3, a4)

	}
}
func Test_0x07D(t *testing.T) {
	sb.Set0x07Device(10, 20, 30, 40)
	a1, a2, a3, a4 := sb.Get0x07Device()
	if a1 != 10 || a2 != 20 || a3 != 30 || a4 != 40 {
		t.Errorf("Ошибка %d %d %d %d ", a1, a2, a3, a4)

	}
}
func Test_0x09D(t *testing.T) {
	var st pudge.Statistic
	st.Period = 10
	st.Type = 1
	st.TLen = 60
	st.Hour = 10
	st.Min = 20
	st.Datas = make([]pudge.DataStat, 0)
	sb.Set0x09Device(&st)
	var stt pudge.Statistic
	stt, err := sb.Get0x09Device()
	if err != nil {
		t.Error(err.Error())
	}
	if !st.Compare(&stt) {
		t.Errorf("Не совпали \n%v\n%v\n ", st, stt)
	}
}
func Test_0x0AD(t *testing.T) {
	var st pudge.Statistic
	st.Period = 10
	st.Type = 1
	st.TLen = 60
	st.Hour = 10
	st.Min = 20
	st.Datas = make([]pudge.DataStat, 0)
	var sd pudge.DataStat
	for i := 1; i < 11; i++ {
		sd.Chanel = i
		sd.Status = i % 3
		sd.Intensiv = i*sd.Status + 21
		st.Datas = append(st.Datas, sd)
	}
	err := sb.Set0x0ADevice(&st)
	if err != nil {
		t.Error(err.Error())
	}
	var stt pudge.Statistic
	stt.Period = 10
	stt.Type = 1
	stt.TLen = 60
	stt.Hour = 10
	stt.Min = 20
	stt.Datas = make([]pudge.DataStat, 0)
	err = sb.Get0x0ADevice(&stt)
	if err != nil {
		t.Error(err.Error())
	}
	if !st.Compare(&stt) {
		t.Errorf("Не совпали \n%v\n%v\n ", st, stt)
	}
}
func Test_0x0BD(t *testing.T) {
	var lg pudge.LogLine
	lg.Record = "1"
	lg.Time = time.Now()
	lg.Info = "120"
	sb.Set0x0BDevice(&lg)
	var lgg pudge.LogLine
	err := sb.Get0x0BDevice(&lgg)
	if err != nil {
		t.Error(err.Error())
	}
	if !lg.Compare(&lgg) {
		t.Errorf("Не совпали \n%v\n%v\n ", lg, lgg)
	}
}
func Test_0x13D(t *testing.T) {
	var ar pudge.ArrayPriv
	ar.Number = 12
	ar.Array = make([]int, 0)
	for i := 1; i < 11; i++ {
		ar.Array = append(ar.Array, i)
	}
	sb.Set0x13Device(&ar)
	var arr pudge.ArrayPriv
	err := sb.Get0x13Device(&arr)
	if err != nil {
		t.Error(err.Error())
	}
	if !ar.Compare(&arr) {
		t.Errorf("Не совпали \n%v\n%v\n ", ar, arr)
	}
}
func Test_0x1CD(t *testing.T) {
	sb.Set0x1CDevice(111)
	a1 := sb.Get0x1CDevice()
	if a1 != 111 {
		t.Errorf("Ошибка %d ", a1)

	}
}
func Test_0x0FD(t *testing.T) {
	c := new(pudge.Controller)
	c.PK = 1
	c.NK = 2
	c.CK = 3
	c.DK.FDK = 9
	sb.Set0x0FDevice(c)
	cc := new(pudge.Controller)
	sb.Get0x0FDevice(cc)
	if !c.Compare(cc) {
		t.Errorf("Не равно \n%v \n%v", c, cc)
	}
}
