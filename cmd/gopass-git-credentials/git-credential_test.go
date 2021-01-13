package main

import (
	"bytes"
	"context"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/fatih/color"
	"github.com/gopasspw/gopass/internal/gptest"
	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/pkg/gopass/apimock"
	"github.com/gopasspw/gopass/pkg/termio"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGitCredentialFormat(t *testing.T) {
	data := []io.Reader{
		strings.NewReader("" +
			"protocol=https\n" +
			"host=example.com\n" +
			"username=bob\n" +
			"foo=bar\n" +
			"path=test\n" +
			"password=secr3=t\n",
		),
		strings.NewReader("" +
			"protocol=https\n" +
			"host=example.com\n" +
			"username=bob\n" +
			"foo=bar\n" +
			"password=secr3=t\n" +
			"test=",
		),
		strings.NewReader("" +
			"protocol=https\n" +
			"host=example.com\n" +
			"username=bob\n" +
			"foo=bar\n" +
			"password=secr3=t\n" +
			"test",
		),
	}
	results := []gitCredentials{
		{
			Host:     "example.com",
			Password: "secr3=t",
			Path:     "test",
			Protocol: "https",
			Username: "bob",
		},
		{},
		{},
	}
	expectsErr := []bool{false, true, true}
	for i := range data {
		result, err := parseGitCredentials(data[i])
		if expectsErr[i] {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
		}
		if err != nil {
			continue
		}
		assert.Equal(t, results[i], *result)
		buf := &bytes.Buffer{}
		n, err := result.WriteTo(buf)
		assert.NoError(t, err, "could not serialize credentials")
		assert.Equal(t, buf.Len(), int(n))
		parseback, err := parseGitCredentials(buf)
		assert.NoError(t, err, "failed parsing my own output")
		assert.Equal(t, results[i], *parseback, "failed parsing my own output")
	}
}

func TestGitCredentialHelper(t *testing.T) {
	ctx := context.Background()
	act := &gc{
		gp: apimock.New(),
	}
	require.NoError(t, act.gp.Set(ctx, "foo", &apimock.Secret{Buf: []byte("bar")}))

	stdout := &bytes.Buffer{}
	out.Stdout = stdout
	Stdout = stdout
	color.NoColor = true
	defer func() {
		out.Stdout = os.Stdout
		Stdout = os.Stdout
		termio.Stdin = os.Stdin
	}()

	c := gptest.CliCtx(ctx, t)

	// before without stdin
	assert.Error(t, act.Before(c))

	// before with stdin
	ctx = ctxutil.WithStdin(ctx, true)
	c.Context = ctx
	assert.NoError(t, act.Before(c))

	s := "protocol=https\n" +
		"host=example.com\n" +
		"username=bob\n"

	termio.Stdin = strings.NewReader(s)
	assert.NoError(t, act.Get(c))
	assert.Equal(t, "", stdout.String())

	termio.Stdin = strings.NewReader(s + "password=secr3=t\n")
	assert.NoError(t, act.Store(c))
	stdout.Reset()

	termio.Stdin = strings.NewReader(s)
	assert.NoError(t, act.Get(c))
	read, err := parseGitCredentials(stdout)
	assert.NoError(t, err)
	assert.Equal(t, "secr3=t", read.Password)
	stdout.Reset()

	termio.Stdin = strings.NewReader("host=example.com\n")
	assert.NoError(t, act.Get(c))
	read, err = parseGitCredentials(stdout)
	assert.NoError(t, err)
	assert.Equal(t, "secr3=t", read.Password)
	assert.Equal(t, "bob", read.Username)
	stdout.Reset()

	termio.Stdin = strings.NewReader(s)
	assert.NoError(t, act.Erase(c))
	assert.Equal(t, "", stdout.String())

	termio.Stdin = strings.NewReader(s)
	assert.NoError(t, act.Get(c))
	assert.Equal(t, "", stdout.String())

	termio.Stdin = strings.NewReader("a")
	assert.Error(t, act.Get(c))
	termio.Stdin = strings.NewReader("a")
	assert.Error(t, act.Store(c))
	termio.Stdin = strings.NewReader("a")
	assert.Error(t, act.Erase(c))
}
