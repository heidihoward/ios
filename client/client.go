package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/golang/glog"
	"github.com/heidi-ann/hydra/api/interactive"
	"github.com/heidi-ann/hydra/api/rest"
	"github.com/heidi-ann/hydra/config"
	"github.com/heidi-ann/hydra/test"
	"io"
	"net"
	"os"
	"time"
)

type API interface {
	Next() (string, bool)
	Return(string)
}

var config_file = flag.String("config", "example.conf", "Client configuration file")
var auto_file = flag.String("auto", "", "If workload is automatically generated, configure file for workload")
var stat_file = flag.String("stat", "latency.csv", "File to write stats to")
var mode = flag.String("mode", "interactive", "interactive, rest or test")

func connect(addrs []string, tries int) (net.Conn, error) {
	var conn net.Conn
	var err error

	for i := range addrs {
		for t := tries; t > 0; t-- {
			conn, err = net.Dial("tcp", addrs[i])

			// if successful
			if err == nil {
				glog.Infof("Connected to %s", addrs[i])
				return conn, err
			}

			//if unsuccessful
			glog.Warning(err)
			time.Sleep(2 * time.Second)
		}
	}

	return conn, err
}

func main() {
	// set up logging
	flag.Parse()
	defer glog.Flush()

	// parse config files
	conf := config.Parse(*config_file)

	// set up stats collection
	filename := *stat_file
	glog.Info("Opening file: ", filename)
	file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0777)
	if err != nil {
		glog.Fatal(err)
	}
	stats := bufio.NewWriter(file)
	defer stats.Flush()

	// connecting to server
	conn, err := connect(conf.Addresses.Address, 3)
	if err != nil {
		glog.Fatal(err)
	}

	// setup API
	var ioapi API
	switch *mode {
	case "interactive":
		ioapi = interactive.Create()
	case "test":
		ioapi = test.Generate(test.ParseAuto(*auto_file))
	case "rest":
		ioapi = rest.Create()
	default:
		glog.Fatal("Invalid mode: ", mode)
	}

	// setup network reader
	net_reader := bufio.NewReader(conn)

	glog.Info("Client is ready to process incoming requests")
	for {

		// get next command
		text, ok := ioapi.Next()
		if !ok {
			glog.Fatal("No more commands")
		}
		glog.Info("API produced ", text)

		// send to server
		startTime := time.Now()
		_, err = conn.Write([]byte(text))
		if err != nil {
			if err == io.EOF {
				continue
			}
			glog.Warning(err)
			// reconnecting to server
			conn, err = connect(conf.Addresses.Address, 3)
			if err != nil {
				glog.Fatal(err)
			}
		}

		glog.Info("Sent ", text)

		// read response
		reply, err := net_reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				continue
			}
			glog.Warning(err)
			// reconnecting to server
			conn, err = connect(conf.Addresses.Address, 3)
			if err != nil {
				glog.Fatal(err)
			}
		}

		// write to latency to log
		str := fmt.Sprintf("%d\n", time.Since(startTime).Nanoseconds())
		n, err := stats.WriteString(str)
		if err != nil {
			glog.Fatal(err)
		}
		glog.Warningf("Written %d bytes to log", n)
		_ = stats.Flush()

		// writing result to user
		// time.Since(startTime)
		ioapi.Return(reply)

	}

}
