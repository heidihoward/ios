// Package store provides a simple key value store
// Not safe for concurreny access
package store

import (
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

func (s *Store) execute(req string) string {
	request := strings.Split(req, " ")

	switch request[0] {
	case "update":
		if len(request) != 3 {
			return "not reconised"
		}
		glog.Infof("Updating %s to %s", request[1], request[2])
		(*s)[request[1]] = request[2]
		return "OK"
	case "get":
		if len(request) != 2 {
			return "not reconised"
		}
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

func (s *Store) Process(req string) string {
	reqs := strings.Split(strings.Trim(req, "\n"), "; ")
	var reply string

	for i := range reqs {
		if i == 0 {
			reply = s.execute(reqs[i])
		} else {
			reply = reply + "; " + s.execute(reqs[i])
		}
	}
	return reply
}

func (s *Store) Print() {
	for key, value := range *s {
		glog.Info("(", key, value, ")")
	}
}
