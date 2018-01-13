package sub

import (
	"context"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	gpgmock "github.com/justwatchcom/gopass/backend/crypto/gpg/mock"
	"github.com/justwatchcom/gopass/store/secret"
	"github.com/stretchr/testify/assert"
)

func createSubStore(dir string) (*Store, error) {
	sd := filepath.Join(dir, "sub")
	_, _, err := createStore(sd, nil, nil)
	if err != nil {
		return nil, err
	}

	if err := os.Setenv("GOPASS_CONFIG", filepath.Join(dir, ".gopass.yml")); err != nil {
		return nil, err
	}
	if err := os.Setenv("GOPASS_HOMEDIR", dir); err != nil {
		return nil, err
	}
	if err := os.Unsetenv("PAGER"); err != nil {
		return nil, err
	}
	if err := os.Setenv("CHECKPOINT_DISABLE", "true"); err != nil {
		return nil, err
	}
	if err := os.Setenv("GOPASS_NO_NOTIFY", "true"); err != nil {
		return nil, err
	}

	gpgDir := filepath.Join(dir, ".gnupg")
	if err := os.Setenv("GNUPGHOME", gpgDir); err != nil {
		return nil, err
	}

	return New(
		"",
		sd,
		gpgmock.New(),
	)
}

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
	assert.NoError(t, err)
	defer func() {
		_ = os.RemoveAll(tempdir)
	}()

	s, err := createSubStore(tempdir)
	assert.NoError(t, err)

	if !s.Equals(s) {
		t.Errorf("Should be equal to myself")
	}
}

func TestIdFile(t *testing.T) {
	ctx := context.Background()

	tempdir, err := ioutil.TempDir("", "gopass-")
	assert.NoError(t, err)
	defer func() {
		_ = os.RemoveAll(tempdir)
	}()

	s, err := createSubStore(tempdir)
	assert.NoError(t, err)

	// test sub-id
	secName := "a"
	for i := 0; i < 99; i++ {
		secName += "/a"
	}
	assert.NoError(t, s.Set(ctx, secName, secret.New("foo", "bar")))
	assert.NoError(t, ioutil.WriteFile(filepath.Join(tempdir, "sub", "a", GPGID), []byte("foobar"), 0600))
	assert.Equal(t, filepath.Join(tempdir, "sub", "a", GPGID), s.idFile(secName))
	assert.Equal(t, true, s.Exists(secName))

	// test abort condition
	secName = "a"
	for i := 0; i < 100; i++ {
		secName += "/a"
	}
	assert.NoError(t, s.Set(ctx, secName, secret.New("foo", "bar")))
	assert.NoError(t, ioutil.WriteFile(filepath.Join(tempdir, "sub", "a", GPGID), []byte("foobar"), 0600))
	assert.Equal(t, filepath.Join(tempdir, "sub", GPGID), s.idFile(secName))
}
