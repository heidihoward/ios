package config

import (
	"github.com/golang/glog"
	"gopkg.in/gcfg.v1"
  "strings"
  "strconv"
)

// Addresses describes the network configuarion of an Ios cluster.
// Peers holds the addresses of all Ios servers on which peers can connect to
// Clients holds the addresses of all Ios servers on which clients can connect to
// Address is of the form ipv4:port e.g. 127.0.0.1:8090

type Addresses struct {
	Peers struct {
		Address []string
	}
	Clients struct {
		Address []string
	}
}

// ParseAddresses filename will parse the given file into an Addresses struct
func ParseAddresses(filename string) Addresses {
	var config Addresses
	err := gcfg.ReadFileInto(&config, filename)
	if err != nil {
		glog.Fatalf("Failed to parse gcfg data: %s", err)
	}
  // checking configuation is sensible
  if len(config.Peers.Address) == 0 {
    glog.Fatal("At least one server is required")
  }
  if len(config.Peers.Address) != len(config.Clients.Address) {
    glog.Fatal("Same number of servers is required in the Peer and Client sections")
  }
  for _,addr := range config.Peers.Address {
    address := strings.Split(addr,":")
    if len(address) != 2 {
      glog.Fatal("Address must be of the form, address:port e.g. 127.0.0.1:8090 ")
    }
    port, err := strconv.Atoi(address[1])
    if err != nil || port < 0 || port > 65535 {
      glog.Fatal("Address must be of the form, address:port e.g. 127.0.0.1:8090 ")
    }
    // TODO: check format of IP address/domain name
  }
	return config
}
