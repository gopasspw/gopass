package sub

import (
	"bufio"
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/justwatchcom/gopass/config"
	gpgmock "github.com/justwatchcom/gopass/gpg/mock"
	"github.com/stretchr/testify/assert"
)

func TestLoadRecipients(t *testing.T) {
	tempdir, err := ioutil.TempDir("", "gopass-")
	if err != nil {
		t.Fatalf("Failed to create tempdir: %s", err)
	}
	defer func() {
		_ = os.RemoveAll(tempdir)
	}()
	genRecs, _, err := createStore(tempdir, nil, nil)
	assert.NoError(t, err)

	s := &Store{
		alias:      "",
		path:       tempdir,
		gpg:        gpgmock.New(),
		recipients: []string{"john.doe"},
	}

	recs, err := s.loadRecipients()
	assert.NoError(t, err)
	assert.Equal(t, genRecs, recs)
}

func TestSaveRecipients(t *testing.T) {
	tempdir, err := ioutil.TempDir("", "gopass-")
	if err != nil {
		t.Fatalf("Failed to create tempdir: %s", err)
	}
	defer func() {
		_ = os.RemoveAll(tempdir)
	}()
	_, _, err = createStore(tempdir, nil, nil)
	assert.NoError(t, err)

	recp := []string{"john.doe"}
	s := &Store{
		alias:      "",
		path:       tempdir,
		gpg:        gpgmock.New(),
		recipients: recp,
	}

	// remove recipients
	_ = os.Remove(filepath.Join(tempdir, GPGID))

	err = s.saveRecipients("test-save-recipients")
	assert.NoError(t, err)

	buf, err := ioutil.ReadFile(s.idFile())
	if err != nil {
		t.Fatalf("Failed to read ID File: %s", err)
	}

	foundRecs := []string{}
	scanner := bufio.NewScanner(bytes.NewReader(buf))
	for scanner.Scan() {
		foundRecs = append(foundRecs, strings.TrimSpace(scanner.Text()))
	}
	sort.Strings(foundRecs)

	for i := 0; i < len(recp); i++ {
		if i >= len(foundRecs) {
			t.Errorf("Read too few recipients")
			break
		}
		if recp[i] != foundRecs[i] {
			t.Errorf("Mismatch at %d: %s vs %s", i, recp[i], foundRecs[i])
		}
	}
}

func TestAddRecipient(t *testing.T) {
	tempdir, err := ioutil.TempDir("", "gopass-")
	if err != nil {
		t.Fatalf("Failed to create tempdir: %s", err)
	}
	defer func() {
		_ = os.RemoveAll(tempdir)
	}()
	_, _, err = createStore(tempdir, nil, nil)
	assert.NoError(t, err)

	s := &Store{
		alias:      "",
		path:       tempdir,
		gpg:        gpgmock.New(),
		recipients: []string{"john.doe"},
	}

	newRecp := "A3683834"

	err = s.AddRecipient(newRecp)
	if err != nil {
		t.Fatalf("Failed to add Recipient: %s", err)
	}
	assert.Equal(t, []string{"john.doe", newRecp}, s.recipients)
}

func TestRemoveRecipient(t *testing.T) {
	tempdir, err := ioutil.TempDir("", "gopass-")
	if err != nil {
		t.Fatalf("Failed to create tempdir: %s", err)
	}
	defer func() {
		_ = os.RemoveAll(tempdir)
	}()
	_, _, err = createStore(tempdir, nil, nil)
	assert.NoError(t, err)

	s := &Store{
		alias:      "",
		path:       tempdir,
		gpg:        gpgmock.New(),
		recipients: []string{"john.doe"},
	}

	err = s.RemoveRecipient("john.doe")
	assert.NoError(t, err)
	assert.Equal(t, []string{}, s.recipients)
}

func TestListRecipients(t *testing.T) {
	tempdir, err := ioutil.TempDir("", "gopass-")
	if err != nil {
		t.Fatalf("Failed to create tempdir: %s", err)
	}
	defer func() {
		_ = os.RemoveAll(tempdir)
	}()
	genRecs, _, err := createStore(tempdir, nil, nil)
	assert.NoError(t, err)

	s, err := New("", &config.Config{Path: tempdir})
	assert.NoError(t, err)

	assert.NoError(t, err)
	assert.Equal(t, genRecs, s.recipients)
}
