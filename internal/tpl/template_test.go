package tpl

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/gopasspw/gopass/internal/pwschemes/argon2i"
	"github.com/gopasspw/gopass/internal/pwschemes/argon2id"
	"github.com/gopasspw/gopass/pkg/gopass"
	"github.com/gopasspw/gopass/pkg/gopass/secrets/secparse"
	"github.com/jsimonetti/pwscheme/md5crypt"
	"github.com/jsimonetti/pwscheme/ssha"
	"github.com/jsimonetti/pwscheme/ssha256"
	"github.com/jsimonetti/pwscheme/ssha512"
	"github.com/stretchr/testify/assert"
)

// TODO add an example func for the documentation

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
		OutputFunc func(string) error
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
			Template:   `{{getval "testdir" "barkeyINVALID"}}`,
			Name:       "testdir",
			Content:    []byte("foobar"),
			Output:     "",
			ShouldFail: true,
		},
		{
			Template: `{{getvals "testdir" "barkey"}}`,
			Name:     "testdir",
			Content:  []byte("foobar"),
			Output:   "[barvalue]",
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
		{
			Template: "{{.Content | ssha \"12\"}}",
			Name:     "testdir",
			Content:  []byte("foobar"),
			OutputFunc: func(s string) error {
				if !strings.HasPrefix(s, "{SSHA}") {
					return fmt.Errorf("wrong prefix")
				}
				ok, err := ssha.Validate("foobar", s)
				if err != nil {
					return fmt.Errorf("can't validate: %w", err)
				}
				if !ok {
					return fmt.Errorf("hash mismatch")
				}
				return nil
			},
		},
		{
			Template: "{{.Content | ssha256 \"12\"}}",
			Name:     "testdir",
			Content:  []byte("foobar"),
			OutputFunc: func(s string) error {
				if !strings.HasPrefix(s, "{SSHA256}") {
					return fmt.Errorf("wrong prefix: %s", s)
				}
				ok, err := ssha256.Validate("foobar", s)
				if err != nil {
					return fmt.Errorf("can't validate: %w", err)
				}
				if !ok {
					return fmt.Errorf("hash mismatch")
				}
				return nil
			},
		},
		{
			Template: "{{.Content | ssha512 \"12\"}}",
			Name:     "testdir",
			Content:  []byte("foobar"),
			OutputFunc: func(s string) error {
				if !strings.HasPrefix(s, "{SSHA512}") {
					return fmt.Errorf("wrong prefix: %s", s)
				}
				ok, err := ssha512.Validate("foobar", s)
				if err != nil {
					return fmt.Errorf("can't validate: %w", err)
				}
				if !ok {
					return fmt.Errorf("hash mismatch")
				}
				return nil
			},
		},
		{
			Template: "{{.Content | ssha512 \"-12\" \"invalid\"}}",
			Name:     "testdir",
			Content:  []byte("foobar"),
			OutputFunc: func(s string) error {
				if !strings.HasPrefix(s, "{SSHA512}") {
					return fmt.Errorf("wrong prefix: %s", s)
				}
				ok, err := ssha512.Validate("foobar", s)
				if err != nil {
					return fmt.Errorf("can't validate: %w", err)
				}
				if !ok {
					return fmt.Errorf("hash mismatch")
				}
				return nil
			},
		},
		{
			Template: "{{.Content | md5crypt \"7\"}}",
			Name:     "testdir",
			Content:  []byte("foobar"),
			OutputFunc: func(s string) error {
				if !strings.HasPrefix(s, "{MD5-CRYPT}") {
					return fmt.Errorf("wrong prefix: %s", s)
				}
				ok, err := md5crypt.Validate("foobar", s)
				if err != nil {
					return fmt.Errorf("can't validate: %w", err)
				}
				if !ok {
					return fmt.Errorf("hash mismatch")
				}
				return nil
			},
		},
		{
			Template: "{{.Content | md5crypt \"0\"}}",
			Name:     "testdir",
			Content:  []byte("foobar"),
			OutputFunc: func(s string) error {
				if !strings.HasPrefix(s, "{MD5-CRYPT}") {
					return fmt.Errorf("wrong prefix: %s", s)
				}
				ok, err := md5crypt.Validate("foobar", s)
				if err != nil {
					return fmt.Errorf("can't validate: %w", err)
				}
				if !ok {
					return fmt.Errorf("hash mismatch")
				}
				return nil
			},
		},
		{
			Template: "{{.Content | argon2i \"64\"}}",
			Name:     "testdir",
			Content:  []byte("foobar"),
			OutputFunc: func(s string) error {
				if !strings.HasPrefix(s, "{ARGON2I}") {
					return fmt.Errorf("wrong prefix: %s", s)
				}
				ok, err := argon2i.Validate("foobar", s)
				if err != nil {
					return fmt.Errorf("can't validate: %w", err)
				}
				if !ok {
					return fmt.Errorf("hash mismatch")
				}
				return nil
			},
		},
		{
			Template: "{{.Content | argon2id \"256\"}}",
			Name:     "testdir",
			Content:  []byte("foobar"),
			OutputFunc: func(s string) error {
				if !strings.HasPrefix(s, "{ARGON2ID}") {
					return fmt.Errorf("wrong prefix: %s", s)
				}
				ok, err := argon2id.Validate("foobar", s)
				if err != nil {
					return fmt.Errorf("can't validate: %w", err)
				}
				if !ok {
					return fmt.Errorf("hash mismatch")
				}
				return nil
			},
		},
		{
			Template:   "{{ argon2id }}",
			Name:       "testdir",
			Content:    []byte("foobar"),
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
			if tc.OutputFunc != nil && tc.Output != "" {
				t.Error("must not set output and output func")
			}
			if tc.OutputFunc != nil {
				assert.NoError(t, tc.OutputFunc(string(buf)), tc.Template)
			} else {
				assert.Equal(t, tc.Output, string(buf), tc.Template)
			}
		})
	}
}
