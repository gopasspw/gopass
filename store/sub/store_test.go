package sub

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	gpgmock "github.com/justwatchcom/gopass/backend/gpg/mock"
	"github.com/stretchr/testify/assert"
)

func createStore(dir string, recipients, entries []string) ([]string, []string, error) {
	if recipients == nil {
		recipients = []string{
			"0xDEADBEEF",
			"0xFEEDBEEF",
		}
	}
	if entries == nil {
		entries = []string{
			"foo/bar/baz",
			"baz/ing/a",
		}
	}
	sort.Strings(entries)
	for _, file := range entries {
		filename := filepath.Join(dir, file+".gpg")
		if err := os.MkdirAll(filepath.Dir(filename), 0700); err != nil {
			return recipients, entries, err
		}
		if err := ioutil.WriteFile(filename, []byte{}, 0644); err != nil {
			return recipients, entries, err
		}
	}
	err := ioutil.WriteFile(filepath.Join(dir, GPGID), []byte(strings.Join(recipients, "\n")), 0600)
	return recipients, entries, err
}

func TestStore(t *testing.T) {
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

	if !s.Equals(s) {
		t.Errorf("Should be equal to myself")
	}
}
