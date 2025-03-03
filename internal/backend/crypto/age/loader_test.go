package age

import (
	"context"
	"testing"

	"github.com/gopasspw/gopass/internal/backend/mock"
	"github.com/stretchr/testify/assert"
)

func TestLoader_New(t *testing.T) {
	ctx := context.Background()
	l := loader{}

	crypto, err := l.New(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, crypto)
}

func TestLoader_Handles(t *testing.T) {
	ctx := context.Background()
	l := loader{}
	s := mock.NewMockStorage()

	// Test case where OldIDFile or OldKeyring exists
	s.SetExists(OldIDFile, true)
	err := l.Handles(ctx, s)
	assert.NoError(t, err)

	// Test case where IDFile exists
	s.SetExists(OldIDFile, false)
	s.SetExists(IDFile, true)
	err = l.Handles(ctx, s)
	assert.NoError(t, err)

	// Test case where neither OldIDFile nor IDFile exists
	s.SetExists(IDFile, false)
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
