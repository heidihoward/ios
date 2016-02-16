package cache

import (
	"github.com/heidi-ann/hydra/msgs"
)

type Cache map[int]msgs.ClientResponse

func Create() *Cache {
	var c Cache
	c = map[int]msgs.ClientResponse{}
	return &c
}

func (c *Cache) Check(req msgs.ClientRequest) (bool, msgs.ClientResponse) {
	last := (*c)[req.ClientID]
	return req.RequestID == last.RequestID, last
}

func (c *Cache) Add(res msgs.ClientResponse) {
	(*c)[res.ClientID] = res
}
