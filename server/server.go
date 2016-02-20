package main

import (
	"bufio"
	"flag"
	"github.com/golang/glog"
	"github.com/heidi-ann/hydra/cache"
	"github.com/heidi-ann/hydra/config"
	"github.com/heidi-ann/hydra/consensus"
	"github.com/heidi-ann/hydra/store"
	"github.com/heidi-ann/hydra/msgs"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

var keyval *store.Store
var disk *bufio.Writer
var c *cache.Cache
var cons_io *msgs.Io

type Peer struct {
	id       int
	address  string
	handled  bool // TOOD: replace with Mutex
}

var peers []Peer

var client_port = flag.Int("client-port", 8080, "port to listen on for clients")
var peer_port = flag.Int("peer-port", 8090, "port to listen on for peers")
var id = flag.Int("id", -1, "server ID")
var config_file = flag.String("config", "example.conf", "Server configuration file")

func handleRequest(req msgs.ClientRequest) msgs.ClientResponse {
	glog.Info("Handling ", req.Request)

	// check if already applied
	found, res := c.Check(req)
	if found {
		glog.Info("Request found in cache")
		return res // FAST PASS
	}

	// write to persistent storage
	n, err := disk.WriteString(req.Request)
	_ = disk.Flush()
	if err != nil {
		glog.Fatal(err)
	}
	glog.Infof("Written %b bytes to persistent storage", n)

	// CONSENESUS ALGORITHM HERE
	glog.Info("Passing request to consensus algorithm")
	(*cons_io).IncomingRequests <- req
	_ = <-(*cons_io).OutgoingRequests
	glog.Info("Request has been safely replicated by consensus algorithm")

	// check if request already applied
	found, res = c.Check(req)
	if found {
		glog.Info("Request found in cache and thus cannot be applied")
		return res
	}

	// apply request
	output := keyval.Process(req.Request)
	keyval.Print()

	// write response to request cache
	reply := msgs.ClientResponse{
		req.ClientID, req.RequestID, output}

	c.Add(reply)
	return reply
}

// iterative through peers and check there is a handler for each
// try to create one if not
func checkPeer() {
	for i := range peers {
		if !peers[i].handled {
			glog.Info("Peer ", i, " is not currently connected")
			cn, err := net.Dial("tcp", peers[i].address)

			if err != nil {
				glog.Warning(err)
				break
			}

			handlePeer(cn, true)
		}
	}
}

func handlePeer(cn net.Conn, _ bool) {
	addr := cn.RemoteAddr().String()
	glog.Info("Peer connection from ", addr)

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

	peers[peer_id].handled = true

	close_err := make(chan error)
	go func() {
		for {
			// read request
			glog.Info("Reading from peer ", peer_id)
			text, err := reader.ReadBytes(byte('\n'))
			if err != nil && err != io.EOF {
				close_err <- err
				break
			}

			(*cons_io).Incoming.BytesToProtoMsg(text)

		}
	}()

	go func() {
		for {
			// send reply
			// TODO: URGENT FIX NEEDED
			proto_msg := (*cons_io).OutgoingUnicast[peer_id]
			b, _ := proto_msg.ProtoMsgToBytes()
			glog.Info("Sending ", string(b))
			_, err := writer.Write(b)
			_, err = writer.Write([]byte("\n"))
			if err != nil {
				close_err <- err
				break
			}
			err = writer.Flush()
		}
	}()

	// block until connection fails
	err = <-close_err
	glog.Warning(err)

	// tidy up
	glog.Infof("No longer about to handle traffic from peer %d at %s ", id, addr)
	peers[peer_id].handled = false
	cn.Close()
}

func handleConnection(cn net.Conn) {
	glog.Info("Incoming client connection from ",
		cn.RemoteAddr().String())

	reader := bufio.NewReader(cn)
	writer := bufio.NewWriter(cn)

	for {

		// read request
		glog.Info("Reading")
		text, err := reader.ReadBytes(byte('\n'))
		if err != nil {
			if err == io.EOF {
				break
			}
			glog.Fatal(err)
		}
		glog.Info(string(text))
		req := new(msgs.ClientRequest)
		msgs.Unmarshal(text, req)

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

	//set up state machine
	keyval = store.New()
	c = cache.Create()

	// setting up persistent log
	filename := "persistent_" + strconv.Itoa(*id) + ".log"
	glog.Info("Opening file: ", filename)
	file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0777)
	if err != nil {
		glog.Fatal(err)
	}
	disk = bufio.NewWriter(file)
	defer disk.Flush()

	// check persistent storage for commands
	disk_reader := bufio.NewReader(file)
	for {
		str, err := disk_reader.ReadString('\n')
		if err != nil {
			glog.Info("No more commands in persistent storage")
			break
		}
		_ = keyval.Process(str)
	}

	// set up client server
	glog.Info("Starting up client server")
	listeningPort := ":" + strconv.Itoa(*client_port)
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
	cons_io = msgs.MakeIo(10,len(conf.Peers.Address))

	//set up peer server
	glog.Info("Starting up peer server")
	listeningPort = ":" + strconv.Itoa(*peer_port)
	lnPeers, err := net.Listen("tcp", listeningPort)
	if err != nil {
		glog.Fatal(err)
	}

	// handle local peer (without sending network traffic)
	peers[*id].handled = true
	from := &((*cons_io).Incoming)
	go from.Forward((*cons_io).OutgoingUnicast[*id])

	// regularly check if all peers are connected and reply if not
	go func() {
		for {
			checkPeer()
			time.Sleep(10 * time.Second)
		}
	}()

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

	// setting up the consensus algorithm
	cons_config := consensus.Config{*id, len(conf.Peers.Address)}
	consensus.Init(cons_io,cons_config)

	// tidy up
	time.Sleep(30 * time.Second)
	glog.Info("Shutting down")
}
