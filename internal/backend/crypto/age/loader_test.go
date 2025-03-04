package age

import (
	"context"
	"testing"

	"github.com/gopasspw/gopass/internal/store/mockstore/inmem"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoader_New(t *testing.T) {
	ctx := context.Background()
	l := loader{}

	crypto, err := l.New(ctx)
	require.NoError(t, err)
	assert.NotNil(t, crypto)
}

func TestLoader_Handles(t *testing.T) {
	ctx := context.Background()
	l := loader{}
	s := inmem.New()

	// Test case where OldIDFile or OldKeyring exists
	require.NoError(t, s.Set(ctx, OldIDFile, []byte("test")))
	err := l.Handles(ctx, s)
	require.NoError(t, err)
	s.Delete(ctx, OldIDFile)

	// Test case where IDFile exists
	require.NoError(t, s.Set(ctx, OldIDFile, []byte("test")))
	require.NoError(t, s.Set(ctx, IDFile, []byte("test")))
	err = l.Handles(ctx, s)
	require.NoError(t, err)
	require.NoError(t, s.Delete(ctx, OldIDFile))
	require.NoError(t, s.Delete(ctx, IDFile))

	// Test case where IDFile exists
	require.NoError(t, s.Set(ctx, IDFile, []byte("test")))
	err = l.Handles(ctx, s)
	require.NoError(t, err)
	require.NoError(t, s.Delete(ctx, IDFile))

	// Test case where neither OldIDFile nor IDFile exists
	err = l.Handles(ctx, s)
	assert.Error(t, err)
}

func TestLoader_Priority(t *testing.T) {
	l := loader{}
	assert.Equal(t, 10, l.Priority())
}

func TestLoader_String(t *testing.T) {
	l := loader{}
	assert.Equal(t, name, l.String())
}
