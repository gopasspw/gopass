package jsonapi

import (
	"bytes"
	"encoding/binary"

	"context"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/justwatchcom/gopass/pkg/backend"
	"github.com/justwatchcom/gopass/pkg/config"
	"github.com/justwatchcom/gopass/pkg/store"
	"github.com/justwatchcom/gopass/pkg/store/root"
	"github.com/justwatchcom/gopass/pkg/store/secret"
	"github.com/stretchr/testify/assert"
)

type storedSecret struct {
	Name   []string
	Secret store.Secret
}

func TestRespondMessageBrokenInput(t *testing.T) {
	// Garbage input
	runRespondRawMessage(t, "1234Xabcd", "", "incomplete message read", []storedSecret{})

	// Too short to determine message size
	runRespondRawMessage(t, " ", "", "not enough bytes read to deterimine message size", []storedSecret{})

	// Empty message
	runRespondMessage(t, "", "", "failed to unmarshal JSON message: unexpected end of JSON input", []storedSecret{})

	// Empty object
	runRespondMessage(t, "{}", "", "Unknown message of type ", []storedSecret{})
}

func TestRespondMessageQuery(t *testing.T) {
	secrets := []storedSecret{
		{[]string{"awesomePrefix", "foo", "bar"}, secret.New("20", "")},
		{[]string{"awesomePrefix", "fixed", "secret"}, secret.New("moar", "")},
		{[]string{"awesomePrefix", "fixed", "yamllogin"}, secret.New("thesecret", "---\nlogin: muh")},
		{[]string{"awesomePrefix", "fixed", "yamlother"}, secret.New("thesecret", "---\nother: meh")},
		{[]string{"awesomePrefix", "some.other.host", "other"}, secret.New("thesecret", "---\nother: meh")},
		{[]string{"awesomePrefix", "b", "some.other.host"}, secret.New("thesecret", "---\nother: meh")},
		{[]string{"awesomePrefix", "evilsome.other.host"}, secret.New("thesecret", "---\nother: meh")},
		{[]string{"evilsome.other.host", "something"}, secret.New("thesecret", "---\nother: meh")},
		{[]string{"awesomePrefix", "other.host", "other"}, secret.New("thesecret", "---\nother: meh")},
	}

	// query for keys without any matching
	runRespondMessage(t,
		`{"type":"query","query":"notfound"}`,
		`\[\]`,
		"", secrets)

	// query for keys with matching one
	runRespondMessage(t,
		`{"type":"query","query":"foo"}`,
		`\["awesomePrefix/foo/bar"\]`,
		"", secrets)

	// query for keys with matching multiple
	runRespondMessage(t,
		`{"type":"query","query":"yaml"}`,
		`\["awesomePrefix/fixed/yamllogin","awesomePrefix/fixed/yamlother"\]`,
		"", secrets)

	// query for host
	runRespondMessage(t,
		`{"type":"queryHost","host":"find.some.other.host"}`,
		`\["awesomePrefix/b/some.other.host","awesomePrefix/some.other.host/other"\]`,
		"", secrets)

	// get username / password for key without value in yaml
	runRespondMessage(t,
		`{"type":"getLogin","entry":"awesomePrefix/fixed/secret"}`,
		`{"username":"secret","password":"moar"}`,
		"", secrets)

	// get username / password for key with login in yaml
	runRespondMessage(t,
		`{"type":"getLogin","entry":"awesomePrefix/fixed/yamllogin"}`,
		`{"username":"muh","password":"thesecret"}`,
		"", secrets)

	// get username / password for key with no login in yaml (fallback)
	runRespondMessage(t,
		`{"type":"getLogin","entry":"awesomePrefix/fixed/yamlother"}`,
		`{"username":"yamlother","password":"thesecret"}`,
		"", secrets)
}

func TestRespondMessageCreate(t *testing.T) {
	secrets := []storedSecret{
		{[]string{"awesomePrefix", "overwrite", "me"}, secret.New("20", "")},
	}

	// store new secret with given password
	runRespondMessages(t, []verifiedRequest{
		{
			`{"type":"create","entry_name":"prefix/stored","login":"myname","password":"mypass","length":16,"generate":false,"use_symbols":true}`,
			`{"username":"myname","password":"mypass"}`,
			"",
		},
		{
			`{"type":"getLogin","entry":"prefix/stored"}`,
			`{"username":"myname","password":"mypass"}`,
			"",
		},
	}, secrets)

	// generate new secret with given length and without symbols
	runRespondMessages(t, []verifiedRequest{
		{
			`{"type":"create","entry_name":"prefix/generated","login":"myname","password":"","length":12,"generate":true,"use_symbols":false}`,
			`{"username":"myname","password":"\w{12}"}`,
			"",
		},
		{
			`{"type":"getLogin","entry":"prefix/generated"}`,
			`{"username":"myname","password":"\w{12}"}`,
			"",
		},
	}, secrets)

	// generate new secret with given length and with symbols
	runRespondMessages(t, []verifiedRequest{
		{
			`{"type":"create","entry_name":"prefix/generated","login":"myname","password":"","length":12,"generate":true,"use_symbols":true}`,
			`{"username":"myname","password":".{12,}"}`,
			"",
		},
		{
			`{"type":"getLogin","entry":"prefix/generated"}`,
			`{"username":"myname","password":".{12,}"}`,
			"",
		},
	}, secrets)

	// store already existing secret
	runRespondMessage(t,
		`{"type":"create","entry_name":"awesomePrefix/overwrite/me","login":"myname","password":"mypass","length":16,"generate":false,"use_symbols":true}`,
		"",
		"secret awesomePrefix/overwrite/me already exists",
		secrets)
}

