// Package server provides I/O for Ios servers

package main

import (
	"bufio"
	"flag"
	"github.com/golang/glog"
	"github.com/heidi-ann/ios/app"
	"github.com/heidi-ann/ios/config"
	"github.com/heidi-ann/ios/consensus"
	"github.com/heidi-ann/ios/msgs"
	"github.com/heidi-ann/ios/unix"
	"io"
	"net"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
)

var application *app.StateMachine
var cons_io *msgs.Io

var notifyClients *unix.Notificator

var id = flag.Int("id", -1, "server ID")
var config_file = flag.String("config", "example.conf", "Server configuration file")
var disk_path = flag.String("disk", ".", "Path to directory to store persistent storage")

func stateMachine() {
	for {
		var req msgs.ClientRequest
		var reply msgs.ClientResponse

		select {
		case response := <-cons_io.OutgoingResponses:
			req = response.Request
			reply = response.Response
		case req = <- cons_io.OutgoingRequestsFailed:
			glog.Info("Request could not been safely replicated by consensus algorithm", req)
			reply = msgs.ClientResponse{
				req.ClientID, req.RequestID, false, ""}
		}

		// if any handleRequests are waiting on this reply, then reply to them
		notifyClients.Notify(req,reply)
	}
}

func handleRequest(req msgs.ClientRequest) msgs.ClientResponse {
	glog.Info("Handling ", req.Request)

	// check if already applied
	if found, res := application.Check(req); found {
		glog.Info("Request found in cache")
		return res // FAST PASS
	}

	// CONSENESUS ALGORITHM HERE
	glog.Info("Passing request to consensus algorithm")
	if req.ForceViewChange {
			cons_io.IncomingRequestsForced <- req
	} else {
			cons_io.IncomingRequests <- req
	}

	if notifyClients.IsSubscribed(req) {
		glog.Warning("Client has multiple outstanding connections for the same request, usually not a good sign")
	}

	// wait for reply
	reply := notifyClients.Subscribe(req)

	// check reply is as expected
	if reply.ClientID != req.ClientID {
		glog.Fatal("ClientID is different")
	}
	if reply.RequestID != req.RequestID {
		glog.Fatal("RequestID is different")
	}

	return reply
}

func handleConnection(cn net.Conn) {
	glog.Info("Incoming client connection from ",
		cn.RemoteAddr().String())

	reader := bufio.NewReader(cn)
	writer := bufio.NewWriter(cn)

	for {

		// read request
		glog.Info("Ready for Reading")
		text, err := reader.ReadBytes(byte('\n'))
		if err != nil {
			if err == io.EOF {
				break
			}
			glog.Warning(err)
			break
		}
		glog.Info("--------------------New request----------------------")
		glog.Info("Request: ", string(text))
		req := new(msgs.ClientRequest)
		err = msgs.Unmarshal(text, req)
		if err != nil {
			glog.Fatal(err)
		}

		// construct reply
		reply := handleRequest(*req)
		b, err := msgs.Marshal(reply)
		if err != nil {
			glog.Fatal("error:", err)
		}
		glog.Info(string(b))

		// send reply
		glog.Info("Sending ", string(b))
		n, err := writer.Write(b)
		if err != nil {
			glog.Fatal(err)
		}
		_, err = writer.Write([]byte("\n"))
		if err != nil {
			glog.Fatal(err)
		}

		// tidy up
		err = writer.Flush()
		glog.Info("Finished sending ", n, " bytes")

	}

	cn.Close()
}

func main() {
	// set up logging
	flag.Parse()
	defer glog.Flush()

	conf := config.ParseServerConfig(*config_file)
	if *id == -1 {
		glog.Fatal("ID is required")
	}
	if *id >= len(conf.Peers.Address) {
		glog.Fatal("Node ID is ",*id," but is configured with a ",len(conf.Peers.Address)," node cluster")
	}

	glog.Info("Starting server ", *id)
	defer glog.Warning("Shutting down server ", *id)

	// setup IO
	cons_io = msgs.MakeIo(2000, len(conf.Peers.Address))

	notifyClients = unix.NewNotificator()
	go stateMachine()

	// set up persistent storage
	logFile := *disk_path + "/persistent_log_" + strconv.Itoa(*id) + ".temp"
	dataFile := *disk_path + "/persistent_data_" + strconv.Itoa(*id) + ".temp"
	snapFile := *disk_path + "/persistent_snapshot_" + strconv.Itoa(*id) + ".temp"
	found, view, log, index, state := unix.SetupPersistentStorage(logFile, dataFile, snapFile, cons_io, conf.Options.Length)
	application = state

	// set up client server
	glog.Info("Starting up client server")
	client_port := strings.Split(conf.Clients.Address[*id],":")[1]
	listeningPort := ":" + client_port
	ln, err := net.Listen("tcp", listeningPort)
	if err != nil {
		glog.Fatal(err)
	}

	// setup peers
	unix.SetupPeers(*id, conf.Clients.Address, cons_io)

	// handle for incoming clients
	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				glog.Fatal(err)
			}
			go handleConnection(conn)
		}
	}()

	// setting up the consensus algorithm
	log_max_length := 1000
	if conf.Options.Length > 0 {
		log_max_length = conf.Options.Length
	}
	quorum := consensus.NewQuorum(conf.Options.QuorumSystem,len(conf.Peers.Address))
	cons_config := consensus.Config{*id, len(conf.Peers.Address),
		log_max_length, conf.Options.BatchInterval, conf.Options.MaxBatch, conf.Options.DelegateReplication, conf.Options.WindowSize,  conf.Options.SnapshotInterval, quorum}
	if !found {
		glog.Info("Starting fresh consensus instance")
		go consensus.Init(cons_io, cons_config, state)
	} else {
		glog.Info("Restoring consensus instance")
		go consensus.Recover(cons_io, cons_config, view, log, state, index)
	}
	//go cons_io.DumpPersistentStorage()

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
