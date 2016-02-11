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
var auto_file = flag.String("auto", "", "If workload is automatically generated, configure file for workload")

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
	interactive_mode := (*auto_file == "")
	var gen store.Generator
	if !interactive_mode {
		gen = store.Generate(config.ParseAuto(*auto_file))
	}

	// set up stats collection
	filename := "latency.csv"
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

	// mian mian
	term_reader := bufio.NewReader(os.Stdin)
	net_reader := bufio.NewReader(conn)

	for {
		text := ""

		if interactive_mode {
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

		// write to latency to log
		str := fmt.Sprintf("%d\n", time.Since(startTime).Nanoseconds())
		n, err := stats.WriteString(str)
		if err != nil {
			glog.Fatal(err)
		}
		glog.Warningf("Written %d bytes to log", n)
		_ = stats.Flush()

		// writing result to user
		if interactive_mode {
			fmt.Print(reply, "request took ", time.Since(startTime))
		}

	}

}
