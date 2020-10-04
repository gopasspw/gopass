package jsonapi

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/gopasspw/gopass/internal/debug"
	"github.com/gopasspw/gopass/internal/otp"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/pkg/gopass"
	"github.com/gopasspw/gopass/pkg/gopass/apimock"
	"github.com/gopasspw/gopass/pkg/gopass/secret/secparse"

	"github.com/blang/semver"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type storedSecret struct {
	Name   []string
	Secret gopass.Secret
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

func newSec(t *testing.T, in string) gopass.Secret {
	debug.Log("in: %s", in)
	sec, err := secparse.Parse([]byte(in))
	require.NoError(t, err)
	return sec
}

func TestRespondMessageQuery(t *testing.T) {
	secrets := []storedSecret{
		{[]string{"awesomePrefix", "foo", "bar"}, newSec(t, "20\n")},
		{[]string{"awesomePrefix", "fixed", "secret"}, newSec(t, "moar\n")},
		{[]string{"awesomePrefix", "fixed", "yamllogin"}, newSec(t, "thesecret\n---\nlogin: muh")},
		{[]string{"awesomePrefix", "fixed", "yamlother"}, newSec(t, "thesecret\n---\nother: meh")},
		{[]string{"awesomePrefix", "some.other.host", "other"}, newSec(t, "thesecret\n---\nother: meh")},
		{[]string{"awesomePrefix", "b", "some.other.host"}, newSec(t, "thesecret\n---\nother: meh")},
		{[]string{"awesomePrefix", "evilsome.other.host"}, newSec(t, "thesecret\n---\nother: meh")},
		{[]string{"evilsome.other.host", "something"}, newSec(t, "thesecret\n---\nother: meh")},
		{[]string{"awesomePrefix", "other.host", "other"}, newSec(t, "thesecret\n---\nother: meh")},
		{[]string{"somename", "github.com"}, newSec(t, "thesecret\n---\nother: meh")},
		{[]string{"login_entry"}, newSec(t, `thepass
---
login: thelogin
ignore: me
login_fields:
  first: 42
  second: ok
nologin_fields:
  subentry: 123`)},
		{[]string{"invalid_login_entry"}, newSec(t, `thepass
---
login: thelogin
ignore: me
login_fields: "invalid"`)},
	}

	for _, tc := range []struct {
		desc string
		in   string
		out  string
	}{
		{
			desc: "query for keys without any matches",
			in:   `{"type":"query","query":"notfound"}`,
			out:  `\[\]`,
		},
		{
			desc: "query for keys with matching one",
			in:   `{"type":"query","query":"foo"}`,
			out:  `\["awesomePrefix/foo/bar"\]`,
		},
		{
			desc: "query for keys with matching multiple",
			in:   `{"type":"query","query":"yaml"}`,
			out:  `\["awesomePrefix/fixed/yamllogin","awesomePrefix/fixed/yamlother"\]`,
		},
		{
			desc: "query for host",
			in:   `{"type":"queryHost","host":"find.some.other.host"}`,
			out:  `\["awesomePrefix/b/some.other.host","awesomePrefix/some.other.host/other"\]`,
		},
		{
			desc: "query for host not matches parent domain",
			in:   `{"type":"queryHost","host":"other.host"}`,
			out:  `\["awesomePrefix/other.host/other"\]`,
		},
		{
			desc: "query for host is query has different domain appended does not return partial match",
			in:   `{"type":"queryHost","host":"some.other.host.different.domain"}`,
			out:  `\[\]`,
		},
		{
			desc: "query returns result with public suffix at the end",
			in:   `{"type":"queryHost","host":"github.com"}`,
			out:  `\["somename/github.com"\]`,
		},
		{
			desc: "get username and password for key without value in yaml",
			in:   `{"type":"getLogin","entry":"awesomePrefix/fixed/secret"}`,
			out:  `{"username":"secret","password":"moar"}`,
		},
		{
			desc: "get username and password for key with login in yaml",
			in:   `{"type":"getLogin","entry":"awesomePrefix/fixed/yamllogin"}`,
			out:  `{"username":"muh","password":"thesecret"}`,
		},
		{
			desc: "get username and password for key with no login in yaml (fallback)",
			in:   `{"type":"getLogin","entry":"awesomePrefix/fixed/yamlother"}`,
			out:  `{"username":"yamlother","password":"thesecret"}`,
		},
		{
			desc: "get entry with invalid login fields",
			in:   `{"type":"getLogin","entry":"invalid_login_entry"}`,
			out:  `{"username":"thelogin","password":"thepass"}`,
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			runRespondMessage(t, tc.in, tc.out, "", secrets)
		})
	}

}

func TestRespondMessageGetData(t *testing.T) {
	totpSuffix := "//totp/github-fake-account?secret=rpna55555qyho42j"
	totpURL := "otpauth:" + totpSuffix
	totpSecret := newSec(t, "totp_are_cool\n"+totpURL)

	secrets := []storedSecret{
		{[]string{"totp"}, totpSecret},
		{[]string{"foo"}, newSec(t, "20\nhallo: welt")},
		{[]string{"bar"}, newSec(t, "20\n---\nlogin: muh")},
		{[]string{"complex"}, newSec(t, `20
---
login: hallo
number: 42
sub:
  subentry: 123
`)},
	}

	totp, _, err := otp.Calculate("_", totpSecret)
	if err != nil {
		assert.NoError(t, err)
	}
	expectedTotp := totp.OTP()

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
		`{"login":"hallo","number":"42","sub":"map.subentry:123."}`,
		"", secrets)

	runRespondMessage(t,
		`{"type":"getData","entry":"totp"}`,
		fmt.Sprintf(`{"current_totp":"%s"}`, expectedTotp),
		"", secrets)
}

func TestRespondMessageCreate(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	runRespondMessages(t, nil, nil)

	t.Skip("broken") // TODO fix this

	secrets := []storedSecret{
		{[]string{"awesomePrefix", "overwrite", "me"}, newSec(t, "20\n")},
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
		{[]string{"foo", "bar"}, newSec(t, "20\n")},
		{[]string{"yamllogin"}, newSec(t, "thesecret\n---\nlogin: muh")},
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

	ctx = ctxutil.WithNotifications(ctx, false)

	store := apimock.New()
	require.NotNil(t, store)
	assert.NoError(t, populateStore(ctx, store, secrets))

	for _, request := range requests {
		var inbuf bytes.Buffer
		var outbuf bytes.Buffer

		api := API{
			store,
			&inbuf,
			&outbuf,
			semver.MustParse("1.2.3-test"),
		}

		_, err := inbuf.Write([]byte(request.InputStr))
		assert.NoError(t, err)

		err = api.ReadAndRespond(ctx)
		if len(request.ErrorStr) > 0 {
			require.Error(t, err)
			assert.Equal(t, len(outbuf.String()), 0)
			continue
		}
		assert.NoError(t, err)
		outputMessage := readAndVerifyMessageLength(t, outbuf.Bytes())
		assert.NotEqual(t, "", request.OutputRegexpStr, "Empty string would match any output")
		assert.Regexp(t, regexp.MustCompile(request.OutputRegexpStr), outputMessage)
	}
}

func populateStore(ctx context.Context, s gopass.Store, secrets []storedSecret) error {
	for _, sec := range secrets {
		if err := s.Set(ctx, strings.Join(sec.Name, "/"), sec.Secret); err != nil {
			return err
		}
	}
	return nil
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
