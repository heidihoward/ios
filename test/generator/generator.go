// Package generator provides test clients for benchmarking Ios's key value store performance.
// Currently, generator only generates get and updates requests.
package generator

import (
	"fmt"
	"github.com/golang/glog"
	"github.com/heidi-ann/ios/config"
	"github.com/heidi-ann/ios/services"
	"math/rand"
	"time"
)

// Generator generates workloads for the key value store application.
type Generator struct {
	config  config.ConfigAuto // workload configuration.
	keys    []string          // key value store keys for the workload to operate on
	store   services.Service  // a local kv store to check consistency of responses
	request string            // current outstanding request
}

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

// randStringBytes n generates a random alphanumeric string of n bytes.
func randStringBytes(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

// Generate creates a workload generator with the specific configuration.
func Generate(config config.WorkloadConfig) *Generator {
	keys := make([]string, config.Config.Keys)
	for i := range keys {
		keys[i] = randStringBytes(config.Config.KeySize)
	}
	return &Generator{config.Config, keys, services.StartService("kv-store"), ""}
}

// Next return the next request in the workload or false if no more are available.
func (g *Generator) Next() (string, bool) {
	//handle termination after n requests
	if g.config.Requests == 0 {
		return "", false
	}
	g.config.Requests--

	delay := 0
	if g.config.Interval > 0 {
		delay = rand.Intn(g.config.Interval)
	}
	time.Sleep(time.Duration(delay) * time.Millisecond)

	// generate key
	key := g.keys[rand.Intn(g.config.Keys)]
	glog.V(1).Info("Key is ", key)

	if rand.Intn(100) < g.config.Reads {
		g.request = fmt.Sprintf("get %s", key)
		return g.request, true
	}
	value := randStringBytes(g.config.ValueSize)
	g.request = fmt.Sprintf("update %s %s", key, value)
	return g.request, true
}

// Return notifies the generator of Ios's response so it can check consistency
func (g *Generator) Return(response string) {
	expected := g.store.Process(g.request)
	if expected != response {
		glog.Fatal("Unexpected response ", response, " expected ", expected)
	}
}
