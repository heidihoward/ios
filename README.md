# Hydra
Geo-replicated and strongly consistent key-value store

### Installation

Most of this project is written using Go version 1.5.3. The [Go lang site](https://golang.org/) details how to install and setup go. Don't forget to add GOPATH to your .profile. The project has the following dependancies:
* [glog](github.com/golang/glog) - logging library, in the style of glog for C++
* [gcfg](gopkg.in/gcfg.v1) - library for parsing git-config style config files

After install go:
```
go get github.com/golang/glog
go get gopkg.in/gcfg.v1
go get github.com/heidi-ann/hydra

cd $GOPATH/github.com/heidi-ann/hydra

cd server/
go install
cd ../client
go install
```

### Usage 

#### Server
To start a Hydra server on port 8080:
```
$GOPATH/bin/server -port 8080 -logtostderr true
```
If port is not specificed it will default to 8080. The server is use a file called persistent.log to store a perisitent copy of the history of requests. If you would like to start a fresh server, remove the local persistent.log file first.

#### Client
The (mode independent) client state is stored in the example.conf file. The client has three possible interfaces:
* Test - a workload is auotmatically generated for hydra. This workload is configuated using a workload.conf file. An example of this is given in test/workload.conf.
* Interactive - requests are entered from the terminal. Requests takes the form of get A or update A B. There can be multiple commands in a single request, seperated by semi-colons
* REST API - a http server on port 12345

#### Logging 

We use glog for logging. Adding `-logtostderr=true` when running executables prints the logging output. For more information, visit https://godoc.org/github.com/golang/glog