func writeMessageWithLength(message string) string {
	buffer := bytes.NewBuffer([]byte{})
	_ = binary.Write(buffer, binary.LittleEndian, uint32(len(message)))
	_, _ = buffer.WriteString(message)
	return buffer.String()
}

func runRespondMessage(t *testing.T, inputStr, outputRegexpStr, errorStr string, secrets []storedSecret) {
	inputMessageStr := writeMessageWithLength(inputStr)
	runRespondRawMessage(t, inputMessageStr, outputRegexpStr, errorStr, secrets)
}

func runRespondRawMessage(t *testing.T, inputStr, outputRegexpStr, errorStr string, secrets []storedSecret) {
	runRespondRawMessages(t, []verifiedRequest{{inputStr, outputRegexpStr, errorStr}}, secrets)
}

type verifiedRequest struct {
	InputStr        string
	OutputRegexpStr string
	ErrorStr        string
}

func runRespondMessages(t *testing.T, requests []verifiedRequest, secrets []storedSecret) {
	for i, request := range requests {
		requests[i].InputStr = writeMessageWithLength(request.InputStr)
	}
	runRespondRawMessages(t, requests, secrets)
}

func runRespondRawMessages(t *testing.T, requests []verifiedRequest, secrets []storedSecret) {
	ctx := context.Background()

	tempdir, err := ioutil.TempDir("", "gopass-")
	assert.NoError(t, err)
	defer func() {
		_ = os.RemoveAll(tempdir)
	}()

	assert.NoError(t, os.Setenv("GOPASS_DISABLE_ENCRYPTION", "true"))
	ctx = backend.WithCryptoBackendString(ctx, "plain")
	store, err := root.New(
		ctx,
		&config.Config{
			Root: &config.StoreConfig{
				Path: backend.FromPath(tempdir),
			},
		},
	)
	assert.NoError(t, err)
	assert.Equal(t, false, store.Initialized(ctx))
	assert.NoError(t, populateStore(tempdir, secrets))

	for _, request := range requests {
		var inbuf bytes.Buffer
		var outbuf bytes.Buffer

		api := API{store, &inbuf, &outbuf}

		_, err = inbuf.Write([]byte(request.InputStr))
		assert.NoError(t, err)

		err = api.ReadAndRespond(ctx)
		if len(request.ErrorStr) > 0 {
			assert.EqualError(t, err, request.ErrorStr)
			assert.Equal(t, len(outbuf.String()), 0)
			continue
		}
		assert.NoError(t, err)
		outputMessage := readAndVerifyMessageLength(t, outbuf.Bytes())
		assert.Regexp(t, regexp.MustCompile(request.OutputRegexpStr), outputMessage)
	}
}

func populateStore(dir string, secrets []storedSecret) error {
	recipients := []string{
		"0xDEADBEEF",
		"0xFEEDBEEF",
	}
	for _, sec := range secrets {
		file := filepath.Join(sec.Name...)
		filename := filepath.Join(dir, file+".txt")
		if err := os.MkdirAll(filepath.Dir(filename), 0700); err != nil {
			return err
		}
		secBytes, err := sec.Secret.Bytes()
		if err != nil {
			return err
		}
		if err := ioutil.WriteFile(filename, secBytes, 0644); err != nil {
			return err
		}
	}
	return ioutil.WriteFile(filepath.Join(dir, ".gpg-id"), []byte(strings.Join(recipients, "\n")), 0600)
}

func readAndVerifyMessageLength(t *testing.T, rawMessage []byte) string {
	stdin := bytes.NewReader(rawMessage)
	lenBytes := make([]byte, 4)

	_, err := stdin.Read(lenBytes)
	assert.NoError(t, err)

	length, err := getMessageLength(lenBytes)
	assert.NoError(t, err)
	assert.Equal(t, len(rawMessage)-4, length)

	msgBytes := make([]byte, length)
	_, err = stdin.Read(msgBytes)
	assert.NoError(t, err)
	return string(msgBytes)
}
