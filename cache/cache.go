// Package cache provides a simple key value store mapping client ID's to the last request sent to them.
// It is safe for concurreny access

// TODO: locking should be more fine grained
package cache

import (
	"github.com/heidi-ann/ios/msgs"
	"sync"
	"github.com/golang/glog"
)

type Cache struct {
	m map[int]msgs.ClientResponse
	sync.RWMutex
}

func Create() *Cache {
	c := map[int]msgs.ClientResponse{}
	return &Cache{c, sync.RWMutex{}}
}

func (c *Cache) Check(req msgs.ClientRequest) (bool, msgs.ClientResponse) {
	c.RLock()
	last := c.m[req.ClientID]
	c.RUnlock()
	if last.RequestID > req.RequestID {
		glog.Fatal("Request has already been applied to state machine and overwritten in request cache")
	}
	return req.RequestID == last.RequestID, last
}

func (c *Cache) Add(res msgs.ClientResponse) {
	c.Lock()
	if c.m[res.ClientID].RequestID != 0 && c.m[res.ClientID].RequestID + 1 != res.RequestID {
		glog.Fatal("Requests must be added to request cache in order")
	}
	c.m[res.ClientID] = res
	c.Unlock()
}
