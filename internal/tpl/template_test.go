package tpl

import (
	"context"
	"testing"

	"github.com/gopasspw/gopass/pkg/gopass"
	"github.com/gopasspw/gopass/pkg/gopass/secrets/secparse"
	"github.com/stretchr/testify/assert"
)

type kvMock struct{}

func (k kvMock) Get(ctx context.Context, key string) (gopass.Secret, error) {
	return secparse.Parse([]byte("barfoo\n---\nbarkey: barvalue\n"))
}

func TestVars(t *testing.T) {
	ctx := context.Background()

	kv := kvMock{}
	for _, tc := range []struct {
		Template   string
		Name       string
		Content    []byte
		Output     string
		ShouldFail bool
	}{
		{
			Template: "{{.Dir}}",
			Name:     "testdir",
			Content:  []byte("foobar"),
			Output:   ".",
		},
		{
			Template: "{{.Path}}",
			Name:     "testdir",
			Content:  []byte("foobar"),
			Output:   "testdir",
		},
		{
			Template: "{{.Name}}",
			Name:     "testdir",
			Content:  []byte("foobar"),
			Output:   "testdir",
		},
		{
			Template: "{{.Content}}",
			Name:     "testdir",
			Content:  []byte("foobar"),
			Output:   "foobar",
		},
		{
			Template: "{{.Content | md5sum}}",
			Name:     "testdir",
			Content:  []byte("foobar"),
			Output:   "3858f62230ac3c915f300c664312c63f",
		},
		{
			Template: "{{.Content | sha1sum}}",
			Name:     "testdir",
			Content:  []byte("foobar"),
			Output:   "8843d7f92416211de9ebb963ff4ce28125932878",
		},
		{
			Template: `{{getpw "testdir"}}`,
			Name:     "testdir",
			Content:  []byte("foobar"),
			Output:   "barfoo",
		},
		{
			Template: `{{get "testdir"}}`,
			Name:     "testdir",
			Content:  []byte("foobar"),
			Output:   "barfoo\n---\nbarkey: barvalue\n",
		},
		{
			Template: `{{getpw "testdir"}}`,
			Name:     "testdir",
			Content:  []byte("foobar"),
			Output:   "barfoo",
		},
		{
			Template: `{{getval "testdir" "barkey"}}`,
			Name:     "testdir",
			Content:  []byte("foobar"),
			Output:   "barvalue",
		},
		{
			Template: `md5{{(print .Content .Name) | md5sum}}`,
			Name:     "testdir",
			Content:  []byte("foobar"),
			Output:   "md55d952fb5e2b5c6258b044a663518349f",
		},
		{
			Template:   `{{|}}`,
			Name:       "testdir",
			Content:    []byte("foobar"),
			Output:     "",
			ShouldFail: true,
		},
	} {
		tc := tc
		t.Run(tc.Template, func(t *testing.T) {
			t.Parallel()
			buf, err := Execute(ctx, tc.Template, tc.Name, tc.Content, kv)
			if tc.ShouldFail {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tc.Output, string(buf), tc.Template)
		})
	}
}
