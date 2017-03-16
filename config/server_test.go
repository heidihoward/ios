package config

import (
	"os"
	"testing"
)

// TestParseServerConfig calls ParseServerConfig for the two example configuration files
func TestParseServerConfig(t *testing.T) {
	ParseServerConfig(os.Getenv("GOPATH") + "/src/github.com/heidi-ann/ios/ios/example.conf")
	ParseServerConfig(os.Getenv("GOPATH") + "/src/github.com/heidi-ann/ios/ios/example3.conf")
}
