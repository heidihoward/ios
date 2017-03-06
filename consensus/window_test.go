package consensus

import (
	"testing"
)

func TestNextIndex(t *testing.T) {
  window := NewReplicationWindow(-1,1)
  index := window.NextIndex()
  if index != 0 {
    t.Error("ReplicationWindow giving wrong index")
  }
}
