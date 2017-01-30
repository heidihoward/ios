// Package cache provides a simple key value store mapping client ID's to the last request sent to them.
// It is safe for concurreny access

// TODO: locking should be more fine grained
package cache

import (
	"github.com/heidi-ann/ios/msgs"
	"sync"
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
	return req.RequestID == last.RequestID, last
}

func (c *Cache) Add(res msgs.ClientResponse) {
	c.Lock()
	c.m[res.ClientID] = res
	c.Unlock()
}
