package pwrules

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseRule(t *testing.T) {
	for _, tc := range []struct {
		in  string
		out Rule
	}{
		{
			in: "minlength: 8; maxlength: 20; required: upper; required: lower; required: digit; max-consecutive: 3; allowed: [@#*()+={}/?~;,.-_];",
			out: Rule{
				Minlen: 8,
				Maxlen: 20,
				Required: []string{
					"upper",
					"lower",
					"digit",
				},
				Allowed: []string{
					"[@#*()+={}/?~;,.-_]",
				},
				Maxconsec: 3,
			},
		},
		{
			in: "minlength: 7; maxlength: 16; required: lower, upper; required: digit; required: [`!@#$%^&*()+~{}'\";:<>?]];",
			out: Rule{
				Minlen: 7,
				Maxlen: 16,
				Required: []string{
					"lower",
					"upper",
					"digit",
					"[`!@#$%^&*()+~{}'\";:<>?]]",
				},
			},
		},
		{
			in: "minlength: 8; maxlength: 16;",
			out: Rule{
				Minlen: 8,
				Maxlen: 16,
			},
		},
	} {
		tc := tc
		t.Run(tc.in, func(t *testing.T) {
			assert.Equal(t, tc.out, ParseRule(tc.in))
		})
	}
}
