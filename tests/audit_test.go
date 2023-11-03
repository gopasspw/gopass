package tests

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAudit(t *testing.T) {
	ts := newTester(t)
	defer ts.teardown()

	ts.initStore()
	ts.initSecrets("")

	t.Run("audit the test store", func(t *testing.T) {
		out, err := ts.run("audit")
		require.Error(t, err)
		assert.Contains(t, out, "Password is too short")
		assert.Contains(t, out, "weak password")
	})
}
