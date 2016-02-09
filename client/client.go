package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/golang/glog"
	"github.com/heidi-ann/hydra/store"
	"io"
	"net"
	"os"
)

var ip = flag.String("ip", "127.0.0.1", "IP address of server")
var port = flag.Int("port", 8080, "Listening port of server")
var auto = flag.Int("auto", -1, "If workload is automatically generated, percentage of reads")

func main() {
	// set up logging
	flag.Parse()
	defer glog.Flush()

	// connecting to server
	address := fmt.Sprintf("%s:%d", *ip, *port)
	conn, err := net.Dial("tcp", address)
	if err != nil {
		glog.Fatal(err)
	}
	glog.Infof("Connected to %s", address)

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
		_, err = conn.Write([]byte(text))
		if err != nil {
			if err == io.EOF {
				continue
			}
			glog.Fatal(err)
		}
		glog.Info("Sent ", text)

		// read response
		reply, err := net_reader.ReadString('\n')
		if err != nil {
			glog.Fatal(err)
		}

		// write to user
		fmt.Print(reply)

	}

}
