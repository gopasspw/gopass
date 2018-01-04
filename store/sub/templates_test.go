package sub

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/fatih/color"
	gpgmock "github.com/justwatchcom/gopass/backend/gpg/mock"
	"github.com/stretchr/testify/assert"
)

func TestTemplates(t *testing.T) {
	tempdir, err := ioutil.TempDir("", "gopass-")
	if err != nil {
		t.Fatalf("Failed to create tempdir: %s", err)
	}
	defer func() {
		_ = os.RemoveAll(tempdir)
	}()

	color.NoColor = true

	_, _, err = createStore(tempdir, nil, nil)
	assert.NoError(t, err)

	s := New(
		"",
		tempdir,
		gpgmock.New(),
	)

	assert.Equal(t, 0, len(s.ListTemplates("")))
	assert.NoError(t, s.SetTemplate("foo", []byte("foobar")))
	assert.Equal(t, 1, len(s.ListTemplates("")))

	tt, err := s.TemplateTree()
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
	assert.Equal(t, 0, len(s.ListTemplates("")))

	assert.Error(t, s.RemoveTemplate("foo"))
}
