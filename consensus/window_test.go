package consensus

import (
	"testing"
)

func TestNextIndex(t *testing.T) {
	window := newReplicationWindow(-1, 1)
	index := window.nextIndex()
	if index != 0 {
		t.Error("ReplicationWindow giving wrong index")
	}
}
