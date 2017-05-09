package main

import (
	"encoding/csv"
	"flag"
	"os"
	"time"
	"strconv"

	"github.com/golang/glog"
	"github.com/heidi-ann/ios/client"
	"github.com/heidi-ann/ios/config"
	"github.com/heidi-ann/ios/test/generator"
)

var configFile = flag.String("config", os.Getenv("GOPATH")+"/src/github.com/heidi-ann/ios/example.conf", "Client configuration file")
var autoFile = flag.String("auto", os.Getenv("GOPATH")+"/src/github.com/heidi-ann/ios/test/workloads/example.conf", "Configure file for workload")
var algorithmFile = flag.String("algorithm", os.Getenv("GOPATH")+"/src/github.com/heidi-ann/ios/configfiles/simple/client.conf", "Algorithm description file") // optional flag
var resultsFile = flag.String("results", "results.csv", "File to write results to")
var clientsMin = flag.Int("clients-min", 1, "Min number of clients to create (inclusive)")
var clientsMax = flag.Int("clients-max", 10, "Max number of clients to create (inclusive)")
var clientsStep = flag.Int("clients-step", 1, "Step in number of clients to create")

// runClient returns when workload is finished or SubmitRequest fails
func runClient(id int, clientConfig config.Config, addresses []config.NetAddress, workload config.ConfigAuto) (requests int, bytes int) {
	requests = 0
	bytes = 0
	c, err := client.StartClientFromConfig(-1, "", clientConfig, addresses)
	if err != nil {
		glog.Fatal(err)
	}
	ioapi := generator.Generate(workload, false)
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
		reply, err := c.SubmitRequest(text, read)
		if err != nil {
			break
		}
		// notify API of result
		ioapi.Return(reply)
		requests++
		bytes += len(text)
	}
	c.StopClient()
	return
}

func main() {
	// set up logging
	flag.Parse()
	defer glog.Flush()

	// set up results file
	file, err := os.OpenFile(*resultsFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0777)
	if err != nil {
		glog.Fatal(err)
	}
	writer := csv.NewWriter(file)
	writer.Write([]string{
		"clients",
		"total time [ms]",
		"throughput [req/sec]",
		"throughput [Kbps]",
	})

	// parse config files (instead of requiring each client to do it)
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

	for clients := *clientsMin; clients <= *clientsMax; clients += *clientsStep {
		startTime := time.Now()
		requestsCompleted := 0
		bytesCommitted := 0
		remaining := clients
		for id := 0; id < clients; id++ {
			go func(id int) {
				requests, bytes := runClient(id, conf, addresses.Clients, workloadConfig)
				requestsCompleted += requests
				bytesCommitted += bytes
				remaining--
				if remaining == 0 {
					finished <- true
				}
			}(id)
		}

		// wait for workload to finish
		<-finished
		totalTime := time.Since(startTime).Seconds() // time in secs
		requestThroughput :=  float64(requestsCompleted) / totalTime // throughput in req/sec
		byteThroughput := float64(8*bytesCommitted/1000) / totalTime // throughput in Kbps
		writer.Write([]string{
			strconv.Itoa(clients),
			strconv.FormatFloat(totalTime*1000000, 'f', 0, 64),
			strconv.FormatFloat(requestThroughput, 'f', 0, 64),
			strconv.FormatFloat(byteThroughput, 'f', 0, 64),
		})
		writer.Flush()
		glog.Info("Client set terminating after completing ", requestsCompleted," requests at ",requestThroughput," [reqs/sec]")
	}

	// finish up
	glog.Flush()
	writer.Flush()
}
