package msgs

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestFailureNotifier(t *testing.T) {
	assert := assert.New(t)
	fn := NewFailureNotifier(5)

	for id := 0; id < 5; id++ {
		assert.False(fn.IsConnected(id), "Node should be initially disconnected")
		err := fn.NowConnected(id)
		assert.Nil(err, "Node could not connect")
		assert.True(fn.IsConnected(id), "Node should be connected")
	}

	// check on false failures
	select {
	case <- fn.NotifyOnFailure(3):
		t.Error("Unexpected failure")
	case <- time.After(100 *time.Millisecond):
	}

	wait := fn.NotifyOnFailure(3)
	fn.NowDisconnected(3)

	// check on false failures
	select {
	case <-wait:
	case <- time.After(100 *time.Millisecond):
		t.Error("Failure not reported")
	}

}
