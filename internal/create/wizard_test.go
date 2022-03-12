package create

import (
	"context"
	"testing"

	"github.com/gopasspw/gopass/internal/store/mockstore/inmem"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
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
	require.Len(t, w.Templates, 3, "wrong number of templates")

	assert.Equal(t, "websites", w.Templates[0].Prefix, "wrong prefix")
	assert.Equal(t, "🧪 Creating Website login", w.Templates[0].Welcome, "wrong welcome")
	assert.Equal(t, 3, len(w.Templates[0].Attributes), "wrong number of attributes")
	assert.Equal(t, "hostname", w.Templates[0].Attributes[0].Type, "wrong type")
	assert.Equal(t, "Website URL", w.Templates[0].Attributes[0].Prompt, "wrong prompt")
	assert.Equal(t, 1, w.Templates[0].Attributes[0].Min, "wrong min")
	assert.Equal(t, 255, w.Templates[0].Attributes[0].Max, "wrong max")
	assert.Equal(t, "string", w.Templates[0].Attributes[1].Type, "wrong type")
	assert.Equal(t, "Login", w.Templates[0].Attributes[1].Prompt, "wrong prompt")
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
