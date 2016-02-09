package store

import "testing"

func TestProcess(t *testing.T) {
	store := New()

	cases := []struct {
		req, res string
	}{
		{"update A 3", "OK"},
		{"get A", "3"},
	}

	for _, c := range cases {
		got := store.Process(c.req)
		if got != c.res {
			t.Errorf("%s returned %s but %s was expected", c.req, got, c.res)
		}
	}
}
