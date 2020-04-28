package action

import (
	"bytes"
	"context"
	"flag"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/fatih/color"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/pkg/out"
	"github.com/gopasspw/gopass/pkg/termio"
	"github.com/gopasspw/gopass/tests/gptest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"
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
	u := gptest.NewUnitTester(t)
	defer u.Remove()

	ctx := context.Background()
	act, err := newMock(ctx, u)
	require.NoError(t, err)
	require.NotNil(t, act)

	stdout := &bytes.Buffer{}
	out.Stdout = stdout
	color.NoColor = true
	defer func() {
		out.Stdout = os.Stdout
	}()

	defer func() {
		termio.Stdin = os.Stdin
	}()

	app := cli.NewApp()

	fs := flag.NewFlagSet("default", flag.ContinueOnError)
	c := cli.NewContext(app, fs, nil)

	// before without stdin
	assert.Error(t, act.GitCredentialBefore(ctx, c))

	// before with stdin
	ctx = ctxutil.WithStdin(ctx, true)
	assert.NoError(t, act.GitCredentialBefore(ctx, c))

	s := "protocol=https\n" +
		"host=example.com\n" +
		"username=bob\n"

	termio.Stdin = strings.NewReader(s)
	assert.NoError(t, act.GitCredentialGet(ctx, c))
	assert.Equal(t, "", stdout.String())

	termio.Stdin = strings.NewReader(s + "password=secr3=t\n")
	assert.NoError(t, act.GitCredentialStore(ctx, c))
	stdout.Reset()

	termio.Stdin = strings.NewReader(s)
	assert.NoError(t, act.GitCredentialGet(ctx, c))
	read, err := parseGitCredentials(stdout)
	assert.NoError(t, err)
	assert.Equal(t, "secr3=t", read.Password)
	stdout.Reset()

	termio.Stdin = strings.NewReader("host=example.com\n")
	assert.NoError(t, act.GitCredentialGet(ctx, c))
	read, err = parseGitCredentials(stdout)
	assert.NoError(t, err)
	assert.Equal(t, "secr3=t", read.Password)
	assert.Equal(t, "bob", read.Username)
	stdout.Reset()

	termio.Stdin = strings.NewReader(s)
	assert.NoError(t, act.GitCredentialErase(ctx, c))
	assert.Equal(t, "", stdout.String())

	termio.Stdin = strings.NewReader(s)
	assert.NoError(t, act.GitCredentialGet(ctx, c))
	assert.Equal(t, "", stdout.String())

	termio.Stdin = strings.NewReader("a")
	assert.Error(t, act.GitCredentialGet(ctx, c))
	termio.Stdin = strings.NewReader("a")
	assert.Error(t, act.GitCredentialStore(ctx, c))
	termio.Stdin = strings.NewReader("a")
	assert.Error(t, act.GitCredentialErase(ctx, c))
}
