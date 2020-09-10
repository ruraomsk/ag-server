package main

import (
	"strings"
	"testing"
	"time"
)

func Test_DataString(t *testing.T) {
	s := time.Now().Format("2006-01-02")
	if strings.Compare(s, "2020-09-02") != 0 {
		t.Errorf("Date=%s", s)
	}

}
