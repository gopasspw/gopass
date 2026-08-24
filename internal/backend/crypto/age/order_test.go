package age

import (
	"context"
	"testing"

	"filippo.io/age"
	"github.com/gopasspw/gopass/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOrderedIdentities(t *testing.T) {
	t.Parallel()

	id1, err := age.GenerateX25519Identity()
	require.NoError(t, err)
	id2, err := age.GenerateX25519Identity()
	require.NoError(t, err)

	r1 := id1.Recipient().String()
	r2 := id2.Recipient().String()

	ids := map[string]age.Identity{
		r1: id1,
		r2: id2,
	}

	ctx := config.NewContextInMemory()

	// without a preference the order must be stable and deterministic
	got := orderedIdentities(ctx, ids)
	require.Len(t, got, 2)
	assert.Equal(t, orderedIdentities(ctx, ids), got, "order must be deterministic")

	// with a preference the preferred identity must come first
	ctx = config.NewInMemory().WithConfig(context.Background())
	cfg, _ := config.FromContext(ctx)
	require.NoError(t, cfg.SetEnv("age.identities", r2))
	got = orderedIdentities(ctx, ids)
	require.Len(t, got, 2)
	assert.Equal(t, id2, got[0], "preferred identity must be tried first")
	assert.Equal(t, id1, got[1])

	// secret key material must never be matched: a config entry containing
	// the secret identity encoding must not influence the ordering at all,
	// i.e. the result must be identical to the no-preference baseline.
	baseline := orderedIdentities(config.NewContextInMemory(), ids)
	ctx = config.NewInMemory().WithConfig(context.Background())
	cfg, _ = config.FromContext(ctx)
	require.NoError(t, cfg.SetEnv("age.identities", id2.String()))
	got = orderedIdentities(ctx, ids)
	require.Len(t, got, 2)
	assert.Equal(t, baseline, got, "secret identity encodings must not be matched")

	// unknown entries in the preference list are ignored
	ctx = config.NewInMemory().WithConfig(context.Background())
	cfg, _ = config.FromContext(ctx)
	require.NoError(t, cfg.SetEnv("age.identities", "age1doesnotexist,"+r1))
	got = orderedIdentities(ctx, ids)
	require.Len(t, got, 2)
	assert.Equal(t, id1, got[0])

	// comma-separated values are split and deduped
	ctx = config.NewInMemory().WithConfig(context.Background())
	cfg, _ = config.FromContext(ctx)
	require.NoError(t, cfg.SetEnv("age.identities", r2+","+r1+","+r2))
	got = orderedIdentities(ctx, ids)
	require.Len(t, got, 2)
	assert.Equal(t, id2, got[0])
	assert.Equal(t, id1, got[1])
}

func TestPreferredIdentities(t *testing.T) {
	t.Parallel()

	ctx := config.NewContextInMemory()
	assert.Empty(t, preferredIdentities(ctx))

	cfg := config.NewInMemory()
	require.NoError(t, cfg.SetEnv("age.identities", " a , b ,,a "))
	ctx = cfg.WithConfig(context.Background())
	assert.Equal(t, []string{"a", "b"}, preferredIdentities(ctx))
}

func TestRecipientOf(t *testing.T) {
	t.Parallel()

	id, err := age.GenerateX25519Identity()
	require.NoError(t, err)

	// must return the public recipient, never the secret key
	assert.Equal(t, id.Recipient().String(), recipientOf(id))
	assert.NotContains(t, recipientOf(id), "AGE-SECRET-KEY")
}
