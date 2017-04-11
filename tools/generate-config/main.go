// writes the configurate file for deploying Ios clusters locally
package main

import (
	"bufio"
	"flag"
	"github.com/golang/glog"
	"os"
	"strconv"
)

func writeConfig(servers int, filename string, peerStart int, clientStart int) {
	file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		glog.Fatal(err)
	}
	w := bufio.NewWriter(file)

	w.WriteString("[servers]\n")
	for id := 0; id < servers; id++ {
		peerPort := strconv.Itoa(peerStart + id)
		clientPort := strconv.Itoa(clientStart + id)
		w.WriteString("address = 127.0.0.1:" + peerPort + ":" + clientPort + "\n")
	}
	w.Flush()
	file.Close()
}

func main() {
	servers := flag.Int("servers", 1, "number of servers to run locally")
	filename := flag.String("file", "local.conf", "file to write configuration to")
	peerStart := flag.Int("peer-port", 9080, "port from which to allocate peer ports")
	clientStart := flag.Int("client-port", 8080, "port from which to allocate client ports")
	flag.Parse()
	defer glog.Flush()

	writeConfig(*servers, *filename, *peerStart, *clientStart)
}
