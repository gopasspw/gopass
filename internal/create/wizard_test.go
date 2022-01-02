package create

import (
	"context"
	"testing"

	"github.com/gopasspw/gopass/internal/store/mockstore/inmem"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	ctx := context.Background()
	s := inmem.New()
	s.Set(ctx, ".create/pin.yml", []byte(`---
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
	if err != nil {
		t.Fatal(err)
	}
	if w.Templates == nil {
		t.Fatal("no templates")
	}
	if len(w.Templates) != 3 {
		t.Fatal("wrong number of templates")
	}
	if w.Templates[0].Prefix != "websites" {
		t.Fatal("wrong prefix")
	}
	if w.Templates[0].Welcome != "🧪 Creating Website login" {
		t.Fatal("wrong welcome")
	}
	if l := len(w.Templates[0].Attributes); l != 3 {
		t.Fatalf("wrong number of attributes. want(%d), got(%d)", 3, l)
	}
	if w.Templates[0].Attributes[0].Type != "hostname" {
		t.Fatal("wrong type")
	}
	if w.Templates[0].Attributes[0].Prompt != "Website URL" {
		t.Fatal("wrong prompt:", w.Templates[0].Attributes[0].Prompt, "want:", "Website URL")
	}
	if w.Templates[0].Attributes[0].Min != 1 {
		t.Fatal("wrong min")
	}
	if w.Templates[0].Attributes[0].Max != 255 {
		t.Fatal("wrong max")
	}
	if w.Templates[0].Attributes[1].Type != "string" {
		t.Fatal("wrong type", w.Templates[0].Attributes[1].Type, "want:", "string")
	}
	if w.Templates[0].Attributes[1].Prompt != "Login" {
		t.Fatal("wrong prompt")
	}
	if w.Templates[0].Attributes[1].Min != 1 {
		t.Fatal("wrong min")
	}
	if w.Templates[0].Attributes[2].Type != "password" {
		t.Fatal("wrong type", w.Templates[0].Attributes[2].Type, "want:", "password")
	}
}

func TestExtractHostname(t *testing.T) {
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
