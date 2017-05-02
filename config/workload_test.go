package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestParseAuto calls ParseAuto for the example configuration file
func TestParseAuto(t *testing.T) {
	files := []string{"example.conf", "balanced.conf", "read-heavy.conf", "write-heavy.conf"}
	for _, file := range files {
		workload, err := ParseWorkloadConfig(os.Getenv("GOPATH") + "/src/github.com/heidi-ann/ios/test/workloads/" + file)
		assert.Nil(t, err)
		assert.Nil(t, CheckWorkloadConfig(workload))
	}
}
