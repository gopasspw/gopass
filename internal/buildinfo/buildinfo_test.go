package buildinfo

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestModuleVersion(t *testing.T) {
	orig := Version
	t.Cleanup(func() {
		Version = orig
	})

	// the stash set by main is authoritative
	Version = "1.2.3"
	v, ok := ModuleVersion()
	require.True(t, ok)
	require.Equal(t, "1.2.3", v)

	// without a stash the embedded build info decides; the exact value
	// depends on how the test binary was built, so only pin the contract
	Version = ""
	v, ok = ModuleVersion()
	if ok {
		require.NotEmpty(t, v)
	} else {
		require.Empty(t, v)
	}
}
