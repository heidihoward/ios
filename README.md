![Ios project logo](../master/misc/logo.png?raw=true)

Welcome to Ios, a distributed and strongly consistent key-value store, built on the Ios distributed consensus protocol. 

### Installation

Most of this project is written using Go version 1.5.3. The [Go lang site](https://golang.org/) details how to install and setup Go. Don't forget to add GOPATH to your .profile. The project has the following dependancies:
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
To start a Hydra server:
```
$GOPATH/bin/server -id 0 -logtostderr true
```
This will start an Hydra server with ID 0, clients can now communicate with the server over port 8080 and other servers can communicate with this server over port 8090. You can modfiy these ports as follows:
```
$GOPATH/bin/server -id 1 -client-port 8081  -peer-port 8091 -logtostderr true
```


The server is using files called persistent_log_1.temp and persistent_data_1.temp to store a perisitent copy hydra's state. If you would like to start a fresh server, make sure to use rm *.temp first.

#### Client
The (mode independent) client state is stored in the example.conf file. The client has three possible interfaces:
* Test - a workload is auotmatically generated for hydra. This workload is configuated using a workload.conf file. An example of this is given in test/workload.conf.
* Interactive - requests are entered from the terminal. Requests takes the form of get A or update A B. There can be multiple commands in a single request, seperated by semi-colons
* REST API - a http server on port 12345
Each client needs a unique id.

#### Logging 

We use glog for logging. Adding `-logtostderr=true` when running executables prints the logging output. For more information, visit https://godoc.org/github.com/golang/glog.

Likewise, the following works with the above example and is useful for debugging:
```
sudo tcpdump -i lo0 -nnAS "(src portrange 8080-8092 or dst portrange 8080-8092) and (length>0)"
```

### Benchmarking

The benchmarking scripts are as follows:
- `all_test.sh` - run all tests
- `batch_testing.sh` - enabling batching to test how it improves throughput
- `failure_testing.sh` - simulating failure by kill processes (NOT WORKING)
- `load_testing.sh` - simulating load using max rate clients
- `major_failure_testing.sh` - simulating failure by kill the process of the master
- `read_testing.sh` - changing read/write ratio of workload
- `scale_testing.sh` - varying clusters size
- `simple_testing.sh` - quick test of a simple system (START HERE)

The ipython3 notebook `graph_plotting.iynp` plots the outputs of the above tests