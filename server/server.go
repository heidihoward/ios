package main

import (
	"bufio"
	"fmt"
	"github.com/heidi-ann/hydra/store"
	"log"
	"net"
	"time"
)

func handleConnection(cn net.Conn) {
	fmt.Printf("Incoming Connection\n")

	reader := bufio.NewReader(cn)
	writer := bufio.NewWriter(cn)
	keyval := store.New()

	for {

		text, err := reader.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("Reading\n")
		fmt.Printf("%s", text)

		reply := keyval.Process(text)

		fmt.Printf("%s", reply)
		n, err := writer.WriteString(reply)
		if err != nil {
			log.Fatal(err)
		}
		err = writer.Flush()
		fmt.Printf("Finished sending %b", n)

	}
}

func main() {
	fmt.Printf("Starting up")
	ln, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatal(err)
	}
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Fatal(err)
		}
		go handleConnection(conn)
	}
	time.Sleep(30 * time.Second)
}
