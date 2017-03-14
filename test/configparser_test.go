package test

import (
	"os"
	"testing"
)

// TestParseAuto calls ParseAuto for the example configuration file
func TestParseAuto(t *testing.T) {
	ParseAuto(os.Getenv("GOPATH") + "/src/github.com/heidi-ann/ios/test/workload.conf")
}
