package cli

import (
	"context"
	"testing"

	"github.com/gopasspw/gopass/pkg/termio"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	client := New()
	assert.NotNil(t, client)
	assert.False(t, client.repeat)
}

func TestSet(t *testing.T) {
	client := New()

	err := client.Set("REPEAT")
	assert.NoError(t, err)
	assert.True(t, client.repeat)

	err = client.Set("OTHER")
	assert.NoError(t, err)
	assert.True(t, client.repeat)
}

func TestOption(t *testing.T) {
	client := New()

	err := client.Option("ANY")
	assert.NoError(t, err)
}

func TestGetPIN(t *testing.T) {
	client := New()

	ctx := termio.WithPassPromptFunc(context.Background(), func(ctx context.Context, s string) (string, error) {
		return "1234", nil
	})

	pin, err := client.GetPINContext(ctx)
	assert.NoError(t, err)
	assert.Equal(t, "1234", pin)
}
