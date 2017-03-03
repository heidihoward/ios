package unix

import (
	"bufio"
	"github.com/golang/glog"
	"github.com/heidi-ann/ios/msgs"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Peer struct {
	id      int
	address string
	handled bool // TODO: replace with Mutex
}

var peers []Peer
var peersMutex sync.RWMutex
var id int
var IO *msgs.Io

// iterative through peers and check there is a handler for each
// try to create one if not
func checkPeer() {
	for i := range peers {
		peersMutex.RLock()
		failed := !peers[i].handled
		peersMutex.RUnlock()
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

	// exchange peer ID's via handshake
	_, _ = writer.WriteString(strconv.Itoa(id) + "\n")
	_ = writer.Flush()
	text, _ := reader.ReadString('\n')
	glog.Info("Received ", text)
	peerID, err := strconv.Atoi(strings.Trim(text, "\n"))
	if err != nil {
		glog.Warning(err)
		return
	}

	// check ID is expected
	if peerID < 0 || peerID >= len(peers) || peerID == id {
		glog.Fatal("Unexpected peer ID ", peerID)
	}

	// check IP address is as expected
	// TODO: allow dynamic changes of IP
	expectedAddr := strings.Split(peers[peerID].address, ":")[0]
	actualAddr := strings.Split(addr, ":")[0]
	if expectedAddr != actualAddr {
		glog.Fatal("Peer ID ", peerID, " has connected from an unexpected address ", actualAddr,
			" expected ", expectedAddr)
	}

	glog.Infof("Ready to handle traffic from peer %d at %s ", peerID, addr)

	peersMutex.Lock()
	peers[peerID].handled = true
	peersMutex.Unlock()

	closeErr := make(chan error)
	go func() {
		for {
			// read request
			glog.Infof("Ready for next message from %d", peerID)
			text, err := reader.ReadBytes(byte('\n'))
			if err != nil {
				glog.Warning(err)
				closeErr <- err
				break
			}
			glog.Infof("Read from peer %d: ", peerID, string(text))
			IO.Incoming.BytesToProtoMsg(text)

		}
	}()

	go func() {
		for {
			// send reply
			glog.Infof("Ready to send message to %d", peerID)
			b, err := IO.OutgoingUnicast[peerID].ProtoMsgToBytes()
			if err != nil {
				glog.Fatal("Could not marshal message")
			}
			glog.Infof("Sending to %d: %s", peerID, string(b))
			_, err = writer.Write(b)
			_, err = writer.Write([]byte("\n"))
			if err != nil {
				glog.Warning(err)
				closeErr <- err
				break
			}
			// TODO: BUG need to retry packet
			err = writer.Flush()
			if err != nil {
				glog.Warning(err)
				closeErr <- err
				break
			}
			glog.Info("Sent")
		}
	}()

	// block until connection fails
	<-closeErr

	// tidy up
	glog.Warningf("No longer able to handle traffic from peer %d at %s ", peerID, addr)
	peersMutex.Lock()
	peers[peerID].handled = false
	peersMutex.Unlock()
	IO.Failure <- peerID
	cn.Close()
}

func SetupPeers(localId int, addresses []string, msgIo *msgs.Io) {
	id = localId
	IO = msgIo
	//set up peer state
	peers = make([]Peer, len(addresses))
	for i := range addresses {
		peers[i] = Peer{
			i, addresses[i], false}
	}
	peersMutex = sync.RWMutex{}

	//set up peer server
	glog.Info("Starting up peer server")
	peerPort := strings.Split(addresses[id], ":")[1]
	listeningPort := ":" + peerPort
	lnPeers, err := net.Listen("tcp", listeningPort)
	if err != nil {
		glog.Fatal(err)
	}

	// handle local peer (without sending network traffic)
	peersMutex.Lock()
	peers[id].handled = true
	peersMutex.Unlock()
	from := &(IO.Incoming)
	go from.Forward(IO.OutgoingUnicast[id])

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
}
