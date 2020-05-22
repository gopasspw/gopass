// +build !windows

package backend_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	yaml "gopkg.in/yaml.v2"
)

func TestUnmarshalYAML(t *testing.T) {
	in := `---
path: gpgcli-gitcli-fs+file:///tmp/foo
`
	cfg := testConfig{}
	require.NoError(t, yaml.Unmarshal([]byte(in), &cfg))
	assert.Equal(t, "/tmp/foo", cfg.Path.Path)
}
