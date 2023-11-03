package tests

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInsert(t *testing.T) {
	ts := newTester(t)
	defer ts.teardown()

	ts.initStore()

	out, err := ts.run("insert")
	require.Error(t, err)
	assert.Equal(t, "\nError: Usage: "+filepath.Base(ts.Binary)+" insert name\n", out)

	_, err = ts.runCmd([]string{ts.Binary, "insert", "some/secret"}, []byte("moar"))
	require.NoError(t, err)

	_, err = ts.runCmd([]string{ts.Binary, "insert", "some/newsecret"}, []byte("and\nmoar"))
	require.NoError(t, err)

	t.Run("Regression test for #1573 without actual pipes", func(t *testing.T) {
		out, err = ts.run("show -f some/secret")
		require.NoError(t, err)
		assert.Equal(t, "moar", out)

		out, err = ts.run("show -f some/newsecret")
		require.NoError(t, err)
		assert.Equal(t, "and\nmoar", out)

		out, err = ts.run("show -f some/secret")
		require.NoError(t, err)
		assert.Equal(t, "moar", out)

		out, err = ts.run("show -f some/newsecret")
		require.NoError(t, err)
		assert.Equal(t, "and\nmoar", out)
	})

	t.Run("Regression test for #1595", func(t *testing.T) {
		t.Skip("TODO")

		_, err = ts.runCmd([]string{ts.Binary, "insert", "some/other"}, []byte("nope"))
		require.NoError(t, err)

		out, err = ts.run("insert some/other")
		require.Error(t, err)
		assert.Equal(t, "\nError: not overwriting your current secret\n", out)

		out, err = ts.run("show -o some/other")
		require.NoError(t, err)
		assert.Equal(t, "nope", out)

		out, err = ts.run("--yes insert some/other")
		require.NoError(t, err)
		assert.Equal(t, "Warning: Password is empty or all whitespace", out)

		out, err = ts.run("insert -f some/other")
		require.NoError(t, err)
		assert.Equal(t, "Warning: Password is empty or all whitespace", out)

		out, err = ts.run("show -o some/other")
		require.Error(t, err)
		assert.Equal(t, "\nError: empty secret\n", out)

		_, err = ts.runCmd([]string{ts.Binary, "insert", "-f", "some/other"}, []byte("final"))
		require.NoError(t, err)

		out, err = ts.run("show -o some/other")
		require.NoError(t, err)
		assert.Equal(t, "final", out)

		// This is arguably not a good behaviour: it should not overwrite the password when we are only working on a key:value.
		out, err = ts.run("insert -f some/other test:inline")
		require.NoError(t, err)
		assert.Equal(t, "", out)

		out, err = ts.run("show some/other test")
		require.NoError(t, err)
		assert.Equal(t, "inline", out)

		out, err = ts.run("insert some/other test:inline2")
		require.Error(t, err)
		assert.Equal(t, "\nError: not overwriting your current secret\n", out)

		out, err = ts.run("show some/other Test")
		require.NoError(t, err)
		assert.Equal(t, "inline", out)

		out, err = ts.run("--yes insert some/other test:inline2")
		require.NoError(t, err)
		assert.Equal(t, "", out)

		out, err = ts.run("show some/other Test")
		require.NoError(t, err)
		assert.Equal(t, "inline2", out)
	})

	t.Run("Regression test for #1650 with JSON", func(t *testing.T) {
		json := `Password: SECRET
--
glossary": {
    "title": "example glossary",
    "GlossDiv": {
        "title": "S",
        "GlossList": {
            "GlossEntry": {
                "ID": "SGML",
                "SortAs": "SGML",
                "GlossTerm": "Standard Generalized Markup Language",
                "Acronym": "SGML",
                "Abbrev": "ISO 8879:1986",
                "GlossDef": {
                    "para": "A meta-markup language, used to create markup languages such as DocBook.",
                    "GlossSeeAlso": ["GML", "XML"]
                },
                "GlossSee": "markup"
            }
        }
    }
}`
		_, err = ts.runCmd([]string{ts.Binary, "insert", "some/json"}, []byte(json))
		require.NoError(t, err)

		// using show -n to disable parsing
		out, err = ts.run("show -f -n some/json")
		require.NoError(t, err)
		assert.Equal(t, json, out)
	})

	t.Run("Regression test for #1600", func(t *testing.T) {
		input := `test1
test2
{
  "Creator": "the creator"
}`
		_, err = ts.runCmd([]string{ts.Binary, "insert", "some/multilinewithbraces"}, []byte(input))
		require.NoError(t, err)

		// using show -n to disable parsing
		out, err = ts.run("show -f -n some/multilinewithbraces")
		require.NoError(t, err)
		assert.Equal(t, input, out)
	})

	t.Run("Regression test for #1601", func(t *testing.T) {
		input := `thepassword
user: a user
web: test.com
user: second user`

		_, err = ts.runCmd([]string{ts.Binary, "insert", "some/multikey"}, []byte(input))
		require.NoError(t, err)

		// using show -n to disable parsing
		out, err = ts.run("show -f -n some/multikey")
		require.NoError(t, err)
		assert.Equal(t, input, out)
	})

	t.Run("Regression test full support for #1601", func(t *testing.T) {
		t.Skip("Skipping until we support actual key-valueS for KV")

		input := `thepassword
user: a user
web: test.com
user: second user`

		output := `thepassword
web: test.com
user: a user
user: second user`

		_, err = ts.runCmd([]string{ts.Binary, "insert", "some/multikeyvalues"}, []byte(input))
		require.NoError(t, err)

		out, err = ts.run("show -f some/multikeyvalues")
		require.NoError(t, err)
		assert.Equal(t, output, out)
	})

	t.Run("Regression test for #1614", func(t *testing.T) {
		input := `yamltest
---
user: 0123`

		output := `yamltest
---
user: 83`

		_, err = ts.runCmd([]string{ts.Binary, "insert", "some/yamloctal"}, []byte(input))
		require.NoError(t, err)

		// with parsing we have 0123 interpreted as octal for 83
		out, err = ts.run("show -f some/yamloctal")
		require.NoError(t, err)
		assert.Equal(t, output, out)

		// using show -n to disable parsing
		out, err = ts.run("show -f -n some/yamloctal")
		require.NoError(t, err)
		assert.Equal(t, input, out)
	})

	t.Run("Regression test for #1594", func(t *testing.T) {
		input := `somepasswd
---
Test / test.com
user:myuser
url: test.com/`

		_, err = ts.runCmd([]string{ts.Binary, "insert", "some/kvwithspace"}, []byte(input))
		require.NoError(t, err)

		out, err = ts.run("show -f some/kvwithspace")
		require.NoError(t, err)
		assert.Equal(t, input, out)

		out, err = ts.run("show -f some/kvwithspace url")
		require.NoError(t, err)
		assert.Equal(t, "test.com/", out)

		out, err = ts.run("show -f some/kvwithspace user")
		require.Error(t, err)
	})
}
