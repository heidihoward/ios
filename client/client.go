// Package client provides Ios client side code for connecting to an Ios cluster
package client

import (
	"bufio"
	"errors"
	"io"
	"math/rand"
	"net"
	"time"

	"github.com/golang/glog"
	"github.com/heidi-ann/ios/config"
	"github.com/heidi-ann/ios/msgs"
)

//TODO: write requestID to disk

// Client holds the data associated with a client
type Client struct {
	id          int // ID of client, must be unique
	requestID   int // ID of current request, starting from 1
	stats       *statsFile // handler for stats collection, maybe nil
	servers     []config.NetAddress // address of Ios servers
	conn        net.Conn
	rd          *bufio.Reader
	timeout     time.Duration
	backoff     time.Duration // time to wait after trying n servers when client cannot connect to Ios cluster
	random      bool          // if enabled, client connects to servers at random instead of systematically
	beforeForce int           // number of times a request should be submitted before client sets ForceViewChange, if -1 then will not set
	serverID    int           // ID of the server currently/last connected to
}

// connectRandom tries to connect to a server specified in addresses
func connectRandom(addrs []config.NetAddress, backoff time.Duration) (net.Conn, int) {
	for {
		for tried := 0; tried < len(addrs); tried++ {
			id := rand.Intn(len(addrs))
			glog.V(1).Info("Trying to connect to ", addrs[id].ToString())
			conn, err := net.Dial("tcp", addrs[id].ToString())
			// if successful
			if err == nil {
				glog.Infof("Connect established to %s", addrs[id].ToString())
				return conn, id
			}
			time.Sleep(backoff)
		}
	}
}

// connectSystematic try to establish a connection with server ID hint
// if unsuccessful, it tries to connect to other servers sytematically, waiting for backoff after trying each server
// once successful, connect will return the net.Conn and the ID of server connected to
// connectSystematic may never return if it cannot connect to any server
func connectSystematic(addrs []config.NetAddress, hint int, backoff time.Duration) (net.Conn, int) {
	// reset invalid hint
	if len(addrs) >= hint {
		hint = 0
	}

	// first, try on to connect to the most likely leader
	glog.Info("Trying to connect to ", addrs[hint].ToString())
	conn, err := net.Dial("tcp", addrs[hint].ToString())
	// if successful
	if err == nil {
		glog.Infof("Connect established to %s", addrs[hint].ToString())
		return conn, hint
	}
	glog.Warning(err) //if unsuccessful

	// if fails, try everyone else
	for {
		// TODO: start from hint instead of from ID:0
		for i := range addrs {
			glog.V(1).Info("Trying to connect to ", addrs[i].ToString())
			conn, err = net.Dial("tcp", addrs[i].ToString())

			// if successful
			if err == nil {
				glog.Infof("Connect established to %s", addrs[i].ToString())
				return conn, i
			}

			//if unsuccessful
			glog.Warning(err)
		}
		time.Sleep(backoff)
	}
}

// dispatcher will send bytes and wait for reply, return bytes returned if succussful or error otherwise
func dispatcher(b []byte, conn net.Conn, r *bufio.Reader, timeout time.Duration) ([]byte, error) {
	// check for nil connection
	if conn == nil {
		glog.Warning("connection missing")
		return nil, errors.New("Connection closed")
	}

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
		return nil, errors.New("Timeout of " + timeout.String())
	}
}

// StartClient creates an Ios client and tries to connect to an Ios cluster
// If ID is -1 then a random one will be generated
func StartClient(id int, statsFilename string, addrs []config.NetAddress, timeout time.Duration, backoff time.Duration, beforeForce int, random bool) (*Client, error) {

	// TODO: find a better way to handle required flags
	if id == -1 {
		rand.Seed(time.Now().UTC().UnixNano())
		id = rand.Int()
		glog.V(1).Info("ID was not provided, ID ", id, " has been assigned")
	}

	glog.Info("Starting up client ", id)

	// set up stats collection
	var stats *statsFile
	var err error
	if statsFilename != "" {
		stats, err = openStatsFile(statsFilename)
		if err != nil {
			return nil, err
		}
	}

	// connecting to server
	var conn net.Conn
	var serverID int
	if random {
		glog.Info("Client trying to connect to servers randomly")
		conn, serverID = connectRandom(addrs, backoff)
	} else {
		glog.Info("Client trying to connect to servers systematically")
		conn, serverID = connectSystematic(addrs, 0, backoff)
	}
	glog.Info("Client is ready to start processing incoming requests")

	rd := bufio.NewReader(conn)
	return &Client{id, 1, stats, addrs, conn, rd, timeout, backoff, random, beforeForce, serverID}, nil
}

