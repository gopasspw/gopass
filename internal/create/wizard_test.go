package create

import (
	"context"
	"testing"

	"github.com/gopasspw/gopass/internal/config"
	"github.com/gopasspw/gopass/internal/store/mockstore/inmem"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeSetter struct{}

func (f *fakeSetter) Set(ctx context.Context, name string, content []byte) error {
	return nil
}

func (f *fakeSetter) Add(ctx context.Context, args ...string) error {
	return nil
}

func (f *fakeSetter) TryAdd(ctx context.Context, args ...string) error {
	return nil
}

func (f *fakeSetter) Commit(ctx context.Context, msg string) error {
	return nil
}

func (f *fakeSetter) TryCommit(ctx context.Context, msg string) error {
	return nil
}

func TestWrite(t *testing.T) {
	t.Parallel()

	ctx := config.NewNoWrites().WithConfig(context.Background())
	w := &Wizard{}

	require.NoError(t, w.writeTemplates(ctx, &fakeSetter{}))
}

func TestNew(t *testing.T) {
	t.Parallel()

	ctx := config.NewNoWrites().WithConfig(context.Background())
	s := inmem.New()
	_ = s.Set(ctx, ".create/pin.yml", []byte(`---
priority: 1
name: "PIN Code (numerical)"
prefix: "pin"
name_from:
  - "authority"
  - "application"
welcome: "🧪 Creating numerical PIN"
attributes:
  - name: "authority"
    type: "string"
    prompt: "Authority"
    min: 1
  - name: "application"
    type: "string"
    prompt: "Entity"
    min: 1
  - name: "password"
    type: "password"
    prompt: "Pin"
    charset: "0123456789"
    min: 1
    max: 64
  - name: "comment"
    type: "string"
`))
	w, err := New(ctx, s)
	require.NoError(t, err)
	require.NotNil(t, w.Templates, "no templates")
	require.Len(t, w.Templates, 1, "wrong number of templates")

	assert.Equal(t, "pin", w.Templates[0].Prefix, "wrong prefix")
	assert.Equal(t, "🧪 Creating numerical PIN", w.Templates[0].Welcome, "wrong welcome")
	assert.Len(t, w.Templates[0].Attributes, 4, "wrong number of attributes")
	assert.Equal(t, "string", w.Templates[0].Attributes[0].Type, "wrong type")
	assert.Equal(t, "Authority", w.Templates[0].Attributes[0].Prompt, "wrong prompt")
	assert.Equal(t, 1, w.Templates[0].Attributes[0].Min, "wrong min")
	assert.Equal(t, 0, w.Templates[0].Attributes[0].Max, "wrong max")
	assert.Equal(t, "string", w.Templates[0].Attributes[1].Type, "wrong type")
	assert.Equal(t, "Entity", w.Templates[0].Attributes[1].Prompt, "wrong prompt")
	assert.Equal(t, 1, w.Templates[0].Attributes[1].Min, "wrong min")
	assert.Equal(t, "password", w.Templates[0].Attributes[2].Type, "wrong type")
}

func TestExtractHostname(t *testing.T) {
	t.Parallel()

	for in, out := range map[string]string{
		"":                                     "",
		"http://www.example.org/":              "www.example.org",
		"++#+++#jhlkadsrezu 33 553q ++++##$§&": "jhlkadsrezu_33_553q",
		"www.example.org/?foo=bar#abc":         "www.example.org",
		"a test":                               "a_test",
	} {
		assert.Equal(t, out, extractHostname(in))
	}
}
