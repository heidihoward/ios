package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/golang/glog"
	"github.com/heidi-ann/hydra/config"
	"github.com/heidi-ann/hydra/store"
	"io"
	"net"
	"os"
	"time"
)

var config_file = flag.String("config", "exampleconfig", "Client configuration file")
var auto = flag.Int("auto", -1, "If workload is automatically generated, percentage of reads")

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

	// parse config file
	conf := config.Parse(*config_file)

	// set up stats collection
	filename := "latency.csv"
	glog.Info("Opening file: %s", filename)
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_APPEND, 0777)
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

	// mian mian
	term_reader := bufio.NewReader(os.Stdin)
	net_reader := bufio.NewReader(conn)
	gen := store.Generate(*auto, 50)

	for {
		text := ""

		if *auto == -1 {
			// reading from terminal
			fmt.Print("Enter command: ")
			text, err = term_reader.ReadString('\n')
			if err != nil {
				glog.Fatal(err)
			}
			glog.Info("User entered", text)
		} else {
			// use generator
			text = gen.Next()
			glog.Info("Generator produced ", text)
		}

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

		// write to user
		_, err = stats.WriteString(fmt.Sprintf("%b,%b", startTime, time.Since(startTime)))
		fmt.Print(reply, "request took ", time.Since(startTime))

	}

}
