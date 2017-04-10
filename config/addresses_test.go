package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestParseAddresses calls ParseAddresses for the single server example config file
func TestParseParseAddresses(t *testing.T) {
	conf := ParseAddresses(os.Getenv("GOPATH") + "/src/github.com/heidi-ann/ios/ios/example.conf")

	assert.Equal(t, 1, len(conf.Clients))
	assert.Equal(t, "127.0.0.1", conf.Clients[0].Address)
	assert.Equal(t, 8080, conf.Clients[0].Port)

	assert.Equal(t, 1, len(conf.Peers))
	assert.Equal(t, "127.0.0.1", conf.Peers[0].Address)
	assert.Equal(t, 8090, conf.Peers[0].Port)
}

// TestParseAddresses3 calls ParseAddresses for the 3 server example config file
func TestParseParseAddresses3(t *testing.T) {
	conf := ParseAddresses(os.Getenv("GOPATH") + "/src/github.com/heidi-ann/ios/ios/example3.conf")

	assert.Equal(t, 3, len(conf.Clients))
	for i := 0; i < 3; i++ {
		assert.Equal(t, "127.0.0.1", conf.Clients[i].Address)
		assert.Equal(t, 8080+i, conf.Clients[i].Port)
	}

	assert.Equal(t, 3, len(conf.Peers))
	for i := 0; i < 3; i++ {
		assert.Equal(t, "127.0.0.1", conf.Peers[i].Address)
		assert.Equal(t, 8090+i, conf.Peers[i].Port)
	}
}
