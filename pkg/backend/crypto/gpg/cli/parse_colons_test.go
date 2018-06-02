package cli

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseColonIdentity(t *testing.T) {
	for _, tc := range []struct {
		in      string
		name    string
		comment string
		email   string
	}{
		{
			in:      "uid:-::::1460666077::780A2FDD0570B3E52E5B1E24EBB406B68526CAFD::ThisIsNotAnAlias:",
			name:    "ThisIsNotAnAlias",
			comment: "",
			email:   "",
		},
		{
			in:      "uid:::::1441103821::AEFC3F5B6CAD79A946D7F0FF83BB8B7E10B578CA::John Doe <john.doe@example.com>:",
			name:    "John Doe",
			comment: "",
			email:   "john.doe@example.com",
		},
		{
			in:      "uid:::::1441103821::AEFC3F5B6CAD79A946D7F0FF83BB8B7E10B578CA::John Doe (user) <john.doe@example.com>:",
			name:    "John Doe",
			comment: "user",
			email:   "john.doe@example.com",
		},
		{
			in:      "uid:::::1441103821::AEFC3F5B6CAD79A946D7F0FF83BB8B7E10B578CA::John Doe (user):",
			name:    "John Doe",
			comment: "user",
			email:   "",
		},
	} {
		gi := parseColonIdentity(strings.Split(tc.in, ":"))
		assert.Equal(t, tc.name, gi.Name)
		assert.Equal(t, tc.comment, gi.Comment)
		assert.Equal(t, tc.email, gi.Email)
	}
}
