package interval_test

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thrawn01/queue-patterns.go/interval"
	"testing"
	"time"
)

func TestInterval(t *testing.T) {
	i := interval.NewInterval(10 * time.Millisecond)
	defer i.Stop()
	i.Next()

	assert.Empty(t, i.C)

	time.Sleep(10 * time.Millisecond)

	// Wait for tick.
	select {
	case <-i.C:
	case <-time.After(100 * time.Millisecond):
		require.Fail(t, "timeout")
	}

	i.Next()
	time.Sleep(10 * time.Millisecond)

	// Wait for tick.
	select {
	case <-i.C:
	case <-time.After(100 * time.Millisecond):
		require.Fail(t, "timeout")
	}
}
