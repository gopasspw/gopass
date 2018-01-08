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
	assert.NoError(t, sec.SetBody("---\nkey: value\n"))

	// get existing key
	val, err := sec.Value("key")
	assert.NoError(t, err)
	assert.Equal(t, "value", val)

	// get non-existing key
	_, err = sec.Value("some-key")
	assert.EqualError(t, err, store.ErrYAMLNoKey.Error())

	// delete existing key
	assert.NoError(t, sec.DeleteKey("key"))

	// delete non-existing key
	assert.NoError(t, sec.DeleteKey("some-key"))

	// set invalid YAML
	assert.Error(t, sec.SetBody("---\nkey-only\n"))
	assert.Equal(t, "---\nkey-only\n", sec.Body())

	// non-YAML body
	assert.NoError(t, sec.SetBody("key-only\n"))

	// try to set value on non-YAML body
	assert.EqualError(t, sec.SetValue("key", "value"), store.ErrYAMLNoMark.Error())

	// delete non-existing key
	assert.EqualError(t, sec.DeleteKey("some-key"), store.ErrYAMLNoMark.Error())
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
		assert.Equal(t, tc.eq, tc.s1.Equal(tc.s2))
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
