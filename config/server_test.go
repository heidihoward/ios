package config

import (
	"testing"
  "os"
)

// TestParseServerConfig calls ParseServerConfig for the two example configuration files
func TestParseServerConfig(t *testing.T) {
  ParseServerConfig(os.Getenv("GOPATH")+"/src/github.com/heidi-ann/ios/server/example.conf")
  ParseServerConfig(os.Getenv("GOPATH")+"/src/github.com/heidi-ann/ios/server/example3.conf")
}
