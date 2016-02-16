package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"github.com/golang/glog"
	"github.com/heidi-ann/hydra/api/interactive"
	"github.com/heidi-ann/hydra/api/rest"
	"github.com/heidi-ann/hydra/config"
	"github.com/heidi-ann/hydra/msgs"
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
var auto_file = flag.String("auto", "../test/workload.conf", "If workload is automatically generated, configure file for workload")
var stat_file = flag.String("stat", "latency.csv", "File to write stats to")
var mode = flag.String("mode", "interactive", "interactive, rest or test")
var id = flag.Int("id", -1, "ID of client (must be unique)")

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

// send bytes and wait for reply, return bytes returned if succussful or error otherwise
func dispatcher(b []byte, conn net.Conn, r *bufio.Reader, timeout time.Duration) ([]byte, error) {
	// setup channels for timeout implementation
	errCh := make(chan error, 1)
	replyCh := make(chan []byte, 1)

	go func() {
		// send request
		_, err := conn.Write(b)
		_, err = conn.Write([]byte("\n"))
		if err != nil && err != io.EOF {
			glog.Warning(err)
			errCh <- err
		}

		glog.Info("Sent")

		// read response
		reply, err := r.ReadBytes('\n')
		if err != nil && err != io.EOF {
			glog.Warning(err)
			errCh <- err
		}

		// success, return reply
		replyCh <- reply
	}()

	//handling outcomes
	select {
	case reply := <-replyCh:
		return reply, nil
	case err := <-errCh:
		return nil, err
	case <-time.After(timeout):
		return nil, errors.New("Timeout")
	}
}

func main() {
	// set up logging
	flag.Parse()
	defer glog.Flush()

	// parse config files
	conf := config.ParseClientConfig(*config_file)
	timeout := time.Millisecond * time.Duration(conf.Parameters.Timeout)
	// TODO: find a better way to handle required flags
	if *id == -1 {
		glog.Fatal("ID must be provided")
	}

	glog.Info("Starting up client ", *id)
	defer glog.Info("Shutting down client ", *id)

	// set up stats collection
	filename := *stat_file
	glog.Info("Opening file: ", filename)
	file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0777)
	if err != nil {
		glog.Fatal(err)
	}
	stats := bufio.NewWriter(file)
	defer stats.Flush()

	// set up request id
	// TODO: write this value to disk
	requestID := 0

	// connecting to server
	conn, err := connect(conf.Addresses.Address, 3)
	if err != nil {
		glog.Fatal(err)
	}
	rd := bufio.NewReader(conn)

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

	glog.Info("Client is ready to process incoming requests")
	for {

		// get next command
		text, ok := ioapi.Next()
		if !ok {
			glog.Fatal("No more commands")
		}
		glog.Info("API produced ", text)

		// encode as request
		req := msgs.ClientRequest{
			*id, requestID, text}
		requestID++
		b, err := msgs.Marshal(req)
		if err != nil {
			glog.Fatal(err)
		}
		glog.Info(string(b))

		startTime := time.Now()

		// dispatch request until successfull
		var replyBytes []byte
		for {
			replyBytes, err = dispatcher(b, conn, rd, timeout)
			if err == nil {
				break
			}
			conn, err = connect(conf.Addresses.Address, conf.Parameters.Retries)
			if err != nil {
				glog.Fatal(err)
			}
			rd = bufio.NewReader(conn)

		}

		//handle reply
		reply := new(msgs.ClientResponse)
		msgs.Unmarshal(replyBytes, reply)

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
		ioapi.Return(reply.Response)

	}

}
