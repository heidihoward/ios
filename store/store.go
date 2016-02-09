package store

import (
	"fmt"
	"github.com/golang/glog"
	"strings"
)

type Store map[string]string

func New() *Store {
	var s Store
	s = map[string]string{
		"A": "0",
		"B": "0",
		"C": "0",
	}
	return &s
}

func (s *Store) Process(req string) string {
	request := strings.Split(strings.Trim(req, "\n"), " ")

	switch request[0] {
	case "update":
		glog.Infof("Updating %s to %s", request[1], request[2])
		(*s)[request[1]] = request[2]
		return "OK"
	case "get":
		glog.Infof("Getting %s", request[1])
		value, ok := (*s)[request[1]]
		if ok {
			return value
		} else {
			return "key not found"
		}
	default:
		return "not reconised"
	}
}

func (s *Store) Print() {
	for key, value := range *s {
		fmt.Println("(", key, value, ")")
	}
}
