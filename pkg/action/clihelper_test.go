package action

import (
	"flag"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v2"
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
		if tc.argOut == nil {
			tc.argOut = argList{}
		}
		if tc.kvOut == nil {
			tc.kvOut = map[string]string{}
		}
		app := cli.NewApp()
		fs := flag.NewFlagSet("default", flag.ContinueOnError)
		assert.NoError(t, fs.Parse(tc.argIn), tc.name)
		args, kvps := parseArgs(cli.NewContext(app, fs, nil))
		assert.Equal(t, tc.argOut, args, tc.name)
		assert.Equal(t, tc.kvOut, kvps, tc.name)
	}
}
