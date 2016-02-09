package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/golang/glog"
	"net"
	"os"
)

var ip = flag.String("ip", "127.0.0.1", "IP address of server")
var port = flag.Int("port", 8080, "Listening port of server")

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

	for {

		// reading from terminal
		fmt.Print("Enter command: ")
		text, err := term_reader.ReadString('\n')
		if err != nil {
			glog.Fatal(err)
		}
		glog.Info("User entered", text)

		// send to server
		_, err = conn.Write([]byte(text))
		if err != nil {
			glog.Fatal(err)
		}
		glog.Info("Sent")

		// read response
		reply, err := net_reader.ReadString('\n')
		if err != nil {
			glog.Fatal(err)
		}

		// write to user
		fmt.Print(reply)

	}

}
