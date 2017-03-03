// Package server provides I/O for Ios servers

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

var id = flag.Int("id", -1, "server ID")
var config_file = flag.String("config", "example.conf", "Server configuration file")
var disk_path = flag.String("disk", ".", "Path to directory to store persistent storage")

func main() {
	// set up logging
	flag.Parse()
	defer glog.Flush()
	glog.Info("Starting server ", *id)
	defer glog.Warning("Shutting down server ", *id)

	// parse configuration
	conf := config.ParseServerConfig(*config_file)
	if *id == -1 {
		glog.Fatal("ID is required")
	}
	if *id >= len(conf.Peers.Address) {
		glog.Fatal("Node ID is ", *id, " but is configured with a ", len(conf.Peers.Address), " node cluster")
	}

	// setup IO
	IO := msgs.MakeIo(2000, len(conf.Peers.Address))

	// setup persistent storage
	logFile := *disk_path + "/persistent_log_" + strconv.Itoa(*id) + ".temp"
	dataFile := *disk_path + "/persistent_data_" + strconv.Itoa(*id) + ".temp"
	snapFile := *disk_path + "/persistent_snapshot_" + strconv.Itoa(*id) + ".temp"
	found, view, log, index, state := unix.SetupPersistentStorage(logFile, dataFile, snapFile, IO, conf.Options.Length)

	// setup peers & clients
	unix.SetupPeers(*id, conf.Peers.Address, IO)
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
		quorum}

	// setup consensus algorithm
	if !found {
		glog.Info("Starting fresh consensus instance")
		go consensus.Init(IO, configuration, state)
	} else {
		glog.Info("Restoring consensus instance")
		go consensus.Recover(IO, configuration, view, log, state, index)
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
