package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestParseServerConfig calls ParseServerConfig for the two example configuration files
func TestParseSingleServerConfig(t *testing.T) {
	conf := ParseServerConfig(os.Getenv("GOPATH") + "/src/github.com/heidi-ann/ios/configfiles/simple/server.conf")
	
	options := conf.Options
	assert.Equal(t, 100000, options.Length)
	assert.Equal(t, 5, options.BatchInterval)
	assert.Equal(t, 100, options.MaxBatch)
	assert.Equal(t, 0, options.DelegateReplication)
	assert.Equal(t, 1, options.WindowSize)
	assert.Equal(t, 1000, options.SnapshotInterval)
	assert.Equal(t, "strict majority", options.QuorumSystem)
	assert.Equal(t, true, options.IndexExclusivity)
	assert.Equal(t, "kv-store", options.Application)

	unsafe := conf.Unsafe
	assert.False(t, unsafe.DumpPersistentStorage)
	assert.Equal(t, "fsync", unsafe.PersistenceMode)

}
