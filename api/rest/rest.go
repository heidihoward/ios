// REST API for Client side Hydra
package rest

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
)

type Rest struct {
	ReplyTo http.ResponseWriter
}

func versionServer(w http.ResponseWriter, req *http.Request) {
	io.WriteString(w, "hydra 0.1\n")
}

// hello world, the web server
func requestServer(w http.ResponseWriter, req *http.Request) {
	reqs := strings.Trim(req.URL.String(), "/request/")
	reqs = strings.Replace(reqs, "/", " ", -1)
	io.WriteString(w, reqs)
}

func Create() *Rest {

	//setup HTTP server
	http.HandleFunc("/request/", requestServer)
	http.HandleFunc("/version", versionServer)
	err := http.ListenAndServe(":12345", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}

	return &(Rest{})

}

func (r *Rest) Next() (string, bool) {
	if r.ReplyTo != nil {

	}
	return "", true
}

func (r *Rest) Return(str string) {
	io.WriteString(r.ReplyTo, reqs)
	r.ReplyTo = nil

}
