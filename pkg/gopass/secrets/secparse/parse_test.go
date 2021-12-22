package secparse

import (
	"fmt"
	"testing"

	"github.com/gopasspw/gopass/pkg/gopass/secrets"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParse(t *testing.T) {
	for _, tc := range []string{
		"foo\n",                                  // Plain
		"foo\nbar\n",                             // Plain
		"foo\nbar: baz\n",                        // KV
		"foo\nbar\n---\nzab: 1\n",                // YAML
		secrets.Ident + "\nFoo: Bar\n\nBarfoo\n", // MIME
		secrets.Ident + "\nFoo: Bar\n\nBarfoo",   // MIME
	} {
		_, err := Parse([]byte(tc))
		require.NoError(t, err)
	}
}

func TestParsedIsSerialized(t *testing.T) {
	for _, tc := range []string{
		"foo",                     // Plain
		"foo\nbar",                // Plain
		"foo\nbar: baz",           // KV
		"foo\nbar\n---\nzab: 1\n", // YAML
		// MIME is forcefully converted to KV
	} {
		sec, err := Parse([]byte(tc))
		require.NoError(t, err)
		fmt.Println()
		assert.Equal(t, tc, string(sec.Bytes()))
	}
}

func FuzzParse(f *testing.F) {
	for _, tc := range []string{
		"foo\n",                                  // Plain
		"foo\nbar\n",                             // Plain
		"foo\nbar: baz\n",                        // KV
		"foo\nbar\n---\nzab: 1\n",                // YAML
		secrets.Ident + "\nFoo: Bar\n\nBarfoo\n", // MIME
		secrets.Ident + "\nFoo: Bar\n\nBarfoo",   // MIME
	} {
		f.Add(tc)
	}
	f.Fuzz(func(t *testing.T, in string) {
		sec, err := Parse([]byte(in))
		if err != nil {
			t.Fatalf("Parse failed to decode a valid secret %q: %v", in, err)
		}
		if sec == nil {
			t.Errorf("secret should not be nil")
		}
	})
}
