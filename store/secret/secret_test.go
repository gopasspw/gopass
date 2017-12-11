package secret

import "testing"

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
			Out:      []byte("password\n"),
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
			if err == nil {
				t.Errorf("Should fail to parse secret")
			}
			continue
		} else if err != nil {
			t.Errorf("Failed to parse secret: %s", err)
			continue
		}
		if sec.Password() != tc.Password {
			t.Errorf("[%s] Wrong password", tc.Desc)
		}
		if sec.Body() != tc.Body {
			t.Errorf("[%s] Wrong body: %s - %s", tc.Desc, sec.Body(), tc.Body)
		}
		for k, v := range tc.Data {
			rv, err := sec.Value(k)
			if err != nil {
				t.Fatalf("failed to retrieve value")
			}
			if rv != v {
				t.Errorf("Wrong value for %s", k)
			}
		}
		b, err := sec.Bytes()
		if err != nil {
			t.Fatalf("failed to marshal secret: %s", err)
		}
		if tc.Out != nil {
			if string(b) != string(tc.Out) {
				t.Errorf("wrong bytes: '%s'", string(b))
			}
		}
	}
}
