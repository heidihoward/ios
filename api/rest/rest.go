// REST API for Client side Hydra
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
	io.WriteString(w, "hydra 0.1\n")
}

// hello world, the web server
func requestServer(w http.ResponseWriter, req *http.Request) {
	glog.Info("Incoming GET request to", req.URL.String())
	reqs := strings.Split(req.URL.String(), "/")
	reqNew := strings.Join(reqs[2:], " ")
	glog.Info("API request is:", reqNew)
	waiting <- RestRequest{reqNew + "\n", w}
	time.Sleep(time.Second)
}

func Create() *Rest {
	port := ":12345"
	glog.Info("Setting up HTTP server on ", port)

	//setup HTTP server
	http.HandleFunc("/request/", requestServer)
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

func (r *Rest) Next() (string, bool) {
	glog.Info("Waiting for next request")
	restreq := <-waiting
	outstanding <- restreq
	glog.Info("Next request received: ", restreq.Req)
	return restreq.Req, true
}

func (r *Rest) Return(str string) {
	glog.Info("Response received: ", str)
	restreq := <-outstanding
	io.WriteString(restreq.ReplyTo, str)
	glog.Info("Response sent")

}
