![Ios project logo](../master/misc/logo.png?raw=true)


[![Build Status](https://travis-ci.org/heidi-ann/ios.svg?branch=master)](https://travis-ci.org/heidi-ann/ios)
[![Go Report Card](https://goreportcard.com/badge/github.com/heidi-ann/ios)](https://goreportcard.com/report/github.com/heidi-ann/ios)
[![GoDoc](https://godoc.org/github.com/heidi-ann/ios?status.svg)](https://godoc.org/github.com/heidi-ann/ios)
[![Coverage Status](https://coveralls.io/repos/github/heidi-ann/ios/badge.svg?branch=master)](https://coveralls.io/github/heidi-ann/ios?branch=master)

Welcome to Ios, a reliable distributed agreement service for cloud applications. Built upon a novel decentralised consensus protocol, Ios provides vital services for your cloud application such as distributed locking, consistent data structures and leader election as well as distributed configuration and coordination.

*This repository is pre-alpha and under active development. APIs will be broken. This code has not been proven correct and is not ready for production deployment.*

## Getting Started
These instructions will get you a simple Ios server and client up and running on your local machine. See deployment for notes on how to deploy Ios across a cluster.

### Prerequisites
Ios is built in [Go](https://golang.org/) version 1.6.2 and currently supports Go version 1.6 and 1.7. The [Golang site](https://golang.org/) details how to install and setup Go on your local machine. Don't forget to add GOPATH to your .profile.

### Installation
After installing Go, run:
```
go get github.com/heidi-ann/ios/...
```
This command will copy the Ios source code to $GOPATH/src/github.com/heidi-ann/ios and then fetch and build the following dependancies:
* [glog](github.com/golang/glog) - logging library, in the style of glog for C++
* [gcfg](gopkg.in/gcfg.v1) - library for parsing git-config style config files
It will then build and install Ios, the server and client binaries will be placed in $GOPATH/bin.

### Up & Running
You can now start a simple 1 node Ios cluster as follows:
```
$GOPATH/bin/ios -id 0
```
This will start an Ios server providing a simple key-value store. The server is listening for clients on port 8080.

You can now start an Ios client as follows:
```
$ $GOPATH/bin/clientcli
Starting Ios client in interactive mode.

The following commands are available:
	get [key]: to return the value of a given key
	exists [key]: to test if a given key is present
	update [key] [value]: to set the value of a given key, if key already exists then overwrite
	delete [key]: to remove a key value pair if present
	count: to return the number of keys
	print: to return all key value pairs

Enter command: update A 1
OK
Enter command: get A
1
...
```
You can now enter commands for the key value store, followed by the enter key. These commands are being sent to the Ios server, executed and the result is returned to the user.

The Ios server is using files called persistent_log_0.temp, persistent_snap_0.temp and persistent_data_0.temp to store Ios's persistent state. If these files are present when a server starts, it will restore its state from these files. You can try this by killing the server process and restarting it, it should carry on from where it left off.

When you would like to start a fresh server instance, use ``rm persistent*.temp`` first to clear these files and then start the server again.

## Building in Docker

Alternatively you can build and run in docker. Make sure Docker is installed and running, then clone this repository.

Build an image named 'ios' using the command

```
docker build -t ios .
```

You can then run server instances in docker passing configuration through directly (be sure to expose the ports from the container). E.g:

```
docker run -p 8080:8080 ios -id 0
```

Note that this will only use storage local to the container instance. If you want persistence/recoverability for instances you will need to store persistence logs on a mounted data volume

## Next steps

In this section, we are going to take a closer look at what is going on underneath. We will then use this information to setup a 3 server Ios cluster on your local machine and automatically generate a workload to put it to the test. PS: you might want to start by opening up a few terminal windows.

#### Server configuration
The server we ran in previous section was using the default configuration file found in [example.conf](example.conf). The first section of this file lists the Ios servers in the cluster and how the peers can connect to them and the second section lists how the client can connect to them. The configuration file [example3.conf](example3.conf) shows what this looks like for 3 servers running on localhost. The same configuration file is used for all the servers, at run time they are each given an ID (starting from 0) and use this to know which ports to listen on. The rest of the configuration file options are documented at https://godoc.org/github.com/heidi-ann/ios/config. After removing the persistent storage, start 3 Ios servers in 3 separate terminal windows as follows:

```
$GOPATH/bin/ios -id [ID] -config $GOPATH/src/github.com/heidi-ann/ios/configfiles/simple/server3.conf -stderrthreshold=INFO
```
For ID 0, 1 and 2

#### Client configuration

Like the servers, the client we ran in the previous section was using the default configuration file found in [example.conf](example.conf). The first section lists the Ios servers in the cluster and how to connect to them. The configuration file [example3.conf](example3.conf) shows what this looks like for 3 servers currently running on localhost.

We are run a client as before and interact with our 3 servers.
```
$GOPATH/bin/clientcli -config $GOPATH/src/github.com/heidi-ann/ios/client/example3.conf
```

You should be able to kill and restart the servers to test when the system is available to the client. Since the Ios cluster you have deployed is configured to use strict majority quorums then the system should be available whenever at least two servers are up.

#### Workload configuration

Typing requests into a terminal is, of course, slow and unrealistic. To help test the system, Ios provides test clients which can automatically generate a workload and measure system performance. To run a client in test mode use:
```
$GOPATH/bin/test -config $GOPATH/src/github.com/heidi-ann/ios/client/example3.conf -auto $GOPATH/src/github.com/heidi-ann/ios/test/workload.conf
```
This client will run the workload described in [test/workloads/example.conf](test/workloads/example.conf) and then terminate. It will write performance metrics into a file called latency.csv. Ios currently also support a REST API mode which listens for HTTP on port 12345.

## Contributing

#### Debugging

We use glog for logging. Adding `-logtostderr=true -v=1` when running executables prints the logging output. For more information, visit https://godoc.org/github.com/golang/glog.

Likewise, the following commands work with the above example and are useful for debugging:
```
sudo tcpdump -i lo0 -nnAS "(src portrange 8080-8092 or dst portrange 8080-8092) and (length>0)"
```
```
sudo strace -p $(pidof server) -T -e fsync -f
sudo strace -p $(pidof server) -T -e trace=write -f
```

### Benchmarking

The benchmarking scripts for Ios can found here https://github.com/heidi-ann/consensus_eval

## License

This project is licensed under the MIT License - see the [LICENSE.md](LICENSE.md) file for details
