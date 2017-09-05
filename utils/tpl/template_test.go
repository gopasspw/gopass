package tpl

import (
	"context"
	"testing"

	"github.com/justwatchcom/gopass/store/secret"
)

type kvMock struct{}

func (k kvMock) Get(ctx context.Context, key string) (*secret.Secret, error) {
	return secret.New("barfoo", ""), nil
}

func TestVars(t *testing.T) {
	ctx := context.Background()

	kv := kvMock{}
	for _, tc := range []struct {
		Template string
		Name     string
		Content  []byte
		Output   string
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
			Template: `md5{{(print .Content .Name) | md5sum}}`,
			Name:     "testdir",
			Content:  []byte("foobar"),
			Output:   "md55d952fb5e2b5c6258b044a663518349f",
		},
	} {
		buf, err := Execute(ctx, tc.Template, tc.Name, tc.Content, kv)
		if err != nil {
			t.Fatalf("Failed to execute template %s: %s", tc.Template, err)
		}
		if string(buf) != tc.Output {
			t.Errorf("Wrong templated output %s vs %s", string(buf), tc.Output)
		}
	}
}
