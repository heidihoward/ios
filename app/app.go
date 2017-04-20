package app

import (
	"encoding/json"
	"github.com/golang/glog"
	"github.com/heidi-ann/ios/msgs"
	"github.com/heidi-ann/ios/services"
)

// StateMachine abstracts over the services state machine and the cache which ensure exactly-once execution
type StateMachine struct {
	Cache *Cache
	Store services.Service
}

// New creates a StateMachine with the given service application
func New(appConfig string) *StateMachine {
	return &StateMachine{newCache(), services.StartService(appConfig)}
}

// Apply request will apply a request (or fetch the result of its application) and return the result
func (s *StateMachine) Apply(req msgs.ClientRequest) msgs.ClientResponse {
	glog.V(1).Info("Request has been safely replicated by consensus algorithm", req)

	// check if request already applied
	if found, reply := s.Cache.check(req); found {
		glog.V(1).Info("Request found in cache and thus need not be applied", req)
		return reply
	}
	// apply request and cache
	reply := msgs.ClientResponse{
		req.ClientID, req.RequestID, true, s.Store.Process(req.Request)}
	s.Cache.add(reply)
	return reply
}

// ApplyRead request will apply a read request and return the result. It will not cache the result
func (s *StateMachine) ApplyRead(req msgs.ClientRequest) msgs.ClientResponse {
	glog.V(1).Info("Read request has been passed by consensus algorithm", req)
	return msgs.ClientResponse{
		req.ClientID, req.RequestID, true, s.Store.Process(req.Request)}
}

// ApplyReads request will apply a slice of read requests and return the results. It will not cache the results.
func (s *StateMachine) ApplyReads(reqs []msgs.ClientRequest) []msgs.ClientResponse {
	glog.V(1).Info("Read requests has been passed to state machine by consensus algorithm")
	responses := make([]msgs.ClientResponse, len(reqs))
	for i := 0; i < len(reqs); i++ {
		responses[i] = msgs.ClientResponse{
			reqs[i].ClientID, reqs[i].RequestID, true, s.Store.Process(reqs[i].Request)}
	}
	return responses
}

// Check request return true and the result of the request if the request has already been applied to the state machine
func (s *StateMachine) Check(req msgs.ClientRequest) (bool, msgs.ClientResponse) {
	return s.Cache.check(req)
}

// MakeSnapshot serializes a state machine into bytes
func (s *StateMachine) MakeSnapshot() ([]byte, error) {
	return json.Marshal(s)
}

// RestoreSnapshot deserializes bytes into a state machine
func RestoreSnapshot(snap []byte, appConfig string) (*StateMachine, error) {
	sm := New(appConfig)
	err := json.Unmarshal(snap, sm)
	return sm, err
}
