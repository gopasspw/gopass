package jsonapi

import (
	"testing"

	"github.com/gopasspw/gopass/pkg/gopass"
)

func TestGetUsername(t *testing.T) {
	a := &API{}
	for _, tc := range []struct {
		Name string
		Sec  gopass.Secret
		Out  string
	}{
		{
			Name: "some/fixed/yamlother",
			Sec:  newSec(t, "thesecret\n---\nother: meh"),
			Out:  "yamlother",
		},
		{
			Name: "some/key/withaname",
			Sec:  newSec(t, "thesecret\n---\nlogin: foo"),
			Out:  "foo",
		},
	} {
		got := a.getUsername(tc.Name, tc.Sec)
		if got != tc.Out {
			t.Errorf("Wrong username: %s != %s", got, tc.Out)
		}
	}
}
