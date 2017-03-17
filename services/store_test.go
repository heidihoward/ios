package services

import (
	"testing"

	"fmt"

	"github.com/stretchr/testify/assert"
)

func TestProcess(t *testing.T) {
	store := newStore()

	cases := []struct {
		req, res string
	}{
		{"update A 3", "OK"},
		{"get A", "3"},
		{"print", "A, 3\n"},
		{"count", "1"},
		{"exists A", "true"},
		{"delete A", "OK"},
		{"count", "0"},
		{"exists A", "false"},
		{"get B", "key not found"},
	}

	for _, c := range cases {
		got := store.Process(c.req)
		if got != c.res {
			t.Errorf("%s returned %s but %s was expected", c.req, got, c.res)
		}
	}
}

func TestCheckFormat(t *testing.T) {
	store := StartService("kv-store")

	cases := []struct {
		req string
		res bool
	}{
		{"update foo bar", true},
		{"get ;234", true},
		{"print", true},
		{"count", true},
		{"exists $£$%", true},
		{"delete _)(*", true},
		{"get B F", false},
		{"", false},
	}
	for _, c := range cases {
		got := store.CheckFormat(c.req)
		if got != c.res {
			assert.Equal(t, c.res, got, fmt.Sprintf("Error for case: %s", c.req))
		}
	}

	dummy := StartService("dummy")
	assert.True(t, dummy.CheckFormat("ping"))
	assert.False(t, dummy.CheckFormat(""))
	assert.False(t, dummy.CheckFormat("update A 1"))
}
