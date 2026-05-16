package action

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v3"
)

func TestParseArgs(t *testing.T) {
	for _, tc := range []struct {
		name   string
		argIn  []string
		argOut argList
		kvOut  map[string]string
	}{
		{
			name: "no args",
		},
		{
			name:   "secret",
			argIn:  []string{"foo/bar"},
			argOut: argList{"foo/bar"},
		},
		{
			name:   "secret with colon",
			argIn:  []string{"foo/bar:test"},
			argOut: argList{"foo/bar:test"},
		},
		{
			name:   "with key",
			argIn:  []string{"foo/bar", "baz"},
			argOut: argList{"foo/bar", "baz"},
		},
		{
			name:   "with k/v (=)",
			argIn:  []string{"foo/bar", "baz=bam"},
			argOut: argList{"foo/bar"},
			kvOut:  map[string]string{"baz": "bam"},
		},
		{
			name:   "with k/v (:)",
			argIn:  []string{"foo/bar", "baz:bam"},
			argOut: argList{"foo/bar"},
			kvOut:  map[string]string{"baz": "bam"},
		},
		{
			name:   "with k/v (mixed)",
			argIn:  []string{"foo/bar", "baz:bam", "foo=zen"},
			argOut: argList{"foo/bar"},
			kvOut:  map[string]string{"baz": "bam", "foo": "zen"},
		},
		{
			name:   "with k/v (mixed order)",
			argIn:  []string{"foo:bar", "foo/bar", "baz:bam"},
			argOut: argList{"foo/bar"},
			kvOut:  map[string]string{"foo": "bar", "baz": "bam"},
		},
		{
			name:   "with k/v (=) key and length",
			argIn:  []string{"foo/bar", "baz=bam", "baz", "42"},
			argOut: argList{"foo/bar", "baz", "42"},
			kvOut:  map[string]string{"baz": "bam"},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if tc.argOut == nil {
				tc.argOut = argList{}
			}

			if tc.kvOut == nil {
				tc.kvOut = map[string]string{}
			}

			var gotArgs argList
			var gotKVPs map[string]string

			cmd := &cli.Command{
				Action: func(c context.Context, cmd *cli.Command) error {
					gotArgs, gotKVPs = parseArgs(c, cmd)

					return nil
				},
			}
			require.NoError(t, cmd.Run(context.Background(), append([]string{"test"}, tc.argIn...)), tc.name)
			assert.Equal(t, tc.argOut, gotArgs, tc.name)
			assert.Equal(t, tc.kvOut, gotKVPs, tc.name)
		})
	}
}
