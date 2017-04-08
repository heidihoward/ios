package config

import (
	"os"
	"testing"
)

// TestParseAuto calls ParseAuto for the example configuration file
func TestParseAuto(t *testing.T) {
	ParseWorkloadConfig(os.Getenv("GOPATH") + "/src/github.com/heidi-ann/ios/configfiles/workload.conf")
}
