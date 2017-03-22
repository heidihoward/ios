package msgs

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestFailureNotifier(t *testing.T) {
	assert := assert.New(t)
	fn := NewFailureNotifier(5)

	for id := 0; id < 5; id++ {
		assert.False(fn.IsConnected(id), "Node should be initially disconnected")
		err := fn.NowConnected(id)
		assert.Nil(err, "Node could not connect")
		assert.True(fn.IsConnected(id), "Node should be initially disconnected")
	}
}
