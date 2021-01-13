package secparse

import (
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
		"foo\n",                   // Plain
		"foo\nbar\n",              // Plain
		"foo\nbar: baz\n",         // KV
		"foo\nbar\n---\nzab: 1\n", // YAML
		// MIME is forcefully converted to KV
	} {
		sec, err := Parse([]byte(tc))
		require.NoError(t, err)
		assert.Equal(t, tc, string(sec.Bytes()))
	}
}
