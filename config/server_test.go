package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseServerConfig(t *testing.T) {
	ParseServerConfig(os.Getenv("GOPATH") + "/src/github.com/heidi-ann/ios/configfiles/delegated/server.conf")
	ParseServerConfig(os.Getenv("GOPATH") + "/src/github.com/heidi-ann/ios/configfiles/fpaxos/server.conf")
	ParseServerConfig(os.Getenv("GOPATH") + "/src/github.com/heidi-ann/ios/configfiles/simple/server.conf")
}


// TestParseServerConfig calls ParseServerConfig for the two example configuration files
func TestParseSingleServerConfig(t *testing.T) {
	conf := ParseServerConfig(os.Getenv("GOPATH") + "/src/github.com/heidi-ann/ios/configfiles/simple/server.conf")

	assert.Equal(t, 100000, conf.Performance.Length)
	assert.Equal(t, 0, conf.Performance.BatchInterval)
	assert.Equal(t, 10, conf.Performance.MaxBatch)
	assert.Equal(t, 0, conf.Algorithm.DelegateReplication)
	assert.Equal(t, 1, conf.Performance.WindowSize)
	assert.Equal(t, 1000, conf.Performance.SnapshotInterval)
	assert.Equal(t, "strict majority", conf.Algorithm.QuorumSystem)
	assert.Equal(t, false, conf.Algorithm.IndexExclusivity)
	assert.Equal(t, "kv-store", conf.Application.Name)

	assert.False(t, conf.Unsafe.DumpPersistentStorage)
	assert.Equal(t, "fsync", conf.Unsafe.PersistenceMode)

}
