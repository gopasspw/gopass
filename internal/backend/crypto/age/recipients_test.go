package age

import (
	"context"
	"crypto/ecdh"
	"crypto/rand"
	"testing"

	"filippo.io/age"
	"filippo.io/age/tag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFindRecipients(t *testing.T) {
	ctx := t.Context()
	a := &Age{
		ghCache: &mockGHCache{},
	}

	tests := []struct {
		name     string
		search   []string
		expected []string
	}{
		{
			name:     "github key",
			search:   []string{"github:username"},
			expected: []string{"ssh-rsa AAAAB3Nza..."},
		},
		{
			name:     "ssh key",
			search:   []string{"ssh-rsa AAAAB3Nza..."},
			expected: []string{"ssh-rsa AAAAB3Nza..."},
		},
		{
			name:     "age key",
			search:   []string{"age1qxy2z..."},
			expected: []string{"age1qxy2z..."},
		},
		{
			name:     "unknown key",
			search:   []string{"unknown:key"},
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recipients, err := a.FindRecipients(ctx, tt.search...)
			require.NoError(t, err)
			assert.ElementsMatch(t, tt.expected, recipients)
		})
	}
}

func TestParseRecipients(t *testing.T) {
	ctx := t.Context()
	a := &Age{
		ghCache: &mockGHCache{},
	}

	hybridID, err := age.GenerateHybridIdentity()
	require.NoError(t, err)

	p256Key, err := ecdh.P256().GenerateKey(rand.Reader)
	require.NoError(t, err)
	uncompressed := p256Key.PublicKey().Bytes()
	compressed := append([]byte{0x02 | (uncompressed[64] & 1)}, uncompressed[1:33]...)
	tagRec, err := tag.NewClassicRecipient(compressed)
	require.NoError(t, err)

	// both age and ssh keys are valid, throw away keys generated for this test case.
	tests := []struct {
		name     string
		input    []string
		expected int
	}{
		{
			name:     "valid age key",
			input:    []string{"age1zf3t7aw2rv39fmcddc469nhtj6lm22kn5kh0gy4fv3a7ds3r29rsr69l89"},
			expected: 1,
		},
		{
			name:     "valid x25519 key with age1pq prefix",
			input:    []string{"age1pqv6rcekrltr2pn59fc2z2lxdjqaaxd3wj676kc24zrr3ezp29mqte5yp2"},
			expected: 1,
		},
		{
			name:     "valid x25519 key with age1tag prefix",
			input:    []string{"age1tagy0nwqamzmsldj9w3swaryquxcwwu3hwc0evuhckv403v77axqcv6rce"},
			expected: 1,
		},
		{
			name:     "valid hybrid pq key",
			input:    []string{hybridID.Recipient().String()},
			expected: 1,
		},
		{
			name:     "valid tag key",
			input:    []string{tagRec.String()},
			expected: 1,
		},
		{
			name:     "valid plugin key",
			input:    []string{"age1yubikey1qt2r3tfk7wvlykudm7ew28dqqm3h8ln9zfsxsq4lcd2w8rh4n4hhz46ur24"},
			expected: 1,
		},
		{
			name:     "valid ssh key",
			input:    []string{"ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQDHnOMnKBLlwDOHWh0EEk8r/VdeLaPe7hOwdS/040YVU04PXG7U0YFoXRe1GP6KtM0SrIOIl2S3QrCc7m1cwZSBBPE3rVprrCShaIG2Wn57nTPv2kb9Qtqlc8nMXBOKITCfLmtuzN39n7E7T0EZGrThocrvcNCsPLdrc8Nd0I+eVidgN215DeWhDB4X0pJmScMRSWOmFgnPEPBpDcHvly9wTT+Iv8V7mvGiVKYBHFBA73lCpLS1+LWa+7GXJkKsLbZtBgOQKj9txmwRMkQCecrBAN3z5skdAQc1XPTc3Nihzw6FnPAe69hmjgVl8YTSdmojxbpaJwLvpkR9/Gv5w9ZH/VYM2lhmhCoXTVTLWDGIbxEG3tjEhB7dfVVEcLRod33X2f1LIzhC5lW+dIwVV9IprJooCAtNnHy06DNpQNE/2YTTjCtUSx+DX+ZLEHaGQ2QXlaARXnUfNgM+ct8VAGRL/UkQnqGDE7NgQ4U6JfsohWfR8QXrEkAvLzctmw2AHc8="},
			expected: 1,
		},
		{
			name:     "unknown key",
			input:    []string{"unknown:key"},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recipients, err := a.parseRecipients(ctx, tt.input)
			require.NoError(t, err)
			assert.Len(t, recipients, tt.expected)
		})
	}
}

type mockGHCache struct{}

func (m *mockGHCache) ListKeys(ctx context.Context, user string) ([]string, error) {
	return []string{"ssh-rsa AAAAB3Nza..."}, nil
}

func (m *mockGHCache) String() string {
	return "mockGHCache"
}
