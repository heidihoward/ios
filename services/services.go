package services

import (
	"github.com/golang/glog"
)

type Service interface {
	Process(req string) string
	CheckFormat(req string) bool
	MarshalJSON() ([]byte, error)
	UnmarshalJSON(snap []byte) error
}

func StartService(config string) Service {
	var serv Service
	switch config {
	case "kv-store":
		serv = newStore()
	case "dummy":
		serv = newDummy()
	default:
		glog.Fatal("No valid service specified")
	}
	return serv
}

func GetInteractiveText(config string) string {
	var s string
	switch config {
	case "kv-store":
		s =
			`The following commands are available:
	get [key]: to return the value of a given key
	exists [key]: to test if a given key is present
	update [key] [value]: to set the value of a given key, if key already exists then overwrite
	delete [key]: to remove a key value pair if present
	count: to return the number of keys
	print: to return all key value pairs
`
	case "dummy":
		s =
			`The following commands are available:
		ping: ping dummy application
`
	}
	return s
}
