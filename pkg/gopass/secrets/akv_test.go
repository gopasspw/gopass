package secrets

import (
	"bufio"
	"crypto/rand"
	"encoding/base64"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAKV(t *testing.T) {
	t.Parallel()

	t.Logf("Retrieve content from invalid YAML (#375)")

	mlValue := `somepasswd
Test / test.com
username: myuser@test.com
url: http://www.test.com/
password: bar
`
	s := ParseAKV([]byte(mlValue))
	assert.NotNil(t, s)

	v, found := s.Get("Test / test.com")
	assert.False(t, found)
	assert.Empty(t, v)

	t.Logf("Secret:\n%+v\n%s\n", s, string(s.Bytes()))

	t.Run("read back the secret", func(t *testing.T) {
		assert.Equal(t, mlValue, string(s.Bytes()))
	})

	t.Run("no_duplicate_keys", func(t *testing.T) {
		assert.Equal(t, []string{"password", "url", "username"}, s.Keys())
	})

	t.Run("read some keys", func(t *testing.T) {
		for k, v := range map[string]string{
			"password": "bar",
			"url":      "http://www.test.com/",
			"username": "myuser@test.com",
		} {
			fv, found := s.Get(k)
			assert.True(t, found)
			assert.Equal(t, v, fv)
		}
		assert.Equal(t, "somepasswd", s.Password())
	})

	t.Run("remove a key", func(t *testing.T) {
		s.Del("username")
		v, ok := s.Get("username")
		assert.False(t, ok)
		assert.Empty(t, v)

		assert.Equal(t, `somepasswd
Test / test.com
url: http://www.test.com/
password: bar
`, string(s.Bytes()))
	})
}

func TestAKVNoNewLine(t *testing.T) {
	t.Parallel()

	mlValue := `foobar
ab: cd`
	s := ParseAKV([]byte(mlValue))
	assert.NotNil(t, s)
	v, _ := s.Get("ab")
	assert.Equal(t, "cd", v)
}

func TestMultiKeyAKVMIME(t *testing.T) {
	t.Parallel()

	in := `passw0rd
foo: baz
foo: bar
zab: 123
`

	sec := ParseAKV([]byte(in))
	assert.Equal(t, in, string(sec.Bytes()))
	assert.Equal(t, "passw0rd", sec.Password())
}

func TestMultilineInsertAKV(t *testing.T) {
	t.Parallel()

	in := `passw0rd
foo: baz
foo: bar
zab: 123
`

	sec := NewAKV()
	_, err := sec.Write([]byte(in))
	require.NoError(t, err)
	assert.Equal(t, in, string(sec.Bytes()))
	assert.Equal(t, "passw0rd", sec.Password())

	_, err = sec.Write([]byte("more text"))
	require.NoError(t, err)
	assert.Equal(t, "passw0rd", sec.Password())
}

func TestSetKeyValuePairToEmptyAKV(t *testing.T) {
	t.Parallel()

	sec := NewAKV()
	require.NoError(t, sec.Set("foo", "bar"))
	v, found := sec.Get("foo")
	assert.True(t, found)
	assert.Equal(t, "bar", v)
}

func TestAKVTrailingWhitespace(t *testing.T) {
	t.Parallel()
	// Expected behaviour is KEY: VALUE, with one space.
	// Fallback should exist for KEY:VALUE, with no spaces.
	mlValue := `foobar
defaultBehaviour: cd
sorroundedBySpace:   cd 	 
withoutSpace:cd`
	s := ParseAKV([]byte(mlValue))
	assert.NotNil(t, s)
	v1, _ := s.Get("defaultBehaviour")
	assert.Equal(t, "cd", v1)
	v2, _ := s.Get("sorroundedBySpace")
	assert.Equal(t, "  cd \t ", v2)
	v3, _ := s.Get("defaultBehaviour")
	assert.Equal(t, "cd", v3)
}

func TestAKVPasswordWhitespace(t *testing.T) {
	t.Parallel()

	helloIsWorldStr := "\nhello: world\n"
	helloIsWorldPair := map[string][]string{
		"hello": {"world"},
	}

	for _, tc := range []struct {
		name string
		in   string
		pw   string
		kvp  map[string][]string
	}{
		{
			name: "justpassword",
			in:   `this is a password.` + helloIsWorldStr,
			pw:   "this is a password.",
			kvp:  helloIsWorldPair,
		},
		{
			name: "spaceonly",
			in:   "   " + helloIsWorldStr,
			pw:   "   ",
			kvp:  helloIsWorldPair,
		},
		{
			name: "tab",
			in:   "\t tab padded password \t" + helloIsWorldStr,
			pw:   "\t tab padded password \t",
			kvp:  helloIsWorldPair,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			a := ParseAKV([]byte(tc.in))

			assert.Equal(t, tc.pw, a.password, tc.name)
			for k, vs := range tc.kvp {
				sort.Strings(vs)
				gvs := a.kvp[k]
				sort.Strings(gvs)
				assert.Equal(t, vs, gvs, k)
			}

			assert.Equal(t, tc.in, string(a.Bytes()), tc.name)
		})
	}
}

