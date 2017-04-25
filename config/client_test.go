package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestParseClientConfig calls ParseConfig for the two example configuration files
func TestParseClientConfig(t *testing.T) {
	conf, err := ParseClientConfig(os.Getenv("GOPATH") + "/src/github.com/heidi-ann/ios/configfiles/simple/client.conf")
	assert.Nil(t, err)
	assert.Nil(t, CheckConfig(conf))
	conf, err = ParseClientConfig(os.Getenv("GOPATH") + "/src/github.com/heidi-ann/ios/configfiles/fpaxos/client.conf")
	assert.Nil(t, err)
	assert.Nil(t, CheckConfig(conf))
	conf, err = ParseClientConfig(os.Getenv("GOPATH") + "/src/github.com/heidi-ann/ios/configfiles/delegated/client.conf")
	assert.Nil(t, err)
	assert.Nil(t, CheckConfig(conf))
}
