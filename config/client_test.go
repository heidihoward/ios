package config

import (
	"os"
	"testing"
)

// TestParseClientConfig calls ParseServerConfig for the two example configuration files
func TestParseClientConfig(t *testing.T) {
	ParseClientConfig(os.Getenv("GOPATH") + "/src/github.com/heidi-ann/ios/configfiles/client.conf")
	ParseClientConfig(os.Getenv("GOPATH") + "/src/github.com/heidi-ann/ios/configfiles/client3.conf")
}