func TestParseAKV(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name string
		in   string
		pw   string
		kvp  map[string][]string
	}{
		{
			name: "simple",
			in: `foobar
hello: world

bla
bla
`,
			pw: "foobar",
			kvp: map[string][]string{
				"hello": {"world"},
			},
		},
		{
			name: "misc",
			in: `

lorem ipsum dolor sunt
hello: world
bla bla bla blabla
key: value1
key: value2

`,
			pw: "",
			kvp: map[string][]string{
				"hello": {"world"},
				"key":   {"value1", "value2"},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			a := ParseAKV([]byte(tc.in))

			assert.Equal(t, tc.pw, a.password, tc.name)
			for k, vs := range tc.kvp {
				sort.Strings(vs)
				gvs := a.kvp[k]
				sort.Strings(gvs)
				assert.Equal(t, vs, gvs, k)
			}

			assert.Equal(t, tc.in, string(a.Bytes()), tc.name)
		})
	}
}

func TestManipulateAKV(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name string
		in   string
		pw   string
		kvp  map[string][]string
		out  string
		op   func(t *testing.T, a *AKV)
	}{
		{
			name: "simple",
			in: `foobar
hello: world
foo: bar

bla
bla
`,
			pw: "barfoo",
			kvp: map[string][]string{
				"hello": {"berlin", "world"},
			},
			op: func(t *testing.T, a *AKV) {
				t.Helper()

				a.SetPassword("barfoo")
				require.NoError(t, a.Add("hello", "berlin"))
				assert.True(t, a.Del("foo"))

				assert.Equal(t, "barfoo", a.Password())
			},
			out: `barfoo
hello: world

bla
bla
hello: berlin
`,
		},
		{
			name: "set",
			in: `foobar
hello: world
foo: bar

bla
bla
`,
			pw: "barfoo",
			kvp: map[string][]string{
				"hello": {"berlin", "world"},
				"foo":   {"bar"},
				"bar":   {"foo"},
			},
			op: func(t *testing.T, a *AKV) {
				t.Helper()

				a.SetPassword("barfoo")
				require.NoError(t, a.Add("hello", "berlin"))
				require.NoError(t, a.Set("bar", "foo"))
			},
			out: `barfoo
hello: world
foo: bar

bla
bla
hello: berlin
bar: foo
`,
		},
		{
			name: "no-new-line",
			in: `foobar
hello: world
foo: bar

bla
bla`,
			pw: "foobar",
			kvp: map[string][]string{
				"hello": {"world"},
				"foo":   {"bar"},
			},
			out: `foobar
hello: world
foo: bar

bla
bla
`,
		},
		{
			name: "empty-set-key",
			in:   "",
			pw:   "",
			op: func(t *testing.T, a *AKV) {
				t.Helper()

				require.NoError(t, a.Set("foo", "bar"))
			},
			out: `
foo: bar
`,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			a := ParseAKV([]byte(tc.in))

			if tc.op != nil {
				tc.op(t, a)
			}

			assert.Equal(t, tc.pw, a.password, tc.name)
			for k, vs := range tc.kvp {
				sort.Strings(vs)
				gvs := a.kvp[k]
				sort.Strings(gvs)
				assert.Equal(t, vs, gvs, k)
			}

			want := tc.in
			if tc.out != "" {
				want = tc.out
			}
			assert.Equal(t, want, string(a.Bytes()), tc.name)
		})
	}
}

