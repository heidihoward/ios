// Package server is the main entry point to run an Ios server on Unix

package main

import (
	"flag"
	"github.com/golang/glog"
	"github.com/heidi-ann/ios/config"
	"github.com/heidi-ann/ios/consensus"
	"github.com/heidi-ann/ios/msgs"
	"github.com/heidi-ann/ios/unix"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
)

// command line flags
var id = flag.Int("id", -1, "server ID [REQUIRED]")                                                                                           // required flag
var config_file = flag.String("config", os.Getenv("GOPATH")+"/src/github.com/heidi-ann/ios/server/example.conf", "Server configuration file") // optional flag
var diskPath = flag.String("disk", ".", "Path to directory to store persistent storage")                                                      // optional flag

// main entry point of server
func main() {
	// set up logging
	flag.Parse()
	defer glog.Flush()

	// parse configuration file
	conf := config.ParseServerConfig(*config_file)
	if *id == -1 {
		glog.Fatal("ID is required")
	}
	if *id >= len(conf.Peers.Address) {
		glog.Fatal("Node ID is ", *id, " but is configured with a ", len(conf.Peers.Address), " node cluster")
	}

	// logging
	glog.Info("Starting server ", *id)
	defer glog.Warning("Shutting down server ", *id)

	// setup iO
	// TODO: remove this hardcoded limit on channel size
	iO := msgs.MakeIo(2000, len(conf.Peers.Address))

	// setup persistent storage
	logFile := *diskPath + "/persistent_log_" + strconv.Itoa(*id) + ".temp"
	dataFile := *diskPath + "/persistent_data_" + strconv.Itoa(*id) + ".temp"
	snapFile := *diskPath + "/persistent_snapshot_" + strconv.Itoa(*id) + ".temp"
	found, view, log, index, state := unix.SetupStorage(
		logFile, dataFile, snapFile, iO, conf.Options.Length,
		conf.Unsafe.DumpPersistentStorage, conf.Unsafe.PersistenceMode, conf.Options.Application)

	// setup peers & clients
	failureDetector := msgs.NewFailureNotifier(len(conf.Peers.Address))
	unix.SetupPeers(*id, conf.Peers.Address, iO, failureDetector)
	unix.SetupClients(strings.Split(conf.Clients.Address[*id], ":")[1], state)

	// configure consensus algorithms
	quorum := consensus.NewQuorum(conf.Options.QuorumSystem, len(conf.Peers.Address))
	configuration := consensus.Config{
		*id,
		len(conf.Peers.Address),
		conf.Options.Length,
		conf.Options.BatchInterval,
		conf.Options.MaxBatch,
		conf.Options.DelegateReplication,
		conf.Options.WindowSize,
		conf.Options.SnapshotInterval,
		quorum,
		conf.Options.IndexExclusivity}

	// setup consensus algorithm
	if !found {
		glog.Info("Starting fresh consensus instance")
		go consensus.Init(iO, configuration, state, failureDetector)
	} else {
		glog.Info("Restoring consensus instance")
		go consensus.Recover(iO, configuration, view, log, state, index, failureDetector)
	}

	// tidy up
	glog.Info("Setup complete")

	// waiting for exit
	// always flush (whatever happens)
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	sig := <-sigs
	glog.Flush()
	glog.Warning("Shutting down due to ", sig)
}
