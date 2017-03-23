package net

import (
	"bufio"
	"github.com/golang/glog"
	"github.com/heidi-ann/ios/msgs"
	"net"
	"strconv"
	"strings"
	"time"
)

type peer struct {
	id      int
	address string
}

type peerHandler struct {
	id int
	peers []peer
	failures *msgs.FailureNotifier
	iO *msgs.Io
}


// iterative through peers and check if there is a handler for each
// try to create one if not, report failure if not possible
func (ph *peerHandler) checkPeer() {
	for i := range ph.peers {
		if !ph.failures.IsConnected(i) {
			//glog.V(1).Info("Peer ", i, " is not currently connected")
			cn, err := net.Dial("tcp", ph.peers[i].address)

			if err != nil {
				//glog.Warning(err)
			} else {
				go ph.handlePeer(cn, true)
			}
		} else {
			//glog.V(1).Info("Peer ", i, " is currently connected")
		}
	}
}

func (ph *peerHandler) handlePeer(cn net.Conn, init bool) {
	addr := cn.RemoteAddr().String()
	if init {
		glog.Info("Outgoing peer connection to ", addr)
	} else {
		glog.Info("Incoming peer connection from ", addr)
	}

	defer cn.Close()
	defer glog.Warningf("Connection closed from %s ", addr)

	// handle requests
	reader := bufio.NewReader(cn)
	writer := bufio.NewWriter(cn)

	// exchange peer ID's via handshake
	_, _ = writer.WriteString(strconv.Itoa(ph.id) + "\n")
	_ = writer.Flush()
	text, _ := reader.ReadString('\n')
	glog.V(1).Info("Received ", text)
	peerID, err := strconv.Atoi(strings.Trim(text, "\n"))
	if err != nil {
		glog.Warning(err)
		return
	}

	// check ID is expected
	if peerID < 0 {
		glog.Fatal("Unexpected peer ID ", peerID, " ID must be > 0")
	}
	if peerID >= len(ph.peers) {
		glog.Fatal("Unexpected peer ID ", peerID, " IDs within this cluster should be between 0 and ", len(ph.peers))
	}
	if peerID == ph.id {
		glog.Fatal("Unexpected peer ID ", peerID, " I seem to be talking to myself")
	}

	// check IP address is as expected
	// TODO: allow dynamic changes of IP
	expectedAddr := strings.Split(ph.peers[peerID].address, ":")[0]
	actualAddr := strings.Split(addr, ":")[0]
	if expectedAddr != actualAddr {
		glog.Fatal("Peer ID ", peerID, " has connected from an unexpected address ", actualAddr,
			" expected ", expectedAddr)
	}

	glog.Infof("Ready to handle traffic from peer %d at %s ", peerID, addr)
	err = ph.failures.NowConnected(peerID)
	if err != nil {
		glog.Warning(err)
		return
	}

	closeErr := make(chan error)
	go func() {
		for {
			// read request
			glog.V(1).Infof("Ready for next message from %d", peerID)
			text, err := reader.ReadBytes(byte('\n'))
			if err != nil {
				glog.Warning(err)
				closeErr <- err
				break
			}
			glog.V(1).Infof("Read from peer %d: ", peerID, string(text))
			ph.iO.Incoming.BytesToProtoMsg(text)

		}
	}()

	go func() {
		for {
			// send reply
			glog.V(1).Infof("Ready to send message to %d", peerID)
			b, err := ph.iO.OutgoingUnicast[peerID].ProtoMsgToBytes()
			if err != nil {
				glog.Fatal("Could not marshal message")
			}
			glog.V(1).Infof("Sending to %d: %s", peerID, string(b))
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
			glog.V(1).Info("Sent")
		}
	}()

	// block until connection fails
	<-closeErr

	// tidy up
	glog.Warningf("No longer able to handle traffic from peer %d at %s ", peerID, addr)
	ph.failures.NowDisconnected(peerID)
}

func SetupPeers(localId int, addresses []string, msgIo *msgs.Io, fail *msgs.FailureNotifier) {
	peerHandler := peerHandler{
		id: localId,
		peers: make([]peer, len(addresses)),
		failures: fail,
		iO: msgIo,
	}

	for i := range addresses {
		peerHandler.peers[i] = peer{
			i, addresses[i]}
	}

	//set up peer server
	glog.Info("Starting up peer server on ", addresses[peerHandler.id])
	peerPort := strings.Split(addresses[peerHandler.id], ":")[1]
	listeningPort := ":" + peerPort
	lnPeers, err := net.Listen("tcp", listeningPort)
	if err != nil {
		glog.Fatal(err)
	}

	// handle local peer (without sending network traffic)
	peerHandler.failures.NowConnected(peerHandler.id)
	from := &(peerHandler.iO.Incoming)
	go from.Forward(peerHandler.iO.OutgoingUnicast[peerHandler.id])

	// handle for incoming peers
	go func() {
		for {
			conn, err := lnPeers.Accept()
			if err != nil {
				glog.Fatal(err)
			}
			go (&peerHandler).handlePeer(conn, false)
		}
	}()

	// regularly check if all peers are connected and retry if not
	go func() {
		for {
			(&peerHandler).checkPeer()
			time.Sleep(500 * time.Millisecond)
		}
	}()
}
