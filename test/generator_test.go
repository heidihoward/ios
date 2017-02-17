package test

import (
	"strings"
	"testing"
)

func checkKey(t *testing.T, str string, key string) {
	if key != str {
		t.Errorf("Invalid key: '%s' '%s'", str, key)
	}
	if len(str)!= 8 {
		t.Errorf("Wrong key length")
	}
}

func checkValue(t *testing.T, str string) {
	if len(str) != 8 {
		t.Errorf("Invalid value: '%s'", str)
	}
}

func checkFormat(t *testing.T, req string, key string) {
	request := strings.Split(strings.Trim(req, "\n"), " ")

	switch request[0] {
	case "update":
		if len(request) != 3 {
			t.Errorf("Misformatted update request: '%s'", req)
		}
		checkKey(t, request[1],key)
		checkValue(t, request[2])
	case "get":
		if len(request) != 2 {
			t.Errorf("Misformatted get request: '%s'", req)
		}
		checkKey(t, request[1],key)
	default:
		t.Errorf("Request is neither get or update: '%s'", req)
	}
}

// check that the generator is producing valid commands
func TestGenerate(t *testing.T) {
	conf := WorkloadConfig{ConfigAuto{50, 0, 8 ,8, 20, 1}}

	gen := Generate(conf)
	key := ""
	for i := 0; i < 100; i++ {
		str, _, ok := gen.Next()
		if !ok {
			if conf.Config.Requests != i {
				t.Errorf("Generator terminated a request %d, should terminate at %d",
					i, conf.Config.Requests)
			}
			break
		}
		if i==0 {
			key=strings.Split(strings.Trim(str, "\n"), " ")[1]
		}
		checkFormat(t, str, key)
	}

}
