package main

import (
	"flag"
	"github.com/golang/glog"
	"github.com/heidi-ann/ios/client"
	"github.com/heidi-ann/ios/config"
	"github.com/heidi-ann/ios/test/generator"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var configFile = flag.String("config", os.Getenv("GOPATH")+"/src/github.com/heidi-ann/ios/client/example.conf", "Client configuration file")
var autoFile = flag.String("auto", os.Getenv("GOPATH")+"/src/github.com/heidi-ann/ios/test/workload.conf", "Configure file for workload")
var number = flag.Int("n",1, "Number of clients to create")

// runClient returns when workload is finished or SubmitRequest fails
func runClient(id int ,addresses conf.Addresses.Address, timeout time.Duration, workloadConfig config.WorkloadConfig) {
	c := client.StartClient(id, "latency_"+id+".csv", addresses, timeout)
	ioapi := generator.Generate(config.ParseWorkloadConfig(*autoFile), *consistencyCheck)

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

	// always flush (whatever happens)
	sigs := make(chan os.Signal, 1)
	finish := make(chan bool, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	// parse config files
	conf := config.ParseClientConfig(*configFile)
	timeout := time.Millisecond * time.Duration(conf.Parameters.Timeout)
	workloadConf := config.ParseWorkloadConfig(*autoFile)

	c := client.StartClient(*id, *statFile, conf.Addresses.Address, timeout)

	// setup API
	ioapi := generator.Generate(config.ParseWorkloadConfig(*autoFile), *consistencyCheck)

	go func() {
		for {
			// get next command
			text, read, ok := ioapi.Next()
			if !ok {
				finish <- true
				break
			}
			// pass to ios client
			success, reply := c.SubmitRequest(text, read)
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
