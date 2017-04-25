// Package kvclient provides a client for interacting with a key-value type Ios cluster
package kvclient

import (
	"strconv"
	"time"

	"fmt"

	"github.com/heidi-ann/ios/client"
	"github.com/heidi-ann/ios/config"
)

type KvClient struct {
	iosClient *client.Client
}

// StartKvClient creates an Ios client and tries to connect to an Ios cluster
// If ID is -1 then a random one will be generated
func StartKvClient(id int, statFile string, addrs []config.NetAddress, timeout time.Duration, backoff time.Duration, beforeForce int, random bool) (*KvClient, error) {
	iosClient, err := client.StartClient(id, statFile, addrs, timeout, backoff, beforeForce, random)
	return &KvClient{iosClient}, err
}

func (kvc *KvClient) Update(key string, value string) error {
	_, err := kvc.iosClient.SubmitRequest(fmt.Sprintf("update %v %v", key, value), false)
	return err
}

func (kvc *KvClient) Get(key string) (string, error) {
	return kvc.iosClient.SubmitRequest(fmt.Sprintf("get %v", key), true)
}

func (kvc *KvClient) Delete(key string) error {
	_, err := kvc.iosClient.SubmitRequest(fmt.Sprintf("delete %v", key), false)
	return err
}

func (kvc *KvClient) Count() (int, error) {
	reply, err := kvc.iosClient.SubmitRequest("count", true)
	if err != nil {
		return 0, err
	}
	count, err := strconv.Atoi(reply)
	return count, err
}

func (kvc *KvClient) Print() (string, error) {
	return kvc.iosClient.SubmitRequest("print", true)
}