func (c *Client) SubmitRequest(text string, readonly bool) (string, error) {
	glog.V(1).Info("Request ", c.requestID, " is: ", text)

	// prepare request
	req := msgs.ClientRequest{
		ClientID:        c.id,
		RequestID:       c.requestID,
		ForceViewChange: false,
		ReadOnly:        readonly,
		Request:         text}
	b, err := msgs.Marshal(req)
	if err != nil {
		glog.Warning(err)
		return "", err
	}
	glog.V(1).Info(string(b))

	if c.stats != nil {
			c.stats.startRequest(c.requestID)
	}
	tries := 0
	var reply *msgs.ClientResponse

	// dispatch request until successful
	for {
		if c.beforeForce != -1 && tries > c.beforeForce {
			glog.Warning("Request ", c.requestID, " is being set to force view change")
			req.ForceViewChange = true
			b, err = msgs.Marshal(req)
			if err != nil {
				glog.Warning(err)
				return "", err
			}
		}

		replyBytes, err := dispatcher(b, c.conn, c.rd, c.timeout)
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

		// continue if request failed
		glog.Warning("Request ", c.requestID, " failed due to: ", err)

		// close connection
		if c.conn != nil {
			err = c.conn.Close()
			if err != nil {
				glog.Warning(err)
			}
		}
		// establish a new connection
		if c.random {
			c.conn, c.serverID = connectRandom(c.servers, c.backoff)
		} else {
			// next try last serverID +1 mod n
			nextID := c.serverID + 1
			if nextID >= len(c.servers) {
				nextID = 0
			}
			c.conn, c.serverID = connectSystematic(c.servers, nextID, c.backoff)
		}
		c.rd = bufio.NewReader(c.conn)
		tries++
	}

	//check reply is not nil
	if *reply == (msgs.ClientResponse{}) {
		return "", errors.New("Response is nil")
	}

	//check reply is as expected
	if reply.ClientID != c.id {
		return "", errors.New("Response received has wrong ClientID")
	}
	if reply.RequestID != c.requestID {
		return "", errors.New("Response received has wrong RequestID")
	}
	if !reply.Success {
		return "", errors.New("Response marked as unsuccessful but not retried")
	}

	if c.stats != nil {
		// write to latency to log
		err = c.stats.stopRequest(tries, readonly)
	}
	c.requestID++
	return reply.Response, err
}

// StartClientFromConfigFile creates an Ios client
// If ID is -1 then a random one will be generated
func StartClientFromConfigFile(id int, statFile string, configFile string, addressFile string) (*Client, error) {
	conf, err := config.ParseClientConfig(configFile)
	if err != nil {
		return nil, err
	}
	if err := config.CheckConfig(conf); err != nil {
		return nil, err
	}
	addresses, err := config.ParseAddresses(addressFile)
	if err != nil {
		return nil, err
	}
	timeout := time.Millisecond * time.Duration(conf.Parameters.Timeout)
	backoff := time.Millisecond * time.Duration(conf.Parameters.Backoff)
	return StartClient(id, statFile, addresses.Clients, timeout, backoff, conf.Parameters.BeforeForce, conf.Parameters.ConnectRandom)
}

// StartClientFromConfig is the same as StartClientFromConfigFile but for config structs instead of files
func StartClientFromConfig(id int, statFile string, conf config.Config, addresses []config.NetAddress) (*Client, error) {
	if err := config.CheckConfig(conf); err != nil {
		return nil, err
	}
	timeout := time.Millisecond * time.Duration(conf.Parameters.Timeout)
	backoff := time.Millisecond * time.Duration(conf.Parameters.Backoff)
	return StartClient(id, statFile, addresses, timeout, backoff, conf.Parameters.BeforeForce, conf.Parameters.ConnectRandom)
}

func (c *Client) StopClient() {
	glog.Info("Shutting down client ", c.id)
	// close stats file
	if c.stats != nil {
		c.stats.closeStatsFile()
	}
	// close connection
	if c.conn != nil {
		c.conn.Close()
	}
}
