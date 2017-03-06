package unix

import (
	"bufio"
	"github.com/golang/glog"
	"github.com/heidi-ann/ios/app"
	"github.com/heidi-ann/ios/msgs"
	"io"
	"net"
)

var notifyClients *msgs.Notificator
var application *app.StateMachine

func stateMachine() {
	for {
		var req msgs.ClientRequest
		var reply msgs.ClientResponse

		select {
		case response := <-IO.OutgoingResponses:
			req = response.Request
			reply = response.Response
		case req = <-IO.OutgoingRequestsFailed:
			glog.Info("Request could not been safely replicated by consensus algorithm", req)
			reply = msgs.ClientResponse{
				req.ClientID, req.RequestID, false, ""}
		}

		// if any handleRequests are waiting on this reply, then reply to them
		notifyClients.Notify(req, reply)
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
		IO.IncomingRequestsForced <- req
	} else {
		IO.IncomingRequests <- req
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
		if err != nil {
			glog.Fatal(err)
		}
		glog.Info("Finished sending ", n, " bytes")

	}

	cn.Close()
}

func SetupClients(port string, app *app.StateMachine) {
	application = app
	notifyClients = msgs.NewNotificator()
	go stateMachine()

	// set up client server
	glog.Info("Starting up client server")
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
			go handleConnection(conn)
		}
	}()

}
