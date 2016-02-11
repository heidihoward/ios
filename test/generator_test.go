package test

import "testing"

func TestGenerate(t *testing.T) {
	_ = Generate(50, 3)

	// STUB
	for i := 0; i < 10; i++ {
		if i > 10 {
			t.Errorf("failure")
		}
	}

}
