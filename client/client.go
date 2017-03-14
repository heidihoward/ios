// Package client provides I/O for Ios clients
package main

import (
	"bufio"
	"encoding/csv"
	"errors"
	"flag"
	"github.com/golang/glog"
	"github.com/heidi-ann/ios/api/interactive"
	"github.com/heidi-ann/ios/api/rest"
	"github.com/heidi-ann/ios/config"
	"github.com/heidi-ann/ios/msgs"
	"github.com/heidi-ann/ios/test"
	"io"
	"math/rand"
	"net"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
)

type API interface {
	Next() (string, bool, bool)
	Return(string)
}

var configFile = flag.String("config", os.Getenv("GOPATH")+"/src/github.com/heidi-ann/ios/client/example.conf", "Client configuration file")
var autoFile = flag.String("auto", os.Getenv("GOPATH")+"/src/github.com/heidi-ann/ios/test/workload.conf", "If workload is automatically generated, configure file for workload")
var statFile = flag.String("stat", "latency.csv", "File to write stats to")
var mode = flag.String("mode", "interactive", "interactive, rest or test")
var id = flag.Int("id", -1, "ID of client (must be unique) or random number will be generated")

func connect(addrs []string, tries int, hint int) (net.Conn, int, error) {
	var conn net.Conn
	var err error

	// reset invalid hint
	if len(addrs) >= hint {
		hint = 0
	}

	// first, try on to connect to the most likely leader
	glog.V(1).Info("Trying to connect to ", addrs[hint])
	conn, err = net.Dial("tcp", addrs[hint])
	// if successful
	if err == nil {
		glog.V(1).Infof("Connect established to %s", addrs[hint])
		return conn, hint, err
	}
	//if unsuccessful
	glog.Warning(err)

	// if fails, try everyone else
	for i := range addrs {
		for t := tries; t > 0; t-- {
			glog.V(1).Info("Trying to connect to ", addrs[i])
			conn, err = net.Dial("tcp", addrs[i])

			// if successful
			if err == nil {
				glog.V(1).Infof("Connect established to %s", addrs[i])
				return conn, i, err
			}

			//if unsuccessful
			glog.Warning(err)
			time.Sleep(100 * time.Millisecond)
		}
	}

	// calc most likely next leader
	hint += 1
	if len(addrs) == hint {
		hint = 0
	}
	return conn, hint, err
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

		glog.V(1).Info("Sent")

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

	// always flush (whatever happens)
	sigs := make(chan os.Signal, 1)
	finish := make(chan bool, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	// parse config files
	conf := config.ParseClientConfig(*configFile)
	timeout := time.Millisecond * time.Duration(conf.Parameters.Timeout)
	// TODO: find a better way to handle required flags
	if *id == -1 {
		rand.Seed(time.Now().UTC().UnixNano())
		*id = rand.Int()
		glog.V(1).Info("ID was not provided, ID ", *id, " has been assigned")
	}

	glog.V(1).Info("Starting up client ", *id)
	defer glog.V(1).Info("Shutting down client ", *id)

	// set up stats collection
	filename := *statFile
	glog.V(1).Info("Opening file: ", filename)
	file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0777)
	if err != nil {
		glog.Fatal(err)
	}
	stats := csv.NewWriter(file)
	defer stats.Flush()

	// set up request id
	// TODO: write this value to disk
	requestID := 1

	// connecting to server
	conn, leader, err := connect(conf.Addresses.Address, 10, 0)
	if err != nil {
		glog.Fatal(err)
	}
	rd := bufio.NewReader(conn)

	// setup API
	var ioapi API
	switch *mode {
	case "interactive":
		ioapi = interactive.Create(conf.Parameters.Application)
	case "test":
		ioapi = test.Generate(test.ParseAuto(*autoFile))
	case "rest":
		ioapi = rest.Create()
	default:
		glog.Fatal("Invalid mode: ", mode)
	}

	glog.V(1).Info("Client is ready to start processing incoming requests")
	go func() {
		for {
			// get next command
			text, _, ok := ioapi.Next()
			if !ok {
				finish <- true
				break
			}
			glog.V(1).Info("Request ", requestID, " is: ", text)

			// prepare request
			req := msgs.ClientRequest{
				*id, requestID, false, text}
			b, err := msgs.Marshal(req)
			if err != nil {
				glog.Fatal(err)
			}
			glog.V(1).Info(string(b))

			startTime := time.Now()
			tries := 0
			var reply *msgs.ClientResponse

			// dispatch request until successful
			for {
				tries++
				if tries > len(conf.Addresses.Address) {
					req.ForceViewChange = true
					b, err = msgs.Marshal(req)
					if err != nil {
						glog.Fatal(err)
					}
				}

				replyBytes, err := dispatcher(b, conn, rd, timeout)
				if err == nil {

					//handle reply
					reply = new(msgs.ClientResponse)
					err = msgs.Unmarshal(replyBytes, reply)

					if err == nil && !reply.Success {
						err = errors.New("request marked by server as unsuccessful")
					}
					if err == nil && reply.Success {
						glog.V(1).Info("request was Successful", reply)
						break
					}
				}
				glog.Warning("Request ", requestID, " failed due to: ", err)

				// try to establish a new connection
				for {
					if conn != nil {
						err = conn.Close()
						if err != nil {
							glog.Warning(err)
						}
					}
					conn, leader, err = connect(conf.Addresses.Address, conf.Parameters.Retries, leader+1)
					if err == nil {
						break
					}
					glog.Warning("Serious connectivity issues")
					time.Sleep(time.Second)
				}

				rd = bufio.NewReader(conn)

			}

			//check reply is not nil
			if *reply == (msgs.ClientResponse{}) {
				glog.Fatal("Response is nil")
			}

			//check reply is as expected
			if reply.ClientID != *id {
				glog.Fatal("Response received has wrong ClientID: expected ",
					*id, " ,received ", reply.ClientID)
			}
			if reply.RequestID != requestID {
				glog.Fatal("Response received has wrong RequestID: expected ",
					requestID, " ,received ", reply.RequestID)
			}
			if !reply.Success {
				glog.Fatal("Response marked as unsuccessful but not retried")
			}

			// write to latency to log
			latency := strconv.FormatInt(time.Since(startTime).Nanoseconds(), 10)
			err = stats.Write([]string{strconv.FormatInt(startTime.UnixNano(), 10), strconv.Itoa(requestID), latency, strconv.Itoa(tries)})
			if err != nil {
				glog.Fatal(err)
			}
			stats.Flush()
			// TODO: call error to check if successful

			requestID++
			// writing result to user
			// time.Since(startTime)
			ioapi.Return(reply.Response)

		}
	}()

	select {
	case sig := <-sigs:
		glog.Warning("Termination due to: ", sig)
	case <-finish:
		glog.V(1).Info("No more commands")
	}
	glog.Flush()

}
