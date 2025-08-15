package age

import (
	"testing"

	"github.com/gopasspw/gopass/pkg/debug"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	ctx := t.Context()
	a, err := New(ctx, "")
	require.NoError(t, err)
	assert.NotNil(t, a)
}

func TestInitialized(t *testing.T) {
	ctx := t.Context()
	a, err := New(ctx, "")
	require.NoError(t, err)
	assert.NotNil(t, a)

	err = a.Initialized(ctx)
	require.NoError(t, err)
}

func TestName(t *testing.T) {
	ctx := t.Context()
	a, err := New(ctx, "")
	require.NoError(t, err)
	assert.NotNil(t, a)

	name := a.Name()
	assert.Equal(t, "age", name)
}

func TestVersion(t *testing.T) {
	ctx := t.Context()
	a, err := New(ctx, "")
	require.NoError(t, err)
	assert.NotNil(t, a)

	version := a.Version(ctx)
	expectedVersion := debug.ModuleVersion("filippo.io/age")
	assert.Equal(t, expectedVersion, version)
}

func TestExt(t *testing.T) {
	ctx := t.Context()
	a, err := New(ctx, "")
	require.NoError(t, err)
	assert.NotNil(t, a)

	ext := a.Ext()
	assert.Equal(t, Ext, ext)
}

func TestIDFile(t *testing.T) {
	ctx := t.Context()
	a, err := New(ctx, "")
	require.NoError(t, err)
	assert.NotNil(t, a)

	idFile := a.IDFile()
	assert.Equal(t, IDFile, idFile)
}

func TestConcurrency(t *testing.T) {
	ctx := t.Context()
	a, err := New(ctx, "")
	require.NoError(t, err)
	assert.NotNil(t, a)

	concurrency := a.Concurrency()
	assert.Equal(t, 1, concurrency)
}
