package unix

import (
	"bufio"
	"github.com/golang/glog"
	"github.com/heidi-ann/ios/msgs"
	"sync"
  "net"
  "strconv"
  "strings"
  "time"
)

type Peer struct {
	id      int
	address string
	handled bool // TODO: replace with Mutex
}

var peers []Peer
var peers_mutex sync.RWMutex
var id int
var io *msgs.Io

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

	// exchange peer ID's via handshake
	_, _ = writer.WriteString(strconv.Itoa(id) + "\n")
	_ = writer.Flush()
	text, _ := reader.ReadString('\n')
	glog.Info("Received ", text)
	peer_id, err := strconv.Atoi(strings.Trim(text, "\n"))
	if err != nil {
		glog.Warning(err)
		return
	}

	// check ID is expected
	if peer_id < 0 || peer_id >= len(peers) || peer_id == id {
		glog.Fatal("Unexpected peer ID ", peer_id)
	}

	// check IP address is as expected
	// TODO: allow dynamic changes of IP
	expectedAddr := strings.Split(peers[peer_id].address,":")[0]
	actualAddr := strings.Split(addr,":")[0]
	if expectedAddr != actualAddr {
		glog.Fatal("Peer ID ",peer_id," has connected from an unexpected address ",actualAddr,
			" expected ",expectedAddr)
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
			io.Incoming.BytesToProtoMsg(text)

		}
	}()

	go func() {
		for {
			// send reply
			glog.Infof("Ready to send message to %d", peer_id)
			b, err := io.OutgoingUnicast[peer_id].ProtoMsgToBytes()
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
	io.Failure <- peer_id
	cn.Close()
}

func SetupPeers(localId int, addresses []string, msgIo *msgs.Io) {
  id = localId
  io = msgIo
  //set up peer state
  peers = make([]Peer, len(addresses))
  for i := range addresses {
    peers[i] = Peer{
      i, addresses[i], false}
  }
  peers_mutex = sync.RWMutex{}

  //set up peer server
  glog.Info("Starting up peer server")
  peer_port := strings.Split(addresses[id],":")[1]
  listeningPort := ":" + peer_port
  lnPeers, err := net.Listen("tcp", listeningPort)
  if err != nil {
    glog.Fatal(err)
  }

  // handle local peer (without sending network traffic)
  peers_mutex.Lock()
  peers[id].handled = true
  peers_mutex.Unlock()
  from := &(io.Incoming)
  go from.Forward(io.OutgoingUnicast[id])

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
