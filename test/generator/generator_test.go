package generator

import (
	"github.com/heidi-ann/ios/config"
	"github.com/heidi-ann/ios/services"
	"strings"
	"testing"
)

func checkFormatSize(t *testing.T, req string, keySize int, valueSize int) {
	request := strings.Split(strings.Trim(req, "\n"), " ")

	// check key size
	if len(request[1]) != keySize {
		t.Errorf("Incorrect key length: '%s'", req)
	}
	// check value size
	if request[0] == "update" && len(request[2]) != valueSize {
		t.Errorf("Incorrect key length: '%s'", req)
	}
}

func checkGenerateConfig(t *testing.T, conf config.WorkloadConfig) {
	kv := services.StartService("kv-store")
	gen := Generate(conf)

	for i := 0; i < conf.Config.Requests; i++ {
		str, _, ok := gen.Next()
		// check for early termination
		if !ok {
			if conf.Config.Requests != i {
				t.Errorf("Generator terminated a request %d, should terminate at %d",
					i, conf.Config.Requests)
			}
			break
		}
		// check format
		checkFormatSize(t, str, conf.Config.KeySize, conf.Config.ValueSize)
		if !kv.CheckFormat(str) {
			t.Errorf("incorrect format for kv store")
		}
	}
	// check for late termination
	_, _, ok := gen.Next()
	if ok {
		t.Errorf("Generator not terminated at %d", conf.Config.Requests+1)
	}

}

// TestGenerates check that the generator is producing valid key value store commands
func TestGenerates(t *testing.T) {
	conf := config.WorkloadConfig{config.ConfigAuto{
		Reads:     50,
		Interval:  0,
		KeySize:   8,
		ValueSize: 8,
		Requests:  20,
		Keys:      1,
	}}
	checkGenerateConfig(t, conf)

}
