package test

import (
	"strconv"
	"strings"
	"testing"
)

func checkKey(t *testing.T, str string) {
	key, err := strconv.Atoi(str)
	if err != nil {
		t.Errorf("Misformatted key: '%s'", str)
	}
	if !(key >= 0 && key <= 9) {
		t.Errorf("Invalid key: '%s'", str)
	}
}

func checkValue(t *testing.T, str string) {
	val, err := strconv.Atoi(str)
	if err != nil {
		t.Errorf("Misformatted value: '%s'", str)
	}
	if val != 7 {
		t.Errorf("Invalid value: '%s'", str)
	}
}

func checkFormat(t *testing.T, req string) {
	request := strings.Split(strings.Trim(req, "\n"), " ")

	switch request[0] {
	case "update":
		if len(request) != 3 {
			t.Errorf("Misformatted update request: '%s'", req)
		}
		checkKey(t, request[1])
		checkValue(t, request[2])
	case "get":
		if len(request) != 2 {
			t.Errorf("Misformatted get request: '%s'", req)
		}
		checkKey(t, request[1])
	default:
		t.Errorf("Request is neither get or update: '%s'", req)
	}
}

// check that the generator is producing valid commands
func TestGenerate(t *testing.T) {
	conf := ConfigAuto{
		Commands{50, 3},
		Termination{20},
	}

	gen := Generate(conf)

	for i := 0; i < 100; i++ {
		str, ok := gen.Next()
		if !ok {
			if conf.Termination.Requests != i {
				t.Errorf("Generator terminated a request %d, should terminate at %d'",
					i, conf.Termination.Requests)
			}
			break
		}
		checkFormat(t, str)
	}

}
