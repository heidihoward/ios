package main

import (
	"fmt"
	"net"
	"log"
	"time"
	"bufio"
	"strings"
)

func handleConnection(cn net.Conn) {
	fmt.Printf("Incoming Connection")

	reader := bufio.NewReader(cn)
	writer := bufio.NewWriter(cn)
	keyval := map[string]string {
		"A":"0",
		"B":"0",
		"C":"0",
	}

	for {

		text, err := reader.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("Reading\n")
		fmt.Printf("%s",text)

		request := strings.Split(text," ")
		var reply string

		switch request[0] {
		case "update":
			keyval[request[1]]=request[2]
			reply = "OK"
		case "get":
			reply = keyval[request[1]]
		default: 
			reply = "not reconised"
		}
		fmt.Printf("%s",reply)
		_, err = writer.WriteString(reply)
		if err != nil {
			log.Fatal(err)
		}
	err = writer.Flush()

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
	time.Sleep(30* time.Second)
}