func TestNewAKV(t *testing.T) {
	a := NewAKVWithData("foobar", map[string][]string{
		"foo":   {"bar"},
		"hello": {"world", "everyone"},
	}, "this is the body\nmore text\n", false)

	assert.Equal(t, "foobar\nfoo: bar\nhello: world\nhello: everyone\nthis is the body\nmore text\n", a.raw.String())

	vs, ok := a.Values("foo")
	assert.True(t, ok)
	assert.Equal(t, []string{"bar"}, vs)

	require.NoError(t, a.Set("foo", "baz"))
	require.NoError(t, a.Set("hello", "mars"))

	assert.Equal(t, "this is the body\nmore text\n", a.Body())

	_, err := a.Write([]byte("even more text\n"))
	require.NoError(t, err)

	assert.Equal(t, "this is the body\nmore text\neven more text\n", a.Body())
}

func TestLargeBase64AKV(t *testing.T) {
	testSize := 100 * bufio.MaxScanTokenSize
	buf := make([]byte, testSize)
	n, err := rand.Read(buf)
	require.NoError(t, err)
	assert.Equal(t, testSize, n)

	sec := NewAKV()
	require.NoError(t, sec.Set("Content-Disposition", "attachment; filename=foo.bar"))
	require.NoError(t, sec.Set("Content-Transfer-Encoding", "Base64"))

	b64in := base64.StdEncoding.EncodeToString(buf) + "\n"
	n, err = sec.Write([]byte(b64in))
	require.NoError(t, err)
	assert.Len(t, b64in, n)

	b64out := sec.Body()
	assert.Equal(t, b64in, b64out)
}

func TestLargeBinaryAKV(t *testing.T) {
	t.Skip("TODO: AKV does not support transparent handling of non-text content, yet.")

	testSize := 2
	buf := make([]byte, testSize)
	n, err := rand.Read(buf)
	require.NoError(t, err)
	assert.Equal(t, testSize, n)

	sec := NewAKV()
	// This hack is required to make sure that the binary content does not end up
	// in the password field.
	_, err = sec.Write([]byte("\n"))
	require.NoError(t, err)

	n, err = sec.Write(buf)
	require.NoError(t, err)
	assert.Len(t, buf, n)

	out := sec.Body()
	assert.Equal(t, string(buf), out)
}

func FuzzParseAKV(f *testing.F) {
	f.Fuzz(func(t *testing.T, in []byte) {
		ParseAKV(in)
	})
}

func TestPwWriter(t *testing.T) {
	a := NewAKV()
	p := pwWriter{w: &a.raw, cb: func(pw string) { a.password = pw }}

	// multi-chunk passwords are supported
	_, err := p.Write([]byte("foo"))
	require.NoError(t, err)

	_, err = p.Write([]byte("bar\n"))
	require.NoError(t, err)

	// but anything after the first line is discarded
	_, err = p.Write([]byte("baz\n"))
	require.NoError(t, err)

	assert.Equal(t, "foobar", a.Password())
	assert.Equal(t, "baz\n", a.Body())
}

func TestInvalidPwWriter(t *testing.T) {
	defer func() {
		r := recover()
		assert.NotNil(t, r)
	}()
	p := pwWriter{}

	// will panic because the writer is nil
	_, err := p.Write([]byte("foo"))
	require.Error(t, err)
}
