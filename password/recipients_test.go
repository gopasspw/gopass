package password

import (
	"bufio"
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

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
	genRecs, _, err := createStore(tempdir)
	assert.NoError(t, err)

	s, err := NewStore("", tempdir, nil)
	assert.NoError(t, err)

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
	genRecs, _, err := createStore(tempdir)
	assert.NoError(t, err)

	s, err := NewStore("", tempdir, nil)
	assert.NoError(t, err)

	// remove recipients
	_ = os.Remove(filepath.Join(tempdir, gpgID))

	err = s.saveRecipients()
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

	for i := 0; i < len(genRecs); i++ {
		if i >= len(foundRecs) {
			t.Errorf("Read too few recipients")
			break
		}
		if genRecs[i] != foundRecs[i] {
			t.Errorf("Mismatch at %d: %s vs %s", i, genRecs[i], foundRecs[i])
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
	genRecs, _, err := createStore(tempdir)
	assert.NoError(t, err)

	s, err := NewStore("", tempdir, nil)
	assert.NoError(t, err)

	newRecp := "A3683834"
	newFP := "1E52C1335AC1F4F4FE02F62AB5B44266A3683834"

	if os.Getenv("GOPASS_INTEGRATION_TESTS") != "true" {
		t.Skip("Skipping test. GOPASS_INTEGRATION_TESTS is not true")
	}

	err = s.AddRecipient(newRecp)
	if err != nil {
		t.Fatalf("Failed to add Recipient: %s", err)
	}
	assert.Equal(t, append(genRecs, newFP), s.recipients)
}

func TestRemoveRecipient(t *testing.T) {
	tempdir, err := ioutil.TempDir("", "gopass-")
	if err != nil {
		t.Fatalf("Failed to create tempdir: %s", err)
	}
	defer func() {
		_ = os.RemoveAll(tempdir)
	}()
	genRecs, _, err := createStore(tempdir)
	assert.NoError(t, err)

	s, err := NewStore("", tempdir, nil)
	assert.NoError(t, err)

	if os.Getenv("GOPASS_INTEGRATION_TESTS") != "true" {
		t.Skip("Skipping test. GOPASS_INTEGRATION_TESTS is not true")
	}

	err = s.RemoveRecipient(genRecs[0])
	assert.NoError(t, err)
	assert.Equal(t, genRecs[1:], s.recipients)
}

func TestListRecipients(t *testing.T) {
	tempdir, err := ioutil.TempDir("", "gopass-")
	if err != nil {
		t.Fatalf("Failed to create tempdir: %s", err)
	}
	defer func() {
		_ = os.RemoveAll(tempdir)
	}()
	genRecs, _, err := createStore(tempdir)
	assert.NoError(t, err)

	s, err := NewStore("", tempdir, nil)
	assert.NoError(t, err)

	assert.NoError(t, err)
	assert.Equal(t, genRecs, s.recipients)
}
