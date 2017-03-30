package main

import (
	"flag"
	"github.com/golang/glog"
	"github.com/heidi-ann/ios/client"
	"github.com/heidi-ann/ios/config"
	"github.com/heidi-ann/ios/test/generator"
	"os"
	"strconv"
	"time"
)

var configFile = flag.String("config", os.Getenv("GOPATH")+"/src/github.com/heidi-ann/ios/client/example.conf", "Client configuration file")
var autoFile = flag.String("auto", os.Getenv("GOPATH")+"/src/github.com/heidi-ann/ios/test/workload.conf", "Configure file for workload")
var clients = flag.Int("clients",1, "Number of clients to create")

// runClient returns when workload is finished or SubmitRequest fails
func runClient(id int, addresses []string, timeout time.Duration, workloadConfig config.WorkloadConfig) {
	c := client.StartClient(-1, "latency_"+strconv.Itoa(id)+".csv", addresses, timeout)
	ioapi := generator.Generate(workloadConfig, false)

	for {
		// get next command
		text, read, ok := ioapi.Next()
		if !ok {
			break
		}
		// pass to ios client
		success, reply := c.SubmitRequest(text, read)
		if !success {
			break
		}
		// notify API of result
		ioapi.Return(reply)
	}
	c.StopClient()
}

func main() {
	// set up logging
	flag.Parse()
	defer glog.Flush()

	// parse config files
	finished := make(chan bool)
	conf := config.ParseClientConfig(*configFile)
	timeout := time.Millisecond * time.Duration(conf.Parameters.Timeout)
	workloadConfig := config.ParseWorkloadConfig(*autoFile)

  remaining := *clients
	for id := 0; id < *clients; id++ {
		go func(id int) {
			runClient(id, conf.Addresses.Address, timeout, workloadConfig)
			remaining--
			if remaining == 0 {
				finished <- true
			}
		}(id)
	}

	// wait for workload to finish
	<- finished
	glog.Info("Client set terminating")
	glog.Flush()

}
