package create

import (
	"context"
	"testing"

	"github.com/gopasspw/gopass/internal/config"
	"github.com/gopasspw/gopass/internal/store/mockstore/inmem"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.yaml.in/yaml/v3"
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

	ctx := config.NewContextInMemory()
	w := &Wizard{}

	require.NoError(t, w.writeTemplates(ctx, &fakeSetter{}))
}

func TestNew(t *testing.T) {
	t.Parallel()

	ctx := config.NewContextInMemory()
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
    always_prompt: true
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
	assert.True(t, w.Templates[0].Attributes[2].AlwaysPrompt, "wrong always_prompt")
}

func TestDefaultTemplates(t *testing.T) {
	t.Parallel()

	ctx := config.NewContextInMemory()
	w := &Wizard{}

	tpls, err := w.parseTemplatesFallback(ctx)
	require.NoError(t, err)
	require.Len(t, tpls, len(defaultTemplates))

	// find the SSO / passwordless template and validate it models a
	// passwordless account: no password attribute, plus a login-method choice.
	var sso *Template
	for i := range tpls {
		if tpls[i].Prefix == "websites" && len(tpls[i].Attributes) > 0 {
			for _, a := range tpls[i].Attributes {
				if a.Name == "login-method" {
					sso = &tpls[i]

					break
				}
			}
		}
		if sso != nil {
			break
		}
	}
	require.NotNil(t, sso, "no SSO / passwordless template found")

	for _, a := range sso.Attributes {
		assert.NotEqual(t, "password", a.Type, "SSO template must not contain a password attribute")
	}

	var lm *Attribute
	for i := range sso.Attributes {
		if sso.Attributes[i].Name == "login-method" {
			lm = &sso.Attributes[i]

			break
		}
	}
	require.NotNil(t, lm)
	assert.Equal(t, "choice", lm.Type, "login-method must be a choice")
	assert.Contains(t, lm.Options, "google", "login-method must offer google")
	assert.NotEmpty(t, lm.Options, "choice must have options")
}

func TestOptionalPasswordParses(t *testing.T) {
	t.Parallel()

	tpl := Template{}
	require.NoError(t, yaml.Unmarshal([]byte(`---
name: "optional pw"
prefix: "opt"
attributes:
  - name: "password"
    type: "password"
    optional: true
`), &tpl))
	require.Len(t, tpl.Attributes, 1)
	assert.True(t, tpl.Attributes[0].Optional, "optional flag must round-trip")
}
