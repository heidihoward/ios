package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestParseAddresses calls ParseAddresses for the single server example config file
func TestParseParseAddresses(t *testing.T) {
	conf := ParseAddresses(os.Getenv("GOPATH") + "/src/github.com/heidi-ann/ios/ios/example.conf")
	addresses := conf.Clients.Address
	if len(addresses) != 1 && addresses[0] != "127.0.0.1:8080" {
		t.Fatal("Error parsing client address, parsed value: ", addresses)
	}

	peerAddresses := conf.Peers.Address
	if len(peerAddresses) != 1 && peerAddresses[0] != "127.0.0.1:8090" {
		t.Fatal("Error parsing peer address, parsed value: ", peerAddresses)
	}
}

// TestParseAddresses3 calls ParseAddresses for the 3 server example config file
func TestParseParseAddresses3(t *testing.T) {
	conf := ParseAddresses(os.Getenv("GOPATH") + "/src/github.com/heidi-ann/ios/ios/example3.conf")

	addresses := conf.Clients.Address
	assert.Equal(t, 3, len(addresses))
	assert.EqualValues(t, []string{"127.0.0.1:8080", "127.0.0.1:8081", "127.0.0.1:8082"}, addresses)

	peerAddresses := conf.Peers.Address
	assert.Equal(t, 3, len(peerAddresses))
	assert.EqualValues(t, []string{"127.0.0.1:8090", "127.0.0.1:8091", "127.0.0.1:8092"}, peerAddresses)

}
