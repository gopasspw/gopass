package secret

import (
	"testing"

	"github.com/justwatchcom/gopass/store"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	sec := New("foo", "---\nbar: baz\n")

	// change password
	sec.SetPassword("bar")
	if sec.Password() != "bar" {
		t.Errorf("Wrong password: %s", sec.Password())
	}

	// set valid YAML
	if err := sec.SetBody("---\nkey: value\n"); err != nil {
		t.Errorf("YAML Error: %s", err)
	}

	// get existing key
	val, err := sec.Value("key")
	if err != nil {
		t.Errorf("Error: %s", err)
	}
	if val != "value" {
		t.Errorf("Wrong value: %s", val)
	}

	// get non-existing key
	_, err = sec.Value("some-key")
	if err != store.ErrYAMLNoKey {
		t.Errorf("Should fail")
	}

	// delete existing key
	if err := sec.DeleteKey("key"); err != nil {
		t.Errorf("Error: %s", err)
	}

	// delete non-existing key
	if err := sec.DeleteKey("some-key"); err != nil {
		t.Errorf("Error: %s", err)
	}

	// set invalid YAML
	if err := sec.SetBody("---\nkey-only\n"); err == nil {
		t.Errorf("Should fail")
	}
	if sec.Body() != "---\nkey-only\n" {
		t.Errorf("Should contain invalid YAML despite parser failure")
	}

	// non-YAML body
	if err := sec.SetBody("key-only\n"); err != nil {
		t.Errorf("YAML Error: %s", err)
	}

	// try to set value on non-YAML body
	if err := sec.SetValue("key", "value"); err != store.ErrYAMLNoMark {
		t.Errorf("Should fail")
	}

	// delete non-existing key
	if err := sec.DeleteKey("some-key"); err != store.ErrYAMLNoMark {
		t.Errorf("Should fail")
	}
}

func TestEqual(t *testing.T) {
	for _, tc := range []struct {
		s1 *Secret
		s2 *Secret
		eq bool
	}{
		{
			s1: nil,
			s2: nil,
			eq: true,
		},
		{
			s1: New("foo", "bar"),
			s2: nil,
			eq: false,
		},
		{
			s1: New("foo", "bar"),
			s2: New("foo", "bar"),
			eq: true,
		},
		{
			s1: New("foo", "bar"),
			s2: New("foo", "baz"),
			eq: false,
		},
		{
			s1: New("foo", "bar"),
			s2: New("bar", "bar"),
			eq: false,
		},
		{
			s1: &Secret{
				password: "foo",
				data: map[string]interface{}{
					"key": &Secret{},
				},
			},
			s2: &Secret{
				password: "foo",
			},
			eq: false,
		},
	} {
		if tc.s1 != nil && tc.s1.data != nil {
			_ = tc.s1.encodeYAML()
		}
		if tc.s2 != nil && tc.s2.data != nil {
			_ = tc.s2.encodeYAML()
		}
		if tc.s1.Equal(tc.s2) != tc.eq {
			t.Errorf("s1 (%+v) and s2 (%+v) should be equal (%t) but are not", tc.s1, tc.s2, tc.eq)
		}
	}
}

func TestParse(t *testing.T) {
	for _, tc := range []struct {
		Desc     string
		In       []byte
		Out      []byte
		Password string
		Body     string
		Data     map[string]interface{}
		Fail     bool
	}{
		{
			Desc:     "Simple Secret",
			In:       []byte(`password`),
			Out:      []byte("password"),
			Password: "password",
		},
		{
			Desc: "Multiline secret",
			In: []byte(`password
hello world
this
is
some random
data`),
			Password: "password",
			Body: `hello world
this
is
some random
data`,
		},
		{
			Desc: "YAML Secret",
			In: []byte(`password
---
key1: value1
key2: value2`),
			Password: "password",
			Data: map[string]interface{}{
				"key1": "value1",
				"key2": "value2",
			},
			Body: `---
key1: value1
key2: value2`,
		},
		{
			Desc: "YAML only Secret",
			In: []byte(`
---
key1: value1
key2: value2`),
			Data: map[string]interface{}{
				"key1": "value1",
				"key2": "value2",
			},
			Body: `---
key1: value1
key2: value2`,
		},
		{
			Desc: "invalid YAML Secret",
			In: []byte(`password
---
	key1: value1
key2: value2`),
			Password: "password",
			Data: map[string]interface{}{
				"key1": "value1",
				"key2": "value2",
			},
			Fail: true,
		},
		{
			Desc: "missing YAML marker",
			In: []byte(`password
key1: value1
key2: value2`),
			Password: "password",
			Body: `key1: value1
key2: value2`,
		},
	} {
		sec, err := Parse(tc.In)
		if tc.Fail {
			assert.Error(t, err)
			continue
		} else if err != nil {
			assert.NoError(t, err)
			continue
		}
		assert.Equal(t, tc.Password, sec.Password())
		assert.Equal(t, tc.Body, sec.Body())
		for k, v := range tc.Data {
			rv, err := sec.Value(k)
			assert.NoError(t, err)
			assert.Equal(t, v, rv)
		}
		b, err := sec.Bytes()
		assert.NoError(t, err)
		if tc.Out != nil {
			assert.Equal(t, string(tc.Out), string(b))
		}
	}
}
