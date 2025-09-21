package tests

import (
	"testing"

	_ "github.com/gopasspw/gopass/internal/backend/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIssue2915(t *testing.T) {
	ts := newTester(t)
	defer ts.teardown()

	// create two GPG keys
	key1, err := ts.addFakeGPGKey("key1@example.com")
	require.NoError(t, err)
	key2, err := ts.addFakeGPGKey("key2@example.com")
	require.NoError(t, err)

	// initialize the root store with the first key
	ts.initStoreWithRecipients(key1)

	// initialize a sub-store with the second key
	_, err = ts.run("init -p sub --crypto gpg " + key2)
	require.NoError(t, err)

	// insert a secret into the root store
	_, err = ts.runCmd([]string{ts.Binary, "insert", "foo"}, []byte("bar"))
	require.NoError(t, err)

	// check the recipients of the created secret
	out, err := ts.run("recipients")
	require.NoError(t, err)

	// assert that the secret is only encrypted for the first key
	assert.Contains(t, out, key1)
	assert.NotContains(t, out, key2)
}
