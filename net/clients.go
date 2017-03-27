package net

import (
	"bufio"
	"github.com/golang/glog"
	"github.com/heidi-ann/ios/app"
	"github.com/heidi-ann/ios/msgs"
	"io"
	"net"
)

type clientHandler struct {
	notify    *msgs.Notificator
	app       *app.StateMachine
	clientNet *msgs.ClientNet
}

func (ch *clientHandler) stateMachine() {
	for {
		var req msgs.ClientRequest
		var reply msgs.ClientResponse

		select {
		case response := <-ch.clientNet.OutgoingResponses:
			req = response.Request
			reply = response.Response
		case req = <-ch.clientNet.OutgoingRequestsFailed:
			glog.V(1).Info("Request could not been safely replicated by consensus algorithm", req)
			reply = msgs.ClientResponse{
				req.ClientID, req.RequestID, false, ""}
		}

		// if any handleRequests are waiting on this reply, then reply to them
		ch.notify.Notify(req, reply)
	}
}

func (ch *clientHandler) handleRequest(req msgs.ClientRequest) msgs.ClientResponse {
	glog.V(1).Info("Handling ", req.Request)

	// check if already applied
	if found, res := ch.app.Check(req); found {
		glog.V(1).Info("Request found in cache")
		return res // FAST PASS
	}

	// CONSENESUS ALGORITHM HERE
	glog.V(1).Info("Passing request to consensus algorithm")
	ch.clientNet.IncomingRequests <- req

	if ch.notify.IsSubscribed(req) {
		glog.Warning("Client has multiple outstanding connections for the same request, usually not a good sign")
	}

	// wait for reply
	reply := ch.notify.Subscribe(req)

	// check reply is as expected
	if reply.ClientID != req.ClientID {
		glog.Fatal("ClientID is different")
	}
	if reply.RequestID != req.RequestID {
		glog.Fatal("RequestID is different")
	}

	return reply
}

func (ch *clientHandler) handleConnection(cn net.Conn) {
	glog.Info("Incoming client connection from ",
		cn.RemoteAddr().String())

	reader := bufio.NewReader(cn)
	writer := bufio.NewWriter(cn)

	for {

		// read request
		glog.V(1).Info("Ready for Reading")
		text, err := reader.ReadBytes(byte('\n'))
		if err != nil {
			if err == io.EOF {
				break
			}
			glog.Warning(err)
			break
		}
		glog.V(1).Info("--------------------New request----------------------")
		glog.V(1).Info("Request: ", string(text))
		req := new(msgs.ClientRequest)
		err = msgs.Unmarshal(text, req)
		if err != nil {
			glog.Fatal(err)
		}

		// construct reply
		reply := ch.handleRequest(*req)
		b, err := msgs.Marshal(reply)
		if err != nil {
			glog.Fatal("error:", err)
		}
		glog.V(1).Info(string(b))

		// send reply
		glog.V(1).Info("Sending ", string(b))
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
		if err != nil {
			glog.Fatal(err)
		}
		glog.V(1).Info("Finished sending ", n, " bytes")

	}

	cn.Close()
}

// SetupClients listen for client on the given port, it forwards their requests to the consensus algorithm and
// then applies them to the state machine
// SetupClients returns when setup is completed, spawning goroutines to listen for clients.
func SetupClients(port string, app *app.StateMachine, clientNet *msgs.ClientNet) {
	ch := clientHandler{
		notify:    msgs.NewNotificator(),
		app:       app,
		clientNet: clientNet}

	go ch.stateMachine()

	// set up client server
	glog.Info("Starting up client server on port ", port)
	listeningPort := ":" + port
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
			go ch.handleConnection(conn)
		}
	}()

}
