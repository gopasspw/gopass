package store

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidateSecretName(t *testing.T) {
	t.Parallel()

	valid := []string{
		"foo",
		"foo/bar",
		"a/b/c",
		".create/pin.yml",
		"a.b/c.d",
		"foo..bar",
	}
	for _, name := range valid {
		t.Run("valid/"+name, func(t *testing.T) {
			t.Parallel()
			require.NoError(t, ValidateSecretName(name))
		})
	}

	invalid := []string{
		"",
		"/",
		"foo/",
		"/foo",
		"foo//bar",
		"..",
		"foo/..",
		"../foo",
		"foo/../bar",
		".",
		"foo/.",
		"./foo",
		"foo/./bar",
	}
	for _, name := range invalid {
		t.Run("invalid/"+name, func(t *testing.T) {
			t.Parallel()
			require.Error(t, ValidateSecretName(name))
		})
	}
}
