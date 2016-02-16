package msgs

import (
	"encoding/json"
)

type ClientRequest struct {
	ClientID  int
	RequestID int
	Request   string
}

type ClientResponse struct {
	ClientID  int
	RequestID int
	Response  string
}

// abstract the fact that JSON is used for comms

func Marshal(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

func Unmarshal(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}
