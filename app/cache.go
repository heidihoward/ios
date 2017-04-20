package app

import (
	"encoding/json"
	"github.com/golang/glog"
	"github.com/heidi-ann/ios/msgs"
	"strconv"
	"sync"
)

// TODO: locking could be more fine grained for improved concurreny

// Cache provides a simple key value store mapping client ID's to the last request sent to them.
// It is safe for concurreny access
type Cache struct {
	m map[int]msgs.ClientResponse
	sync.RWMutex
}

func newCache() *Cache {
	c := map[int]msgs.ClientResponse{}
	return &Cache{c, sync.RWMutex{}}
}

func (c *Cache) check(req msgs.ClientRequest) (bool, msgs.ClientResponse) {
	c.RLock()
	last := c.m[req.ClientID]
	c.RUnlock()
	if last.RequestID > req.RequestID {
		glog.Fatal("Request has already been applied to state machine and overwritten in request cache")
	}
	return req.RequestID == last.RequestID, last
}

func (c *Cache) add(res msgs.ClientResponse) {
	c.Lock()
	if c.m[res.ClientID].RequestID != 0 && c.m[res.ClientID].RequestID > res.RequestID {
		glog.Fatal("Requests must be added to request cache in order, expected ", c.m[res.ClientID].RequestID+1,
			" but received ", res.RequestID)
	}
	c.m[res.ClientID] = res
	c.Unlock()
}

// MarshalJSON marshals a cache into bytes
// default JSON marshalling requires string map keys thus custom function is provided
func (c *Cache) MarshalJSON() ([]byte, error) {
	c.Lock()
	// convert to string map
	strMap := map[string]msgs.ClientResponse{}
	for k, v := range c.m {
		strMap[strconv.Itoa(k)] = v
	}
	b, err := json.Marshal(strMap)
	if err != nil {
		glog.Warning("Unable to snapshot request cache: ", err)
	}
	c.Unlock()
	return b, err
}

// UnmarshalJSON unmarshals bytes into a cache
func (c *Cache) UnmarshalJSON(snap []byte) error {
	var strMap map[string]msgs.ClientResponse
	err := json.Unmarshal(snap, &strMap)
	if err != nil {
		glog.Warning("Unable to restore from snapshot: ", err)
		return err
	}
	// convert to int map
	c.m = map[int]msgs.ClientResponse{}
	for k, v := range strMap {
		i, err := strconv.Atoi(k)
		if err != nil {
			glog.Warning("Unable to restore from snapshot: ", err)
			return err
		}
		c.m[i] = v
	}
	return nil
}
