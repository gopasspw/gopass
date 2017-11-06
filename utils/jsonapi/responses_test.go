package jsonapi

import (
	"testing"

	"github.com/justwatchcom/gopass/store/secret"
)

func TestGetUsername(t *testing.T) {
	a := &API{}
	for _, tc := range []struct {
		Name string
		Sec  *secret.Secret
		Out  string
	}{
		{
			Name: "some/fixed/yamlother",
			Sec:  secret.New("thesecret", "---\nother: meh"),
			Out:  "yamlother",
		},
		{
			Name: "some/key/withaname",
			Sec:  secret.New("thesecret", "---\nlogin: foo"),
			Out:  "foo",
		},
	} {
		got := a.getUsername(tc.Name, tc.Sec)
		if got != tc.Out {
			t.Errorf("Wrong username: %s != %s", got, tc.Out)
		}
	}
}
