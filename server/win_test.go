package server

import (
	"os"
	"testing"
)

func TestUnPackUECrash(t *testing.T) {
	data, _ := os.ReadFile("H:\\ZPlan\\zplan-client\\Saved\\Crashes\\UECC-Windows-78DDD6704B9249D64B197989C57EF369_0000.pak")
	if err := unPackUECrash(data, "H:\\ZPlan\\zplan-client\\Saved\\Crashes\\Go"); err != nil {
		t.Errorf(err.Error())
	}
}
