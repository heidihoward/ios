package config

import (
	"github.com/golang/glog"
	"gopkg.in/gcfg.v1"
	"strconv"
	"strings"
)

// AddressFile describes the addresses as parsed by the address file
type AddressFile struct {
	Servers struct {
		Address []string
	}
}

// NetAddress holds a network address
type NetAddress struct {
	Address string
	Port    int
}

func (n NetAddress) ToString() string {
	return n.Address + ":" + strconv.Itoa(n.Port)
}

// Addresses describes the network configuarion of an Ios cluster.
type Addresses struct {
	Peers   []NetAddress
	Clients []NetAddress
}

// ParseAddresses filename will parse the given file into an Addresses struct
func ParseAddresses(filename string) Addresses {
	var config AddressFile
	err := gcfg.ReadFileInto(&config, filename)
	if err != nil {
		glog.Fatalf("Failed to parse gcfg data: %s", err)
	}
	// checking configuation is sensible
	n := len(config.Servers.Address)
	if n == 0 {
		glog.Fatal("At least one server is required")
	}
	// parse into addresses
	addresses := Addresses{make([]NetAddress, n), make([]NetAddress, n)}

	for i, addr := range config.Servers.Address {
		address := strings.Split(addr, ":")
		if len(address) != 3 {
			glog.Fatal("Address must be of the form, ipv4:serverport:clientport e.g. 127.0.0.1:8090:8080 ")
		}
		// TODO: check format of IP address/domain name

		// parse peer port
		peerPort, err := strconv.Atoi(address[1])
		if err != nil || peerPort < 0 || peerPort > 65535 {
			glog.Fatal("Address must be of the form, ipv4:serverport:clientport e.g. 127.0.0.1:8090:8080 ")
		}
		addresses.Peers[i] = NetAddress{address[0], peerPort}

		// parse client port
		clientPort, err := strconv.Atoi(address[2])
		if err != nil || clientPort < 0 || clientPort > 65535 {
			glog.Fatal("Address must be of the form, ipv4:serverport:clientport e.g. 127.0.0.1:8090:8080 ")
		}
		addresses.Clients[i] = NetAddress{address[0], clientPort}

	}
	return addresses
}
