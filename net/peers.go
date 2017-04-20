package net

import (
	"bufio"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/heidi-ann/ios/config"
	"github.com/heidi-ann/ios/msgs"

	"github.com/golang/glog"
)

type peerHandler struct {
	id       int
	peers    []config.NetAddress
	failures *msgs.FailureNotifier
	net      *msgs.PeerNet
}

// iterative through peers and check if there is a handler for each
// try to create one if not, report failure if not possible
func (ph *peerHandler) checkPeer() {
	for i := range ph.peers {
		if !ph.failures.IsConnected(i) {
			cn, err := net.Dial("tcp", ph.peers[i].ToString())
			if err == nil {
				go ph.handlePeer(cn, true)
			} else {
				go ph.net.OutgoingUnicast[i].Discard()
			}
		}
	}
}

// handlePeer handles a peer connection until closed
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
	if peerID < 0 || peerID >= len(ph.peers) || peerID == ph.id  {
		glog.Warning("Unexpected peer ID ", peerID)
		return
	}
	
	// check IP address is as expected
	// TODO: allow dynamic changes of IP
	actualAddr := strings.Split(addr, ":")[0]
	if ph.peers[peerID].Address != actualAddr {
		glog.Warning("Peer ID ", peerID, " has connected from an unexpected address ", actualAddr,
			" expected ", ph.peers[peerID].Address)
		return
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
			ph.net.Incoming.BytesToProtoMsg(text)

		}
	}()

	go func() {
		for {
			// send reply
			glog.V(1).Infof("Ready to send message to %d", peerID)
			b, err := ph.net.OutgoingUnicast[peerID].ProtoMsgToBytes()
			if err != nil {
				glog.Fatal("Could not marshal message")
			}
			glog.V(1).Infof("Sending to %d: %s", peerID, string(b))
			_, err = writer.Write(b)
			if err != nil {
				glog.Warning(err)
				// return packet for retry
				ph.net.OutgoingUnicast[peerID].BytesToProtoMsg(b)
				closeErr <- err
				break
			}
			_, err = writer.Write([]byte("\n"))
			if err != nil {
				glog.Warning(err)
				// return packet for retry
				ph.net.OutgoingUnicast[peerID].BytesToProtoMsg(b)
				closeErr <- err
				break
			}
			// TODO: BUG need to retry packet
			err = writer.Flush()
			if err != nil {
				glog.Warning(err)
				// return packet for retry
				ph.net.OutgoingUnicast[peerID].BytesToProtoMsg(b)
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

// SetupPeers is an async function to handle/start peer connections
// TODO: switch to sync function
func SetupPeers(localId int, addresses []config.NetAddress, peerNet *msgs.PeerNet, fail *msgs.FailureNotifier) error {
	peerHandler := peerHandler{
		id:       localId,
		peers:    addresses,
		failures: fail,
		net:      peerNet,
	}

	//set up peer server
	glog.Info("Starting up peer server on ", addresses[peerHandler.id].Port)
	listeningPort := ":" + strconv.Itoa(addresses[peerHandler.id].Port)
	lnPeers, err := net.Listen("tcp", listeningPort)
	if err != nil {
		glog.Info("Unable to start listen for peers")
		return err
	}

	// handle local peer (without sending network traffic)
	peerHandler.failures.NowConnected(peerHandler.id)
	from := &(peerHandler.net.Incoming)
	go from.Forward(peerHandler.net.OutgoingUnicast[peerHandler.id])

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

	return nil
}
