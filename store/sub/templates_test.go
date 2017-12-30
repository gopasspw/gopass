package sub

import (
	"io/ioutil"
	"os"
	"testing"

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

	_, _, err = createStore(tempdir, nil, nil)
	assert.NoError(t, err)

	s := New(
		"",
		tempdir,
		gpgmock.New(),
	)

	if len(s.ListTemplates("")) != 0 {
		t.Errorf("Should have no templates")
	}

	if err := s.SetTemplate("foo", []byte("foobar")); err != nil {
		t.Errorf("Failed to write template: %s", err)
	}

	if len(s.ListTemplates("")) != 1 {
		t.Errorf("Should have one template")
	}

	b, err := s.GetTemplate("foo")
	if err != nil {
		t.Errorf("Error: %s", err)
	}
	if string(b) != "foobar" {
		t.Errorf("Wrong template: %s", b)
	}

	b, found := s.LookupTemplate("foo/bar")
	if !found {
		t.Errorf("No template found")
	}
	if string(b) != "foobar" {
		t.Errorf("Wrong template: %s", b)
	}

	if err := s.RemoveTemplate("foo"); err != nil {
		t.Errorf("Failed to remove template: %s", err)
	}

	if len(s.ListTemplates("")) != 0 {
		t.Errorf("Should have no templates")
	}
}
