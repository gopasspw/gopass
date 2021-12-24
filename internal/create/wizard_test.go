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
  authority:
    type: "string"
    prompt: "Authority"
    min: 1
  application:
    type: "string"
    prompt: "Entity"
    min: 1
  password:
    type: "password"
    prompt: "Pin"
    charset: "0123456789"
    min: 1
    max: 64
  comment:
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
	if w.Templates[0].Attributes["url"].Type != "hostname" {
		t.Fatal("wrong type")
	}
	if w.Templates[0].Attributes["url"].Prompt != "Website name" {
		t.Fatal("wrong prompt")
	}
	if w.Templates[0].Attributes["url"].Min != 1 {
		t.Fatal("wrong min")
	}
	if w.Templates[0].Attributes["url"].Max != 255 {
		t.Fatal("wrong max")
	}
	if w.Templates[0].Attributes["username"].Type != "string" {
		t.Fatal("wrong type")
	}
	if w.Templates[0].Attributes["username"].Prompt != "Login" {
		t.Fatal("wrong prompt")
	}
	if w.Templates[0].Attributes["username"].Min != 1 {
		t.Fatal("wrong min")
	}
	if w.Templates[0].Attributes["password"].Type != "password" {
		t.Fatal("wrong type")
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
