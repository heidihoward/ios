package main

import (
	"bufio"
	"flag"
	"github.com/golang/glog"
	"github.com/heidi-ann/hydra/store"
	"net"
	"time"
)

func handleConnection(cn net.Conn) {
	glog.Info("Incoming Connection from ",
		cn.RemoteAddr().String())

	reader := bufio.NewReader(cn)
	writer := bufio.NewWriter(cn)
	keyval := store.New()

	for {

		// read request
		glog.Info("Reading")
		text, err := reader.ReadString('\n')
		if err != nil {
			glog.Fatal(err)
		}
		glog.Info("Received ", text)

		// apply request
		reply := keyval.Process(text)
		keyval.Print()

		// send reply
		glog.Info("Sending ", reply)
		reply = reply + "\n"
		n, err := writer.WriteString(reply)
		if err != nil {
			glog.Fatal(err)
		}

		// tidy up
		err = writer.Flush()
		glog.Info("Finished sending ", n, " bytes")

	}

	cn.Close()
}

func main() {
	// set up logging
	flag.Parse()
	defer glog.Flush()

	// set up server
	glog.Info("Starting up")
	ln, err := net.Listen("tcp", ":8080")
	if err != nil {
		glog.Fatal(err)
	}

	// handle for incoming clients
	for {
		conn, err := ln.Accept()
		if err != nil {
			glog.Fatal(err)
		}
		go handleConnection(conn)
	}

	// tidy up
	time.Sleep(30 * time.Second)
	glog.Info("Shutting down")
}
