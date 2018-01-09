package sub

import (
	"context"
	"io/ioutil"
	"os"
	"testing"

	"github.com/fatih/color"
	gpgmock "github.com/justwatchcom/gopass/backend/gpg/mock"
	"github.com/stretchr/testify/assert"
)

func TestTemplates(t *testing.T) {
	ctx := context.Background()

	tempdir, err := ioutil.TempDir("", "gopass-")
	assert.NoError(t, err)
	defer func() {
		_ = os.RemoveAll(tempdir)
	}()

	color.NoColor = true

	_, _, err = createStore(tempdir, nil, nil)
	assert.NoError(t, err)

	s, err := New(
		"",
		tempdir,
		gpgmock.New(),
	)
	assert.NoError(t, err)

	assert.Equal(t, 0, len(s.ListTemplates(ctx, "")))
	assert.NoError(t, s.SetTemplate("foo", []byte("foobar")))
	assert.Equal(t, 1, len(s.ListTemplates(ctx, "")))

	tt, err := s.TemplateTree(ctx)
	assert.NoError(t, err)
	assert.Equal(t, "gopass\n└── foo\n", tt.Format(0))

	assert.Equal(t, true, s.HasTemplate("foo"))

	b, err := s.GetTemplate("foo")
	assert.NoError(t, err)
	assert.Equal(t, "foobar", string(b))

	b, found := s.LookupTemplate("foo/bar")
	assert.Equal(t, true, found)
	assert.Equal(t, "foobar", string(b))

	assert.NoError(t, s.RemoveTemplate("foo"))
	assert.Equal(t, 0, len(s.ListTemplates(ctx, "")))

	assert.Error(t, s.RemoveTemplate("foo"))
}
