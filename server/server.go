package main

import (
	"bufio"
	"flag"
	"github.com/golang/glog"
	"github.com/heidi-ann/hydra/store"
	"io"
	"net"
	"os"
	"strconv"
	"time"
)

var keyval *store.Store
var disk *bufio.Writer

var port = flag.Int("port", 8080, "port to listen on")

func handleConnection(cn net.Conn) {
	glog.Info("Incoming Connection from ",
		cn.RemoteAddr().String())

	reader := bufio.NewReader(cn)
	writer := bufio.NewWriter(cn)

	for {

		// read request
		glog.Info("Reading")
		text, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			glog.Fatal(err)
		}
		glog.Info("Received ", text)

		// write to persistent storage
		n, err := disk.WriteString(text)
		_ = disk.Flush()
		if err != nil {
			glog.Fatal(err)
		}
		glog.Infof("Written %b bytes to persistent storage", n)

		// apply request
		reply := keyval.Process(text)
		keyval.Print()
		time.Sleep(100 * time.Millisecond)

		// send reply
		glog.Info("Sending ", reply)
		reply = reply + "\n"
		n, err = writer.WriteString(reply)
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
	filename := "persistent.log"

	// set up logging
	flag.Parse()
	defer glog.Flush()

	//set up state machine
	keyval = store.New()

	// setting up persistent log
	glog.Info("Opening file: ", filename)
	file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0777)
	if err != nil {
		glog.Fatal(err)
	}
	disk = bufio.NewWriter(file)
	defer disk.Flush()

	// check persistent storage for commands
	disk_reader := bufio.NewReader(file)
	for {
		str, err := disk_reader.ReadString('\n')
		if err != nil {
			glog.Info("No more commands in persistent storage")
			break
		}
		_ = keyval.Process(str)
	}

	// set up server
	glog.Info("Starting up")
	listeningPort := ":" + strconv.Itoa(*port)
	ln, err := net.Listen("tcp", listeningPort)
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
