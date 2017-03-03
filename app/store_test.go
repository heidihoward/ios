package app

import (
	"reflect"
	"testing"
)

func Testprocess(t *testing.T) {
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

func TestrestoreSnapshot(t *testing.T) {
	var store Store
	store = map[string]string{
		"A": "0",
		"B": "0",
		"C": "0",
	}
	snapshot := store.MakeSnapshot()
	restore := RestoreSnapshot(snapshot)
	if !reflect.DeepEqual(store, *restore) {
		t.Error("orginal store and restored store are not the same", store, *restore)
	}
}
