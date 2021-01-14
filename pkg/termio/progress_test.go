package termio

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func ExampleProgressBar() {
	max := 100
	pb := NewProgressBar(int64(max), false)
	for i := 0; i < max+20; i++ {
		pb.Inc()
		time.Sleep(150 * time.Millisecond)
	}
	time.Sleep(5 * time.Second)
	pb.Done()
}

func TestProgress(t *testing.T) {
	max := 2
	pb := NewProgressBar(int64(max), true)
	pb.Inc()
	assert.Equal(t, int64(1), pb.current)
}
