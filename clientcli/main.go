package main

import (
	"flag"
	"github.com/golang/glog"
	"github.com/heidi-ann/ios/client"
	"github.com/heidi-ann/ios/clientcli/cli"
	"github.com/heidi-ann/ios/config"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var configFile = flag.String("config", os.Getenv("GOPATH")+"/src/github.com/heidi-ann/ios/client/example.conf", "Client configuration file")
var statFile = flag.String("stat", "latency.csv", "File to write stats to")
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

	c := client.StartClient(*id, *statFile, conf.Addresses.Address, timeout)

	// setup API
	ioapi := cli.Create(conf.Parameters.Application)

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
