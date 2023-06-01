package secrets

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
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
	assert.Equal(t, "", v)

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
		assert.Equal(t, "", v)

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
		tc := tc

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
				assert.NoError(t, a.Add("hello", "berlin"))
				assert.Equal(t, true, a.Del("foo"))

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
				assert.NoError(t, a.Add("hello", "berlin"))
				assert.NoError(t, a.Set("bar", "foo"))
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

				assert.NoError(t, a.Set("foo", "bar"))
			},
			out: `
foo: bar
`,
		},
	} {
		tc := tc

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

	assert.NoError(t, a.Set("foo", "baz"))
	assert.NoError(t, a.Set("hello", "mars"))

	assert.Equal(t, "this is the body\nmore text\n", a.Body())

	_, err := a.Write([]byte("even more text\n"))
	assert.NoError(t, err)

	assert.Equal(t, "this is the body\nmore text\neven more text\n", a.Body())
}

func FuzzParseAKV(f *testing.F) {
	f.Fuzz(func(t *testing.T, in []byte) {
		ParseAKV(in)
	})
}
