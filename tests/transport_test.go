package main

import (
	"github.com/ruraomsk/ag-server/pudge"
	"github.com/ruraomsk/ag-server/transport"
	"testing"
	"time"
)

func Test_ParseServer(t *testing.T) {
	var hs transport.HeaderServer
	hs.IDServer = 0xa78d
	hs.Time = time.Now()
	hs.Code = 0x7f
	hs.Number = 128
	var ms transport.SubMessage
	mss := make([]transport.SubMessage, 0)
	ms.Set0x01Server(10)
	mss = append(mss, ms)
	ms.Set0x02Server(true)
	mss = append(mss, ms)
	ms.Set0x03Server()
	mss = append(mss, ms)
	ms.Set0x04Server(true, true)
	mss = append(mss, ms)
	ms.Set0x05Server(11)
	mss = append(mss, ms)
	ms.Set0x06Server(3)
	mss = append(mss, ms)
	ms.Set0x07Server(12)
	mss = append(mss, ms)
	ms.Set0x09Server(13)
	mss = append(mss, ms)
	ms.Set0x0AServer(14)
	mss = append(mss, ms)
	hs.UpackMessages(mss)
	buffer := hs.MakeBuffer()
	var nhs transport.HeaderServer
	err := nhs.Parse(buffer)
	if err != nil {
		t.Error(err.Error())
	}
	if !hs.Compare(&nhs) {
		t.Errorf("Не совпали HeaderServer \n%v \n%v\n", hs, nhs)
	}
	smess := hs.ParseMessage()
	nsmess := nhs.ParseMessage()
	for n, mes := range smess {
		nmes := nsmess[n]
		if !nmes.Compare(&mes) {
			t.Error(nmes.ToString(), " not equal ", mes.ToString())
		}
	}

}
func Test_EmptyDevice(t *testing.T) {
	var hd transport.HeaderDevice
	hd.ID = 25000
	hd.TypeDevice = 30
	hd.Time = time.Now()
	hd.Code = 0x7f
	hd.Number = 128
	// var ms transport.SubMessage
	mss := make([]transport.SubMessage, 0)
	hd.UpackMessages(mss)
	buffer := hd.MakeBuffer()
	var nhd transport.HeaderDevice
	err := nhd.Parse(buffer)
	if err != nil {
		t.Error(err.Error())
	}
	if !hd.Compare(&nhd) {
		t.Errorf("Не совпали \n%v\n%v\n ", hd, nhd)
	}

}
func Test_ParseDevice(t *testing.T) {
	var hd transport.HeaderDevice
	hd.ID = 25000
	hd.TypeDevice = 30
	hd.Time = time.Now()
	hd.Code = 0x7f
	hd.Number = 133
	var ms transport.SubMessage
	mss := make([]transport.SubMessage, 0)
	ms.Set0x00Device()
	mss = append(mss, ms)
	ms.Set0x01Device(1, 2, 3, 4, 5)
	mss = append(mss, ms)
	ms.Set0x04Device(10, 11, 9, 12)
	mss = append(mss, ms)
	ms.Set0x07Device(7, 8, 9, 10)
	mss = append(mss, ms)
	var st pudge.Statistic
	ms.Set0x09Device(&st)
	mss = append(mss, ms)
	ms.Set0x0ADevice(&st)
	mss = append(mss, ms)
	var c pudge.Controller
	ms.Set0x0FDevice(&c)
	mss = append(mss, ms)
	ms.Set0x10Device(&c)
	mss = append(mss, ms)
	ms.Set0x11Device(&c)
	mss = append(mss, ms)
	ms.Set0x1DDevice(&c)
	mss = append(mss, ms)
	ms.Set0x12Device(&c)
	mss = append(mss, ms)
	var ar pudge.ArrayPriv
	ar.Array = make([]int, 0)
	ms.Set0x13Device(&ar)
	mss = append(mss, ms)

	hd.UpackMessages(mss)
	buffer := hd.MakeBuffer()
	var nhd transport.HeaderDevice
	err := nhd.Parse(buffer)
	if err != nil {
		t.Error(err.Error())
	}
	if !hd.Compare(&nhd) {
		t.Errorf("Не совпали \n%v\n%v\n ", hd, nhd)
	}
	smess := hd.ParseMessage()
	nsmess := nhd.ParseMessage()
	for n, mes := range smess {
		nmes := nsmess[n]
		if !nmes.Compare(&mes) {
			t.Error(nmes.ToString(), " not equal ", mes.ToString())
		}
	}
}
func Test_MakeSet(t *testing.T) {
	var cc pudge.Controller
	cc.TexRezim = 11
	cc.Base = false
	cc.PK = 3
	cc.CK = 4
	cc.NK = 5
	cc.StatusCommandDU.IsPK = true
	cc.StatusCommandDU.IsDUDK2 = true
	cc.DK1.RDK = 3
	cc.DK1.FDK = 2
	cc.DK1.DDK = 1
	cc.DK1.EDK = 4
	cc.DK1.PDK = true
	cc.DK1.EEDK = 5
	cc.DK1.ODK = true
	cc.DK1.LDK = 6
	cc.DK1.FTUDK = 11
	cc.DK1.TDK = 15
	cc.DK1.FTSDK = 16
	cc.DK1.TTCDK = 17
	cc.DK2.RDK = 5
	cc.DK2.FDK = 4
	cc.DK2.DDK = 3
	cc.DK2.EDK = 2
	cc.DK2.PDK = true
	cc.DK2.EEDK = 6
	cc.DK2.ODK = true
	cc.DK2.LDK = 5
	cc.DK2.FTUDK = 10
	cc.DK2.TDK = 11
	cc.DK2.FTSDK = 12
	cc.DK2.TTCDK = 13
	var ms transport.SubMessage
	err := ms.Set0x0FDevice(&cc)
	if err != nil {
		t.Error(err.Error())
	}
	var ncc pudge.Controller
	err = ms.Get0x0FDevice(&ncc)
	if err != nil {
		t.Error(err.Error())
	}
	if !cc.Compare(&ncc) {
		t.Errorf("Не совпали \n%v\n%v\n ", cc, ncc)
	}

}
