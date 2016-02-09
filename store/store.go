package store

import (
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
	request := strings.Split(req, " ")

	switch request[0] {
	case "update":
		(*s)[request[1]] = request[2]
		return "OK"
	case "get":
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
