// Package kvclient provides a client for interacting with a key-value type Ios cluster
package kvclient

import (
	"strconv"
	"time"

	"fmt"

	"github.com/heidi-ann/ios/client"
)

type KvClient struct {
	iosClient *client.Client
}

// StartKvClient creates an Ios client and tries to connect to an Ios cluster
// If ID is -1 then a random one will be generated
func StartKvClient(id int, statFile string, addrs []string, timeout time.Duration) *KvClient {
	iosClient := client.StartClient(id, statFile, addrs, timeout)
	return &KvClient{iosClient}
}

func (kvc *KvClient) Update(key string, value string) {
	kvc.iosClient.SubmitRequest(fmt.Sprintf("update %v %v", key, value))
}

func (kvc *KvClient) Get(key string) string {
	_, reply := kvc.iosClient.SubmitRequest(fmt.Sprintf("get %v", key))
	return reply
}

func (kvc *KvClient) Delete(key string) {
	kvc.iosClient.SubmitRequest(fmt.Sprintf("delete %v", key))
}

func (kvc *KvClient) Count() int {
	_, reply := kvc.iosClient.SubmitRequest("count")
	count, _ := strconv.Atoi(reply)
	return count
}

func (kvc *KvClient) Print() string {
	_, reply := kvc.iosClient.SubmitRequest("print")
	return reply
}
