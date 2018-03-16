package tpl

import (
	"context"
	"testing"

	"github.com/justwatchcom/gopass/pkg/store"
	"github.com/justwatchcom/gopass/pkg/store/secret"
)

type kvMock struct{}

func (k kvMock) Get(ctx context.Context, key string) (store.Secret, error) {
	return secret.New("barfoo", "---\nbarkey: barvalue\n"), nil
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
		buf, err := Execute(ctx, tc.Template, tc.Name, tc.Content, kv)
		if err != nil && !tc.ShouldFail {
			t.Fatalf("[%s] Failed to execute template %s: %s", tc.Template, tc.Template, err)
		}
		if err == nil && tc.ShouldFail {
			t.Errorf("[%s] Should fail", tc.Template)
		}
		if string(buf) != tc.Output {
			t.Errorf("[%s] Wrong templated output %s vs %s", tc.Template, string(buf), tc.Output)
		}
	}
}
