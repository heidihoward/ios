package main

import (
	"fmt"
	"flag"
	"net"
	"log"
	"bufio"
	"os"
)

var ip = flag.String("ip","127.0.0.1","IP address of server")
var port = flag.Int("port",8080,"Listening port of server")

func main() {
	flag.Parse()

	var address = fmt.Sprintf("%s:%d",*ip,*port)
	fmt.Printf("Talking to %s \n",address)

	conn, err := net.Dial("tcp", address)
	if err != nil {
		log.Fatal(err)
	}

    term_reader := bufio.NewReader(os.Stdin)
    net_reader := bufio.NewReader(conn)

    for {
	    fmt.Print("Enter command: ")
	    text, _ := term_reader.ReadString('\n')
	    _,err = conn.Write([]byte(text))

	    reply, _ := net_reader.ReadString('\n')
	    fmt.Print(reply)
	    if err != nil {
			log.Fatal(err)
		}
	}

}