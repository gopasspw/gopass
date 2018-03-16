package agent

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

type fakePinentry struct {
	pin []byte
}

func (f *fakePinentry) Close() {}
func (f *fakePinentry) Confirm() bool {
	return true
}
func (f *fakePinentry) Set(key, value string) error {
	return nil
}
func (f *fakePinentry) GetPin() ([]byte, error) {
	return f.pin, nil
}

func TestServePing(t *testing.T) {
	w := httptest.NewRecorder()
	r, err := http.NewRequest("GET", "/", nil)
	assert.NoError(t, err)
	a := New(os.TempDir())

	a.servePing(w, r)
	assert.Equal(t, "OK", w.Body.String())
}

func TestServeRemove(t *testing.T) {
	w := httptest.NewRecorder()
	r, err := http.NewRequest("GET", "/?key=foo", nil)
	assert.NoError(t, err)
	a := New(os.TempDir())

	a.serveRemove(w, r)
	assert.Equal(t, "OK", w.Body.String())
}

func TestServePurge(t *testing.T) {
	w := httptest.NewRecorder()
	r, err := http.NewRequest("GET", "/", nil)
	assert.NoError(t, err)
	a := New(os.TempDir())

	a.serveRemove(w, r)
	assert.Equal(t, "OK", w.Body.String())
}

func TestServePassphrase(t *testing.T) {
	w := httptest.NewRecorder()
	r, err := http.NewRequest("GET", "/?key=foo&reason=bar", nil)
	assert.NoError(t, err)
	a := New(os.TempDir())
	a.pinentry = func() (piner, error) {
		return &fakePinentry{[]byte("foobar")}, nil
	}

	a.servePassphrase(w, r)
	assert.Equal(t, "foobar", w.Body.String())
}
