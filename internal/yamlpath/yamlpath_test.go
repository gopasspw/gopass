package yamlpath

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestYAMLPath(t *testing.T) {
	data := map[string]interface{}{
		"these": map[string]interface{}{
			"are": map[string]interface{}{
				"not": map[string]interface{}{
					"the": map[string]interface{}{
						"droids": "you are looking for",
					},
				},
			},
		},
		"simple": "key",
		"array": []string{
			"foo",
			"bar",
		},
		"two": map[string]interface{}{
			"levels": "nested",
		},
	}

	for _, tc := range []struct {
		path string
		want string
	}{
		{
			path: "simple",
			want: "key",
		},
		{
			path: "these/are/not/the/droids",
			want: "you are looking for",
		},
		{
			path: "two/levels",
			want: "nested",
		},
		{
			path: "/two/levels",
			want: "nested",
		},
		{
			path: "/two/levels/",
			want: "nested",
		},
		{
			path: "two",
			want: "map[levels:nested]",
		},
	} {
		tc := tc
		t.Run(tc.path, func(t *testing.T) {
			out, err := YAMLPath(data, tc.path)
			outString := fmt.Sprintf("%v", out)
			assert.NoError(t, err)
			assert.Equal(t, tc.want, outString)
		})
	}
}
