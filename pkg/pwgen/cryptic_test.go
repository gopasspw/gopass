package pwgen

import (
	"fmt"
	"sort"
	"testing"

	"github.com/gopasspw/gopass/pkg/pwgen/pwrules"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCrypticForDomain(t *testing.T) {
	rules := pwrules.AllRules()
	keys := make([]string, 0, len(rules))
	for k := range rules {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, domain := range keys {
		t.Run(domain, func(t *testing.T) {
			for _, length := range []int{1, 4, 8, 200} {
				tcName := fmt.Sprintf("%s - %d", domain, length)
				c := NewCrypticForDomain(length, domain)
				require.NotNil(t, c, tcName)
				pw := c.Password()
				assert.NotEqual(t, "", pw, tcName)
				t.Logf("%s -> %s (%d)", tcName, pw, len(pw))
			}
		})
	}
}

func TestUniqueChars(t *testing.T) {
	for in, out := range map[string]string{
		"foobar": "abfor",
		"abced":  "abcde",
	} {
		assert.Equal(t, out, uniqueChars(in))
	}
}
