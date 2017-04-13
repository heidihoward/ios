package main

import (
	"flag"
	"github.com/golang/glog"
	"github.com/heidi-ann/ios/client"
	"github.com/heidi-ann/ios/config"
	"github.com/heidi-ann/ios/test/generator"
	"os"
	"strconv"
)

var configFile = flag.String("config", os.Getenv("GOPATH")+"/src/github.com/heidi-ann/ios/example.conf", "Client configuration file")
var autoFile = flag.String("auto", os.Getenv("GOPATH")+"/src/github.com/heidi-ann/ios/test/workloads/example.conf", "Configure file for workload")
var algorithmFile = flag.String("algorithm", os.Getenv("GOPATH")+"/src/github.com/heidi-ann/ios/configfiles/simple/client.conf", "Algorithm description file") // optional flag
var clients = flag.Int("clients", 1, "Number of clients to create")

// runClient returns when workload is finished or SubmitRequest fails
func runClient(id int, clientConfig config.Config, addresses []config.NetAddress, workloadConfig config.WorkloadConfig) {
	c := client.StartClientFromConfig(-1, "latency_"+strconv.Itoa(id)+".csv", clientConfig, addresses)
	ioapi := generator.Generate(workloadConfig, false)
	hist := openHistoryFile("history_"+strconv.Itoa(id)+".csv")
	for {
		// get next command
		text, read, ok := ioapi.Next()
		if !ok {
			break
		}
		// pass to ios client
		hist.startRequest(text)
		success, reply := c.SubmitRequest(text, read)
		if !success {
			break
		}
		// notify API of result
		hist.stopRequest(reply)
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
	conf := config.ParseClientConfig(*algorithmFile)
	addresses := config.ParseAddresses(*configFile).Clients
	workloadConfig := config.ParseWorkloadConfig(*autoFile)

	remaining := *clients
	for id := 0; id < *clients; id++ {
		go func(id int) {
			runClient(id, conf, addresses, workloadConfig)
			remaining--
			if remaining == 0 {
				finished <- true
			}
		}(id)
	}

	// wait for workload to finish
	<-finished
	glog.Info("Client set terminating")
	glog.Flush()

}
