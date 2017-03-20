package services

import (
	"strconv"
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

func TestMarshalling(t *testing.T) {
	store := StartService("kv-store")
	// add data to store
	store.Process("update A 1")
	for i := 0; i < 100; i++ {
		store.Process(fmt.Sprintf("update %v %v", i, i))
	}
	store.Process("delete 0")

	//marshal store
	json, err := store.MarshalJSON()
	if err != nil {
		t.Fatal("err marshalling store. Error: ", err.Error())
	}
	//make more changes (should be overwritten when unmarshalling)
	store.Process("delete 1")

	//unmarshal and validate store
	err = store.UnmarshalJSON(json)
	if err != nil {
		t.Fatal("err unmarshalling store. Error: ", err.Error())
	}

	assert.Equal(t, "1", store.Process("get A"))
	assert.Equal(t, "false", store.Process("exists 0"))
	assert.Equal(t, "100", store.Process("count"))
	for i := 1; i < 100; i++ {
		assert.Equal(t, strconv.Itoa(i), store.Process(fmt.Sprintf("get %v", i)))
	}

}
