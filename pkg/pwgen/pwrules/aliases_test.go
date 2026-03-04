package pwrules

import (
	"testing"

	"github.com/gopasspw/gopass/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadCustomRules(t *testing.T) {
	t.Parallel()

	cfg := config.NewInMemory()
	aliases := map[string]string{
		"real.com": "alias.com",
		"real.de":  "copy.de",
	}

	for k, v := range aliases {
		require.NoError(t, cfg.Set("", "domain-alias."+k+".insteadof", v))
	}

	ctx := t.Context()
	ctx = cfg.WithConfig(ctx)

	a := LookupAliases(ctx, "alias.com")
	assert.Equal(t, []string{"real.com"}, a)

	a = LookupAliases(ctx, "copy.de")
	assert.Equal(t, []string{"real.de"}, a)

	assert.Greater(t, len(AllAliases(ctx)), 256)
}
