package services

import (
	"encoding/json"
)

type dummy struct {
  requests int
}

func newDummy() *dummy {
	return &dummy{0}
}

func (d *dummy) Process(request string) string {
  if request=="ping"{
    d.requests++
    return "pong"
  }
  return ""
}

func (d *dummy) MarshalJSON() ([]byte, error) {
	return json.Marshal(*d)
}

func (d *dummy) UnmarshalJSON(_ []byte) error {
  // TODO: finish placeholder
	return nil
}
