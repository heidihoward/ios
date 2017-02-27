// Package server provides I/O for Ios servers

package main

import (
	"bufio"
	"flag"
	"github.com/golang/glog"
	"github.com/heidi-ann/ios/cache"
	"github.com/heidi-ann/ios/config"
	"github.com/heidi-ann/ios/consensus"
	"github.com/heidi-ann/ios/msgs"
	"github.com/heidi-ann/ios/unix"
	"github.com/heidi-ann/ios/store"
	"io"
	"net"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

var keyval *store.Store
var c *cache.Cache
var cons_io *msgs.Io

var notifyclient map[msgs.ClientRequest](chan msgs.ClientResponse)
var notifyclient_mutex sync.RWMutex

type Peer struct {
	id      int
	address string
	handled bool // TODO: replace with Mutex
}

var peers []Peer
var peers_mutex sync.RWMutex

var id = flag.Int("id", -1, "server ID")
var config_file = flag.String("config", "example.conf", "Server configuration file")
var disk_path = flag.String("disk", ".", "Path to directory to store persistent storage")

func stateMachine() {
	for {
		var req msgs.ClientRequest
		var reply msgs.ClientResponse

		select {
		case req = <-cons_io.OutgoingRequests:
			glog.Info("Request has been safely replicated by consensus algorithm", req)

			// check if request already applied
			var found bool
			found, reply = c.Check(req)
			if found {
				glog.Info("Request found in cache and thus cannot be applied")
			} else {
				// apply request
				output := keyval.Process(req.Request)
				//keyval.Print()

				// write response to request cache
				reply = msgs.ClientResponse{
					req.ClientID, req.RequestID, true, output}
				c.Add(reply)
			}
		case req = <- cons_io.OutgoingRequestsFailed:
			glog.Info("Request could not been safely replicated by consensus algorithm", req)
			reply = msgs.ClientResponse{
				req.ClientID, req.RequestID, false, ""}
		}


		// if any handleRequests are waiting on this reply, then reply to them
		notifyclient_mutex.Lock()
		if notifyclient[req] != nil {
			notifyclient[req] <- reply
		}
		notifyclient_mutex.Unlock()
	}
}

func handleRequest(req msgs.ClientRequest) msgs.ClientResponse {
	glog.Info("Handling ", req.Request)

	// check if already applied
	found, res := c.Check(req)
	if found {
		glog.Info("Request found in cache")
		return res // FAST PASS
	}

	// check is request is already in progress
	if notifyclient[req] == nil {
		// CONSENESUS ALGORITHM HERE
		glog.Info("Passing request to consensus algorithm")
		if req.ForceViewChange {
				cons_io.IncomingRequestsForced <- req
		} else {
				cons_io.IncomingRequests <- req
		}

		// wait for reply
		notifyclient_mutex.Lock()
		notifyclient[req] = make(chan msgs.ClientResponse)
		notifyclient_mutex.Unlock()
	}


	reply := <-notifyclient[req]

	// check reply
	if reply.ClientID != req.ClientID {
		glog.Fatal("ClientID is different")
	}
	if reply.RequestID != req.RequestID {
		glog.Fatal("RequestID is different")
	}

	return reply
}

// iterative through peers and check there is a handler for each
// try to create one if not
func checkPeer() {
	for i := range peers {
		peers_mutex.RLock()
		failed := !peers[i].handled
		peers_mutex.RUnlock()
		if failed {
			//glog.Info("Peer ", i, " is not currently connected")
			cn, err := net.Dial("tcp", peers[i].address)

			if err != nil {
				//glog.Warning(err)
			} else {
				go handlePeer(cn, true)
			}
		} else {
			//glog.Info("Peer ", i, " is currently connected")
		}
	}
}

func handlePeer(cn net.Conn, init bool) {
	addr := cn.RemoteAddr().String()
	if init {
		glog.Info("Outgoing peer connection to ", addr)
	} else {
		glog.Info("Incoming peer connection from ", addr)
	}

	defer glog.Warningf("Connection closed from %s ", addr)

	// handle requests
	reader := bufio.NewReader(cn)
	writer := bufio.NewWriter(cn)

	// exchange peer ID's
	_, _ = writer.WriteString(strconv.Itoa(*id) + "\n")
	_ = writer.Flush()
	text, _ := reader.ReadString('\n')
	glog.Info("Received ", text)
	peer_id, err := strconv.Atoi(strings.Trim(text, "\n"))
	if err != nil {
		glog.Warning(err)
		return
	}

	glog.Infof("Ready to handle traffic from peer %d at %s ", peer_id, addr)

	peers_mutex.Lock()
	peers[peer_id].handled = true
	peers_mutex.Unlock()

	close_err := make(chan error)
	go func() {
		for {
			// read request
			glog.Infof("Ready for next message from %d", peer_id)
			text, err := reader.ReadBytes(byte('\n'))
			if err != nil {
				glog.Warning(err)
				close_err <- err
				break
			}
			glog.Infof("Read from peer %d: ", peer_id, string(text))
			cons_io.Incoming.BytesToProtoMsg(text)

		}
	}()

	go func() {
		for {
			// send reply
			glog.Infof("Ready to send message to %d", peer_id)
			b, err := cons_io.OutgoingUnicast[peer_id].ProtoMsgToBytes()
			if err != nil {
				glog.Fatal("Could not marshal message")
			}
			glog.Infof("Sending to %d: %s", peer_id, string(b))
			_, err = writer.Write(b)
			_, err = writer.Write([]byte("\n"))
			if err != nil {
				glog.Warning(err)
				close_err <- err
				break
			}
			// TODO: BUG need to retry packet
			err = writer.Flush()
			if err != nil {
				glog.Warning(err)
				close_err <- err
				break
			}
			glog.Info("Sent")
		}
	}()

	// block until connection fails
	<-close_err

	// tidy up
	glog.Warningf("No longer able to handle traffic from peer %d at %s ", peer_id, addr)
	peers_mutex.Lock()
	peers[peer_id].handled = false
	peers_mutex.Unlock()
	cons_io.Failure <- peer_id
	cn.Close()
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
		// TODO: FIX currently all server send back replies
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

	glog.Info("Starting server ", *id)
	defer glog.Warning("Shutting down server ", *id)

	//set up state machine
	c = cache.Create()
	// setup IO
	cons_io = msgs.MakeIo(2000, len(conf.Peers.Address))

	notifyclient = make(map[msgs.ClientRequest](chan msgs.ClientResponse))
	notifyclient_mutex = sync.RWMutex{}
	go stateMachine()

	// set up persistent storage
	logFile := *disk_path + "/persistent_log_" + strconv.Itoa(*id) + ".temp"
	dataFile := *disk_path + "/persistent_data_" + strconv.Itoa(*id) + ".temp"
	snapFile := *disk_path + "/persistent_snapshot_" + strconv.Itoa(*id) + ".temp"
	found, view, log, index, snapshot := unix.SetupPersistentStorage(logFile, dataFile, snapFile, cons_io, conf.Options.Length)
	keyval = snapshot

	// set up client server
	glog.Info("Starting up client server")
	client_port := strings.Split(conf.Clients.Address[*id],":")[1]
	listeningPort := ":" + client_port
	ln, err := net.Listen("tcp", listeningPort)
	if err != nil {
		glog.Fatal(err)
	}

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

	//set up peer state
	peers = make([]Peer, len(conf.Peers.Address))
	for i := range conf.Peers.Address {
		peers[i] = Peer{
			i, conf.Peers.Address[i], false}
	}
	peers_mutex = sync.RWMutex{}

	//set up peer server
	glog.Info("Starting up peer server")
	peer_port := strings.Split(conf.Peers.Address[*id],":")[1]
	listeningPort = ":" + peer_port
	lnPeers, err := net.Listen("tcp", listeningPort)
	if err != nil {
		glog.Fatal(err)
	}

	// handle local peer (without sending network traffic)
	peers_mutex.Lock()
	peers[*id].handled = true
	peers_mutex.Unlock()
	from := &(cons_io.Incoming)
	go from.Forward(cons_io.OutgoingUnicast[*id])

	// handle for incoming peers
	go func() {
		for {
			conn, err := lnPeers.Accept()
			if err != nil {
				glog.Fatal(err)
			}
			go handlePeer(conn, false)
		}
	}()

	// regularly check if all peers are connected and retry if not
	go func() {
		for {
			checkPeer()
			time.Sleep(500 * time.Millisecond)
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
		go consensus.Init(cons_io, cons_config, keyval)
	} else {
		glog.Info("Restoring consensus instance")
		go consensus.Recover(cons_io, cons_config, view, log, keyval, index)
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
