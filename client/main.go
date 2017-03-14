// Package client provides I/O for Ios clients
package main

import (
	"flag"
	"github.com/golang/glog"
	"github.com/heidi-ann/ios/api/interactive"
	"github.com/heidi-ann/ios/api/rest"
	"github.com/heidi-ann/ios/config"
	"github.com/heidi-ann/ios/test"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type API interface {
	Next() (string, bool)
	Return(string)
}

var configFile = flag.String("config", os.Getenv("GOPATH")+"/src/github.com/heidi-ann/ios/client/example.conf", "Client configuration file")
var autoFile = flag.String("auto", os.Getenv("GOPATH")+"/src/github.com/heidi-ann/ios/test/workload.conf", "If workload is automatically generated, configure file for workload")
var statFile = flag.String("stat", "latency.csv", "File to write stats to")
var mode = flag.String("mode", "interactive", "interactive, rest or test")
var id = flag.Int("id", -1, "ID of client (must be unique) or random number will be generated")

func main() {
	// set up logging
	flag.Parse()
	defer glog.Flush()

	// always flush (whatever happens)
	sigs := make(chan os.Signal, 1)
	finish := make(chan bool, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	// parse config files
	conf := config.ParseClientConfig(*configFile)
	timeout := time.Millisecond * time.Duration(conf.Parameters.Timeout)
	// TODO: find a better way to handle required flags
	if *id == -1 {
		rand.Seed(time.Now().UTC().UnixNano())
		*id = rand.Int()
		glog.V(1).Info("ID was not provided, ID ", *id, " has been assigned")
	}

	c := StartClient(*id, *statFile, conf.Addresses.Address, timeout)

	// setup API
	var ioapi API
	switch *mode {
	case "interactive":
		ioapi = interactive.Create(conf.Parameters.Application)
	case "test":
		ioapi = test.Generate(config.ParseWorkloadConfig(*autoFile))
	case "rest":
		ioapi = rest.Create()
	default:
		glog.Fatal("Invalid mode: ", mode)
	}

	go func() {
		for {
			// get next command
			text, ok := ioapi.Next()
			if !ok {
				finish <- true
				break
			}
			// pass to ios client
			success, reply := c.SubmitRequest(text)
			if !success {
				finish <- true
				break
			}
			// notify API of result
			ioapi.Return(reply)

		}
	}()

	select {
	case sig := <-sigs:
		glog.Warning("Termination due to: ", sig)
	case <-finish:
		glog.Info("No more commands")
	}
	c.StopClient()
	glog.Flush()

}
