// Package main is the entry point to directly run an Ios server as an executable
package main

import (
	"flag"
	"github.com/golang/glog"
	"github.com/heidi-ann/ios/config"
	"github.com/heidi-ann/ios/ios/server"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
)

// command line flags
var id = flag.Int("id", -1, "server ID [REQUIRED]")                                                                                       // required flag
var configFile = flag.String("config", os.Getenv("GOPATH")+"/src/github.com/heidi-ann/ios/iso/example.conf", "Server configuration file") // optional flag
var algorithmFile = flag.String("algorithm", os.Getenv("GOPATH")+"/src/github.com/heidi-ann/ios/configfiles/simple/server.conf", "Algorithm description file") // optional flag
var diskPath = flag.String("disk", "persistent_id", "Path to directory to store persistent storage")                                      // optional flag
var peerPort = flag.Int("listen-peers", 0, "Overwrite the port specified in config file to listen for peers on")                          // optional flag
var clientPort = flag.Int("listen-clients", 0, "Overwrite the port specified in config file to listen for clients on")                    // optional flag

// entry point of server executable
func main() {
	// set up logging
	flag.Parse()
	defer glog.Flush()

	// check go path is set
	if os.Getenv("GOPATH") == "" {
		glog.Fatal("GOPATH not set, please set GOPATH and try again")
	}

	// parse configuration files
	conf := config.ParseServerConfig(*algorithmFile)
	addresses := config.ParseAddresses(*configFile)
	if *id == -1 {
		glog.Fatal("ID is required")
	}
	// add ID to diskPath
	disk := *diskPath
	if disk == "persistent_id" {
		disk = "persistent_id" + strconv.Itoa(*id)
	}

	// overwrite ports if given
	if *peerPort != 0 {
		glog.Info("Peer port overwritten to ", *peerPort)
		ip := strings.Split(addresses.Peers.Address[*id], ":")[0]
		addresses.Peers.Address[*id] = ip + ":" + strconv.Itoa(*peerPort)
	}

	if *clientPort != 0 {
		glog.Info("Client port overwritten to ", *clientPort)
		ip := strings.Split(addresses.Clients.Address[*id], ":")[0]
		addresses.Clients.Address[*id] = ip + ":" + strconv.Itoa(*clientPort)
	}

	// logging
	glog.Info("Starting server ", *id)
	defer glog.Warning("Shutting down server ", *id)

	// start Ios server
	go server.RunIos(*id, conf, addresses, disk)

	// waiting for exit
	// always flush (whatever happens)
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	sig := <-sigs
	glog.Flush()
	glog.Warning("Shutting down due to ", sig)
}
