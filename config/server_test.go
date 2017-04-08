package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestParseServerConfig calls ParseServerConfig for the two example configuration files
func TestParseSingleServerConfig(t *testing.T) {
	conf := ParseServerConfig(os.Getenv("GOPATH") + "/src/github.com/heidi-ann/ios/configfiles/server.conf")
	addresses := conf.Clients.Address
	if len(addresses) != 1 && addresses[0] != "127.0.0.1:8080" {
		t.Fatal("Error parsing client address, parsed value: ", addresses)
	}

	peerAddresses := conf.Peers.Address
	if len(peerAddresses) != 1 && peerAddresses[0] != "127.0.0.1:8090" {
		t.Fatal("Error parsing peer address, parsed value: ", peerAddresses)
	}

	options := conf.Options
	assert.Equal(t, 100000, options.Length)
	assert.Equal(t, 5, options.BatchInterval)
	assert.Equal(t, 100, options.MaxBatch)
	assert.Equal(t, -1, options.DelegateReplication)
	assert.Equal(t, 1, options.WindowSize)
	assert.Equal(t, 100, options.SnapshotInterval)
	assert.Equal(t, "strict majority", options.QuorumSystem)
	assert.Equal(t, true, options.IndexExclusivity)
	assert.Equal(t, "kv-store", options.Application)

	unsafe := conf.Unsafe
	assert.False(t, unsafe.DumpPersistentStorage)
	assert.Equal(t, "fsync", unsafe.PersistenceMode)

}

func TestParseMultiServerConfig(t *testing.T) {
	conf := ParseServerConfig(os.Getenv("GOPATH") + "/src/github.com/heidi-ann/ios/configfiles/server3.conf")

	addresses := conf.Clients.Address
	assert.Equal(t, 3, len(addresses))
	assert.EqualValues(t, []string{"127.0.0.1:8080", "127.0.0.1:8081", "127.0.0.1:8082"}, addresses)

	peerAddresses := conf.Peers.Address
	assert.Equal(t, 3, len(peerAddresses))
	assert.EqualValues(t, []string{"127.0.0.1:8090", "127.0.0.1:8091", "127.0.0.1:8092"}, peerAddresses)

	options := conf.Options
	assert.Equal(t, 100000, options.Length)
	assert.Equal(t, 5, options.BatchInterval)
	assert.Equal(t, 100, options.MaxBatch)
	assert.Equal(t, -1, options.DelegateReplication)
	assert.Equal(t, 1, options.WindowSize)
	assert.Equal(t, 100, options.SnapshotInterval)
	assert.Equal(t, "strict majority", options.QuorumSystem)
	assert.Equal(t, true, options.IndexExclusivity)
	assert.Equal(t, "kv-store", options.Application)

	unsafe := conf.Unsafe
	assert.False(t, unsafe.DumpPersistentStorage)
	assert.Equal(t, "fsync", unsafe.PersistenceMode)
}
