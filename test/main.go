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
)

var configFile = flag.String("config", os.Getenv("GOPATH")+"/src/github.com/heidi-ann/ios/example.conf", "Client configuration file")
var autoFile = flag.String("auto", os.Getenv("GOPATH")+"/src/github.com/heidi-ann/ios/test/workloads/example.conf", "Configure file for workload")
var algorithmFile = flag.String("algorithm", os.Getenv("GOPATH")+"/src/github.com/heidi-ann/ios/configfiles/simple/client.conf", "Algorithm description file") // optional flag
var statFile = flag.String("stat", "latency.csv", "File to write stats to")
var consistencyCheck = flag.Bool("check", false, "Enable consistency checking (use with only one client)")
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
	c, err := client.StartClientFromConfigFile(*id, *statFile, *algorithmFile, *configFile)
	if err != nil {
		glog.Fatal(err)
	}

	// setup API
	workload, err := config.ParseWorkloadConfig(*autoFile)
	if err != nil {
		glog.Fatal(err)
	}
	if err = config.CheckWorkloadConfig(workload); err != nil {
		glog.Fatal(err)
	}
	ioapi := generator.Generate(workload, *consistencyCheck)

	go func() {
		for {
			// get next command
			text, read, ok := ioapi.Next()
			if !ok {
				finish <- true
				break
			}
			// pass to ios client
			reply, err := c.SubmitRequest(text, read)
			if err != nil {
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
