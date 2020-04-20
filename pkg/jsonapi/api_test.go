package jsonapi

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/blang/semver"
	"github.com/gopasspw/gopass/pkg/backend"
	"github.com/gopasspw/gopass/pkg/config"
	"github.com/gopasspw/gopass/pkg/otp"
	"github.com/gopasspw/gopass/pkg/store"
	"github.com/gopasspw/gopass/pkg/store/root"
	"github.com/gopasspw/gopass/pkg/store/secret"
	"github.com/gopasspw/gopass/tests/gptest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type storedSecret struct {
	Name   []string
	Secret store.Secret
}

func TestRespondMessageBrokenInput(t *testing.T) {
	// Garbage input
	runRespondRawMessage(t, "1234Xabcd", "", "incomplete message read", []storedSecret{})

	// Too short to determine message size
	runRespondRawMessage(t, " ", "", "not enough bytes read to determine message size", []storedSecret{})

	// Empty message
	runRespondMessage(t, "", "", "failed to unmarshal JSON message: unexpected end of JSON input", []storedSecret{})

	// Empty object
	runRespondMessage(t, "{}", "", "unknown message of type ", []storedSecret{})
}

func TestRespondGetVersion(t *testing.T) {
	runRespondMessage(t,
		`{"type": "getVersion"}`,
		`{"version":"1.2.3-test","major":1,"minor":2,"patch":3}`,
		"",
		nil)
}

func TestRespondMessageGetData(t *testing.T) {
	secrets := []storedSecret{
		{[]string{"foo"}, secret.New("20", "hallo: welt")},
		{[]string{"bar"}, secret.New("20", "---\nlogin: muh")},
		{[]string{"complex"}, secret.New("20", `---
login: hallo
number: 42
sub:
  subentry: 123
`)},
	}

	runRespondMessage(t,
		`{"type":"getData","entry":"foo"}`,
		`{"hallo":"welt"}`,
		"", secrets)

	runRespondMessage(t,
		`{"type":"getData","entry":"bar"}`,
		`{"login":"muh"}`,
		"", secrets)
	runRespondMessage(t,
		`{"type":"getData","entry":"complex"}`,
		`{"login":"hallo","number":42,"sub":{"subentry":123}}`,
		"", secrets)
}

func TestRespondMessageGetDataTOTP(t *testing.T) {
	totpSuffix := "//totp/github-fake-account?secret=rpna55555qyho42j"
	totpURL := "otpauth:" + totpSuffix
	totpSecret := secret.New("totp_are_cool", totpURL)

	secrets := []storedSecret{
		{[]string{"totp"}, totpSecret},
		{[]string{"foo"}, secret.New("20", "hallo: welt")},
		{[]string{"bar"}, secret.New("20", "---\nlogin: muh")},
		{[]string{"complex"}, secret.New("20", `---
login: hallo
number: 42
sub:
  subentry: 123
`)},
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// this fails if the time second switches between `otp.Calculate` below and in pkg/jsonapi/responses
	// if this is the case a different (wrong) totp is generated
	// we need this in order to make it testable: https://github.com/gokyle/twofactor/pull/17
	ok := gptest.Retry(t, 10, time.Millisecond*5, func(r *gptest.R) {
		totp, _, err := otp.Calculate(ctx, "_", totpSecret)
		if err != nil {
			assert.NoError(t, err)
		}
		expectedTotp := totp.OTP()
		input := writeMessageWithLength(`{"type":"getData","entry":"totp"}`)
		outRegex := regexp.MustCompile(fmt.Sprintf(`{"current_totp":"%s","otpauth":"(.+)"}`, expectedTotp))

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		tempdir, err := ioutil.TempDir("", "gopass-")
		require.NoError(t, err)
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
		require.NoError(t, err)
		require.NotNil(t, store)
		inited, err := store.Initialized(ctx)
		require.NoError(t, err)
		assert.Equal(t, false, inited)
		assert.NoError(t, populateStore(tempdir, secrets))
		var inbuf bytes.Buffer
		var outbuf bytes.Buffer
		api := API{store, &inbuf, &outbuf, semver.MustParse("1.2.3-test")}
		_, err = inbuf.Write([]byte(input))
		assert.NoError(t, err)
		err = api.ReadAndRespond(ctx)
		assert.NoError(t, err, "api should not respond with error")
		outputMessage := readAndVerifyMessageLength(t, outbuf.Bytes())

		if !outRegex.MatchString(outputMessage) {
			r.Fail()
		}
	})

	assert.True(t, ok, "getting totp should not fail")

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

func TestCopyToClipboard(t *testing.T) {
	secrets := []storedSecret{
		{[]string{"foo", "bar"}, secret.New("20", "")},
		{[]string{"yamllogin"}, secret.New("thesecret", "---\nlogin: muh")},
	}

	// copy nonexisting entry returns error
	runRespondMessage(t,
		`{"type": "copyToClipboard","entry":"doesnotexist"}`,
		``,
		"failed to get secret: Entry is not in the password store",
		secrets)

	// copy existing entry
	runRespondMessage(t,
		`{"type": "copyToClipboard","entry":"foo/bar"}`,
		`{"status":"ok"}`,
		"",
		secrets)

	// copy nonexisting subkey
	runRespondMessage(t,
		`{"type": "copyToClipboard","entry":"foo/bar","key":"baz"}`,
		``,
		"failed to get secret sub entry: key not found in YAML document",
		secrets)

	// copy existing subkey
	runRespondMessage(t,
		`{"type": "copyToClipboard","entry":"yamllogin","key":"login"}`,
		`{"status":"ok"}`,
		"",
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
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	tempdir, err := ioutil.TempDir("", "gopass-")
	require.NoError(t, err)
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
	require.NoError(t, err)
	require.NotNil(t, store)
	inited, err := store.Initialized(ctx)
	require.NoError(t, err)
	assert.Equal(t, false, inited)
	assert.NoError(t, populateStore(tempdir, secrets))

	for _, request := range requests {
		var inbuf bytes.Buffer
		var outbuf bytes.Buffer

		api := API{store, &inbuf, &outbuf, semver.MustParse("1.2.3-test")}

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
		assert.NotEqual(t, "", request.OutputRegexpStr, "Empty string would match any output")
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
	input := bytes.NewReader(rawMessage)
	lenBytes := make([]byte, 4)

	_, err := input.Read(lenBytes)
	assert.NoError(t, err)

	length, err := getMessageLength(lenBytes)
	assert.NoError(t, err)
	assert.Equal(t, len(rawMessage)-4, length)

	msgBytes := make([]byte, length)
	_, err = input.Read(msgBytes)
	assert.NoError(t, err)
	return string(msgBytes)
}
