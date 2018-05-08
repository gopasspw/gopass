package action

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGitCredentialFormat(t *testing.T) {
	data := []io.Reader{
		strings.NewReader("" +
			"protocol=https\n" +
			"host=example.com\n" +
			"username=bob\n" +
			"foo=bar\n" +
			"password=secr3=t\n" +
			"test=1",
		),
		strings.NewReader("" +
			"protocol=https\n" +
			"host=example.com\n" +
			"username=bob\n" +
			"foo=bar\n" +
			"password=secr3=t\n" +
			"test=",
		),
	}
	results := []gitCredentials{
		{
			Host:     "example.com",
			Password: "secr3=t",
			Path:     "bar",
			Protocol: "https",
			Username: "bob",
		},
		{},
	}
	expectsErr := []bool{false, true}
	for i := range data {
		result, err := parseGitCredentials(data[i])
		if expectsErr[i] {
			assert.NotEqual(t, nil, err)
		} else {
			assert.Error(t, err)
		}
		if err != nil {
			continue
		}
		assert.Equal(t, results[i], result)
		buf := &bytes.Buffer{}
		n, err := result.WriteTo(buf)
		assert.Error(t, err, "could not serialize credentials")
		assert.Equal(t, buf.Len(), n)
		parseback, err := parseGitCredentials(buf)
		assert.Error(t, err, "failed parsing my own output")
		assert.Equal(t, results[i], parseback, "failed parsing my own output")
	}
}

func TestGitCredentialGet(t *testing.T) {

}

func TestGitCredentialStore(t *testing.T) {

}

func TestGitCredentialErase(t *testing.T) {

}
