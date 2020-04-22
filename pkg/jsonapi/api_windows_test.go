package jsonapi

import (
	"testing"

	"github.com/gopasspw/gopass/pkg/store/secret"

	_ "github.com/gopasspw/gopass/pkg/backend/crypto"
	_ "github.com/gopasspw/gopass/pkg/backend/rcs"
	_ "github.com/gopasspw/gopass/pkg/backend/storage"
)

func TestRespondMessageQuery(t *testing.T) {
	secrets := []storedSecret{
		{[]string{"awesomePrefix", "foo", "bar"}, secret.New("20", "")},
		{[]string{"awesomePrefix", "fixed", "secret"}, secret.New("moar", "")},
		{[]string{"awesomePrefix", "fixed", "yamllogin"}, secret.New("thesecret", "---\nlogin: muh")},
		{[]string{"awesomePrefix", "fixed", "yamlother"}, secret.New("thesecret", "---\nother: meh")},
		{[]string{"awesomePrefix", "some.other.host", "there"}, secret.New("thesecret", "---\nother: meh")},
		{[]string{"awesomePrefix", "b", "some.other.host"}, secret.New("thesecret", "---\nother: meh")},
		{[]string{"awesomePrefix", "evilsome.other.host"}, secret.New("thesecret", "---\nother: meh")},
		{[]string{"evilsome.other.host", "something"}, secret.New("thesecret", "---\nother: meh")},
		{[]string{"awesomePrefix", "other.host", "there"}, secret.New("thesecret", "---\nother: meh")},
		{[]string{"somename", "github.com"}, secret.New("thesecret", "---\nother: meh")},
		{[]string{"login_entry"}, secret.New("thepass", `---
login: thelogin
ignore: me
login_fields:
  first: 42
  second: ok
nologin_fields:
  subentry: 123`)},
		{[]string{"invalid_login_entry"}, secret.New("thepass", `---
login: thelogin
ignore: me
login_fields: "invalid"`)},
	}

	// query for keys without any matching
	runRespondMessage(t,
		`{"type":"query","query":"notfound"}`,
		`\[\]`,
		"", secrets)

	// query for keys with matching one
	runRespondMessage(t,
		`{"type":"query","query":"foo"}`,
		`\["awesomePrefix\\\\foo\\\\bar"\]`,
		"", secrets)

	// query for keys with matching multiple
	runRespondMessage(t,
		`{"type":"query","query":"yaml"}`,
		`\["awesomePrefix\\\\fixed\\\\yamllogin","awesomePrefix\\\\fixed\\\\yamlother"\]`,
		"", secrets)

	// query for host
	runRespondMessage(t,
		`{"type":"queryHost","host":"find.some.other.host"}`,
		`\["awesomePrefix\\\\b\\\\some.other.host","awesomePrefix\\\\some.other.host\\\\there"\]`,
		"", secrets)

	// query for host not matches parent domain
	runRespondMessage(t,
		`{"type":"queryHost","host":"other.host"}`,
		`\["awesomePrefix\\\\other.host\\\\there"\]`,
		"", secrets)

	// query for host is query has different domain appended does not return partial match
	runRespondMessage(t,
		`{"type":"queryHost","host":"some.other.host.different.domain"}`,
		`\[\]`,
		"", secrets)

	// query returns result with public suffix at the end
	runRespondMessage(t,
		`{"type":"queryHost","host":"github.com"}`,
		`\["somename\\\\github.com"\]`,
		"", secrets)

	// get username / password for key without value in yaml
	runRespondMessage(t,
		`{"type":"getLogin","entry":"awesomePrefix\\fixed\\secret"}`,
		`{"username":"secret","password":"moar"}`,
		"", secrets)

	// get username / password for key with login in yaml
	runRespondMessage(t,
		`{"type":"getLogin","entry":"awesomePrefix\\fixed\\yamllogin"}`,
		`{"username":"muh","password":"thesecret"}`,
		"", secrets)

	// get username / password for key with no login in yaml (fallback)
	runRespondMessage(t,
		`{"type":"getLogin","entry":"awesomePrefix\\fixed\\yamlother"}`,
		`{"username":"yamlother","password":"thesecret"}`,
		"", secrets)

	// get entry with login fields
	runRespondMessage(t,
		`{"type":"getLogin","entry":"login_entry"}`,
		`{"username":"thelogin","password":"thepass","login_fields":{"first":42,"second":"ok"}}`,
		"", secrets)

	// get entry with invalid login fields
	runRespondMessage(t,
		`{"type":"getLogin","entry":"invalid_login_entry"}`,
		`{"username":"thelogin","password":"thepass"}`,
		"", secrets)
}
