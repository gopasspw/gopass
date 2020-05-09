package simple

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFile(t *testing.T) {
	for _, tc := range []struct {
		File   File
		String string
		Format string
		Last   bool
	}{
		{
			File: File{
				Name: "foo",
			},
			String: "foo",
			Format: "├── foo\n",
			Last:   false,
		},
		{
			File: File{
				Name: "foo",
			},
			String: "foo",
			Format: "└── foo\n",
			Last:   true,
		},
	} {
		assert.Equal(t, tc.File.String(), tc.String)
		assert.Equal(t, tc.File.format("", tc.Last, 1, 1), tc.Format)
	}
}
