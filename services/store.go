// Package services provides a simple key value store
// Not safe for concurreny access
// TODO: replace map with https://github.com/orcaman/concurrent-map
package services

import (
	"encoding/json"
	"github.com/golang/glog"
	"strconv"
	"strings"
)

type store map[string]string

func newStore() *store {
	var s store
	s = map[string]string{}
	return &s
}

func (s *store) execute(req string) string {
	request := strings.Split(req, " ")

	switch request[0] {
	case "update":
		if len(request) != 3 {
			return "not recognised"
		}
		glog.V(1).Infof("Updating %s to %s", request[1], request[2])
		(*s)[request[1]] = request[2]
		return "OK"
	case "exists":
		if len(request) != 2 {
			return "request not recognised"
		}
		glog.V(1).Infof("Checking if %s exists", request[1])
		_, exists := (*s)[request[1]]
		return strconv.FormatBool(exists)
	case "get":
		if len(request) != 2 {
			return "request not recognised"
		}
		glog.V(1).Infof("Getting %s", request[1])
		value, ok := (*s)[request[1]]
		if ok {
			return value
		} else {
			return "key not found"
		}
	case "delete":
		if len(request) != 2 {
			return "request not recognised"
		}
		glog.V(1).Infof("Deleting %s", request[1])
		delete(*s, request[1])
		return "OK"
	case "count":
		if len(request) != 1 {
			return "request not recognised"
		}
		glog.V(1).Infof("Counting size of key-value store")
		return strconv.Itoa(len(*s))
	case "print":
		if len(request) != 1 {
			return "request not recognised"
		}
		glog.V(1).Infof("Printing key-value store")
		return s.print()
	default:
		return "request not recognised"
	}
}

func (s *store) Process(req string) string {
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

func (s *store) CheckFormat(req string) bool {
	request := strings.Split(strings.Trim(req, "\n"), " ")
	switch request[0] {
	case "update":
		return len(request) == 3
	case "exists", "get", "delete":
		return len(request) == 2
	case "count", "print":
		return len(request) == 1
	default:
		return false
	}
}

func (s *store) print() string {
	str := ""
	for key, value := range *s {
		str += key + ", " + value + "\n"
	}
	return str
}

func (s *store) MarshalJSON() ([]byte, error) {
	return json.Marshal(*s)
}

func (s *store) UnmarshalJSON(snap []byte) error {
	// this seems like a strange approach but unmarshalling directly into store causes memory leak
	var sTemp map[string]string
	json.Unmarshal(snap, &sTemp)
	for key, value := range sTemp {
		(*s)[key] = value
	}
	return nil
}
