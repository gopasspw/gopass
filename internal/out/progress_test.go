package out

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func ExampleProgressBar() {
	ctx := context.Background()
	max := 100
	pb := NewProgressBar(ctx, int64(max))
	for i := 0; i < max+20; i++ {
		pb.Inc()
		time.Sleep(150 * time.Millisecond)
	}
	time.Sleep(5 * time.Second)
	pb.Done()
}

func TestProgress(t *testing.T) {
	ctx := context.Background()
	max := 2
	pb := NewProgressBar(WithHidden(ctx, true), int64(max))
	pb.Inc()
	assert.Equal(t, int64(1), pb.current)
}
