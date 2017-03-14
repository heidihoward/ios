// Package rest provides a REST API for client to interact with Ios clusters.
package rest

import (
	"github.com/golang/glog"
	"io"
	"net/http"
	"strings"
	"time"
)

// Rest is a placeholder
type Rest struct{}

type restrequest struct {
	Req     string
	ReplyTo http.ResponseWriter
}

var waiting chan restrequest
var outstanding chan restrequest

func versionServer(w http.ResponseWriter, req *http.Request) {
	io.WriteString(w, "ios 0.1\n")
}

func closeServer(w http.ResponseWriter, req *http.Request) {
	close(waiting)
	io.WriteString(w, "Will do\n")
}

// main request handler
func requestServer(w http.ResponseWriter, req *http.Request) {
	// NB: ResponseWriter needs to be used before this function exits
	glog.V(1).Info("Incoming GET request to", req.URL.String())
	reqs := strings.Split(req.URL.String(), "/")
	reqNew := strings.Join(reqs[2:], " ")
	glog.V(1).Info("API request is:", reqNew)
	waiting <- restrequest{reqNew + "\n", w}

	//wait for response, else give up
	time.Sleep(time.Second)
}

func Create() *Rest {
	port := ":12345"
	glog.V(1).Info("Setting up HTTP server on ", port)

	//setup HTTP server
	http.HandleFunc("/request/", requestServer)
	http.HandleFunc("/close", closeServer)
	http.HandleFunc("/version", versionServer)
	go func() {
		err := http.ListenAndServe(port, nil)
		if err != nil {
			glog.Fatal("ListenAndServe: ", err)
		}
	}()

	// setup channels
	waiting = make(chan restrequest, 10)
	outstanding = make(chan restrequest, 1)

	return &(Rest{})

}

// Next returns the next request for the state machine or false if wishes to terminate
func (r *Rest) Next() (string, bool) {
	glog.V(1).Info("Waiting for next request")
	restreq, ok := <-waiting
	if !ok {
		return "", false
	}
	outstanding <- restreq
	glog.V(1).Info("Next request received: ", restreq.Req)
	return restreq.Req, true
}

// Return provides the REST API with the response to the current outstanding request
func (r *Rest) Return(str string) {
	glog.V(1).Info("Response received: ", str)
	restreq := <-outstanding
	io.WriteString(restreq.ReplyTo, str)
	glog.V(1).Info("Response sent")

}
