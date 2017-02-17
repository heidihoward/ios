![Ios project logo](../master/misc/logo.png?raw=true)


[![Build Status](https://travis-ci.org/heidi-ann/ios.svg?branch=master)](https://travis-ci.org/heidi-ann/ios)
[![Go Report Card](https://goreportcard.com/badge/github.com/heidi-ann/ios)](https://goreportcard.com/report/github.com/heidi-ann/ios)
[![GoDoc](https://godoc.org/github.com/heidi-ann/ios?status.svg)](https://godoc.org/github.com/heidi-ann/ios)

Welcome to Ios, a distributed and strongly consistent key-value store, built upon a novel delegated and decentralised consensus protocol.

*This repository is pre-alpha and under active development. APIs will be broken.*


### Installation

Most of this project is written using Go version 1.6.2 The [Go lang site](https://golang.org/) details how to install and setup Go. Don't forget to add GOPATH to your .profile. The project has the following dependancies:
* [glog](github.com/golang/glog) - logging library, in the style of glog for C++
* [gcfg](gopkg.in/gcfg.v1) - library for parsing git-config style config files

After installing go:
```
go get github.com/heidi-ann/ios/...
```

### Quick start
You can start a 1 node Ios cluster as follows:
```
cd $GOPATH/src/github.com/heidi-ann/ios/server
$GOPATH/bin/server -id 0 -config example.conf -logtostderr true
```
This will start an Ios server with ID 0, clients can now communicate with the server over port 8080 as follows:
```
$ cd $GOPATH/src/github.com/heidi-ann/ios/client
$ $GOPATH/bin/client -id=0 -config example.conf
Starting Ios client in interactive mode.

The following commands are available:
	get [key]: to return the value of a given key
	update [key] [value]: to set the value of a given key

Enter command: update A 1
OK
Enter command: get A
1
...
```
The server is using files called persistent_log_0.temp and persistent_data_0.temp to store Ios's persistent state. If these files are present when the server starts, it will restore the state from these files, if you would like to start a fresh server, make sure to use ``rm *.temp`` first.

### Usage

#### Client

The (mode independent) client state is stored in the example.conf file. The client has three possible interfaces:
* Test - a workload is automatically generated for Ios. This workload is configured using a workload.conf file. An example of this is given in test/workload.conf.
* Interactive - requests are entered from the terminal. Requests takes the form of either ``get [key]`` or ``update [key] [value]``. There can be multiple commands in a single request, separated by semi-colons
* REST API - a HTTP server on port 12345

Each client needs a unique id.

#### Logging

We use glog for logging. Adding `-logtostderr=true` when running executables prints the logging output. For more information, visit https://godoc.org/github.com/golang/glog.

Likewise, the following works with the above example and is useful for debugging:
```
sudo tcpdump -i lo0 -nnAS "(src portrange 8080-8092 or dst portrange 8080-8092) and (length>0)"
```

### Benchmarking

The benchmarking scripts for Ios can found here https://github.com/heidi-ann/consensus_eval
