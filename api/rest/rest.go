// REST API for Client side Ios
package rest

import (
	"github.com/golang/glog"
	"io"
	"net/http"
	"strings"
	"time"
)

type Rest struct{}

type RestRequest struct {
	Req     string
	ReplyTo http.ResponseWriter
}

var waiting chan RestRequest
var outstanding chan RestRequest

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
	glog.Info("Incoming GET request to", req.URL.String())
	reqs := strings.Split(req.URL.String(), "/")
	reqNew := strings.Join(reqs[2:], " ")
	glog.Info("API request is:", reqNew)
	waiting <- RestRequest{reqNew + "\n", w}

	//wait for response, else give up
	time.Sleep(time.Second)
}

func Create() *Rest {
	port := ":12345"
	glog.Info("Setting up HTTP server on ", port)

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
	waiting = make(chan RestRequest, 10)
	outstanding = make(chan RestRequest, 1)

	return &(Rest{})

}

func (r *Rest) Next() (string, bool, bool) {
	glog.Info("Waiting for next request")
	restreq, ok := <-waiting
	if !ok {
		return "", false, false
	}
	outstanding <- restreq
	glog.Info("Next request received: ", restreq.Req)
	return restreq.Req, true, true
}

func (r *Rest) Return(str string) {
	glog.Info("Response received: ", str)
	restreq := <-outstanding
	io.WriteString(restreq.ReplyTo, str)
	glog.Info("Response sent")

}
