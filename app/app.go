package app

import (
	"encoding/json"
	"github.com/golang/glog"
	"github.com/heidi-ann/ios/msgs"
)

type StateMachine struct {
	Cache *Cache
	Store *Store
}

func New() *StateMachine {
	return &StateMachine{newCache(), newStore()}
}

func (s *StateMachine) Apply(req msgs.ClientRequest) msgs.ClientResponse {
	glog.V(1).Info("Request has been safely replicated by consensus algorithm", req)

	// check if request already applied
	if found, reply := s.Cache.check(req); found {
		glog.V(1).Info("Request found in cache and thus need not be applied", req)
		return reply
	}
	// apply request and cache
	reply := msgs.ClientResponse{
		req.ClientID, req.RequestID, true, s.Store.process(req.Request)}
	s.Cache.add(reply)
	return reply
}

func (s *StateMachine) Check(req msgs.ClientRequest) (bool, msgs.ClientResponse) {
	return s.Cache.check(req)
}

func (s *StateMachine) MakeSnapshot() []byte {
	b, err := json.Marshal(s)
	if err != nil {
		glog.Fatal("Unable to snapshot state machine: ", err)
	}
	return b
}

func RestoreSnapshot(snap []byte) *StateMachine {
	var sm StateMachine
	err := json.Unmarshal(snap, &sm)
	if err != nil {
		glog.Fatal("Unable to restore from snapshot: ", err)
	}
	return &sm
}
