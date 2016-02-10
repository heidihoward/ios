# Hydra
Geo-replicated and strongly consistent key-value store

### Installation

Most of this project is written using Go version 1.5.3. The [Go lang site](https://golang.org/) details how to install and setup go. Don't forget to add GOPATH to your .profile. The project has the following dependancies:
* [glog](github.com/golang/glog)

### Usage 

#### Logging 

We use glog for logging. Adding `-logtostderr=true` when running executables prints the logging output. For more information, visit https://godoc.org/github.com/golang/glog