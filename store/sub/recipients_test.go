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

func TestGetRecipientsDefault(t *testing.T) {
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
		alias: "",
		path:  tempdir,
		gpg:   gpgmock.New(),
	}

	recs, err := s.getRecipients("")
	assert.NoError(t, err)
	assert.Equal(t, genRecs, recs)
}

func TestGetRecipientsSubID(t *testing.T) {
	tempdir, err := ioutil.TempDir("", "gopass-")
	if err != nil {
		t.Fatalf("Failed to create tempdir: %s", err)
	}
	defer func() {
		//_ = os.RemoveAll(tempdir)
	}()

	genRecs, _, err := createStore(tempdir, nil, nil)
	assert.NoError(t, err)

	s := &Store{
		alias: "",
		path:  tempdir,
		gpg:   gpgmock.New(),
	}

	recs, err := s.getRecipients("")
	assert.NoError(t, err)
	assert.Equal(t, genRecs, recs)

	err = ioutil.WriteFile(filepath.Join(tempdir, "foo", "bar", GPGID), []byte("john.doe\n"), 0600)
	assert.NoError(t, err)

	recs, err = s.getRecipients("foo/bar/baz")
	assert.NoError(t, err)
	assert.Equal(t, []string{"john.doe"}, recs)
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
		alias: "",
		path:  tempdir,
		gpg:   gpgmock.New(),
	}

	// remove recipients
	_ = os.Remove(filepath.Join(tempdir, GPGID))

	err = s.saveRecipients(recp, "test-save-recipients", true)
	assert.NoError(t, err)

	buf, err := ioutil.ReadFile(s.idFile(""))
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

	genRecs, _, err := createStore(tempdir, nil, nil)
	assert.NoError(t, err)

	s := &Store{
		alias: "",
		path:  tempdir,
		gpg:   gpgmock.New(),
	}

	newRecp := "A3683834"

	err = s.AddRecipient(newRecp)
	assert.NoError(t, err)

	rs, err := s.getRecipients("")
	assert.NoError(t, err)
	assert.Equal(t, append(genRecs, newRecp), rs)
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
		alias: "",
		path:  tempdir,
		gpg:   gpgmock.New(),
	}

	err = s.RemoveRecipient("0xDEADBEEF")
	assert.NoError(t, err)

	rs, err := s.getRecipients("")
	assert.NoError(t, err)
	assert.Equal(t, []string{"0xFEEDBEEF"}, rs)
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

	rs, err := s.getRecipients("")
	assert.NoError(t, err)
	assert.Equal(t, genRecs, rs)
}
