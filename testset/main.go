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
func runClient(id int, clientConfig config.Config, addresses []config.NetAddress, workload config.ConfigAuto) {
	c, err := client.StartClientFromConfig(-1, "latency_"+strconv.Itoa(id)+".csv", clientConfig, addresses)
	if err != nil {
		glog.Fatal(err)
	}
	ioapi := generator.Generate(workload, false)
	hist, err := openHistoryFile("history_" + strconv.Itoa(id) + ".csv")
	if err != nil {
		glog.Fatal(err)
	}
	for {
		// get next command
		text, read, ok := ioapi.Next()
		if !ok {
			break
		}
		// pass to ios client
		hist.startRequest(text)
		reply, err := c.SubmitRequest(text, read)
		if err != nil {
			break
		}
		// notify API of result
		err = hist.stopRequest(reply)
		if err != nil {
			glog.Fatal(err)
		}
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
	conf, err := config.ParseClientConfig(*algorithmFile)
	if err != nil {
		glog.Fatal(err)
	}
	addresses, err := config.ParseAddresses(*configFile)
	if err != nil {
		glog.Fatal(err)
	}
	workloadConfig, err := config.ParseWorkloadConfig(*autoFile)
	if err != nil {
		glog.Fatal(err)
	}

	remaining := *clients
	for id := 0; id < *clients; id++ {
		go func(id int) {
			runClient(id, conf, addresses.Clients, workloadConfig)
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
