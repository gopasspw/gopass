package termio

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func ExampleProgressBar() { //nolint:testableexamples
	maxVal := 100
	pb := NewProgressBar(int64(maxVal))

	for range maxVal + 20 {
		pb.Inc()
		pb.Add(23)
		pb.Set(42)
		time.Sleep(150 * time.Millisecond)
	}

	time.Sleep(5 * time.Second)
	pb.Done()
}

func TestProgress(t *testing.T) {
	maxVal := 2
	pb := NewProgressBar(int64(maxVal))
	pb.Hidden = true
	pb.Inc()
	assert.Equal(t, int64(1), pb.current)
}

func TestProgressNil(t *testing.T) {
	t.Parallel()

	var pb *ProgressBar
	pb.Inc()
	pb.Add(4)
	pb.Done()
}

func TestProgressBytes(t *testing.T) {
	maxSize := 2 << 24
	pb := NewProgressBar(int64(maxSize))
	pb.Hidden = true
	pb.Bytes = true

	for i := range 24 {
		pb.Set(2 << (i + 1))
	}

	assert.Equal(t, int64(maxSize), pb.current)
	pb.Done()
}
