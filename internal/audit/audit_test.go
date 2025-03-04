package audit

import (
	"context"
	"testing"
	"time"

	"github.com/gopasspw/gopass/internal/backend"
	"github.com/gopasspw/gopass/pkg/gopass"
	"github.com/gopasspw/gopass/pkg/gopass/secrets"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockSecretGetter struct{}

func (m *mockSecretGetter) Get(ctx context.Context, name string) (gopass.Secret, error) {
	sec := secrets.New()
	sec.SetPassword("password")

	return sec, nil
}

func (m *mockSecretGetter) ListRevisions(ctx context.Context, name string) ([]backend.Revision, error) {
	return []backend.Revision{
		{Date: time.Now().Add(-time.Hour * 24 * 365)},
	}, nil
}

func (m *mockSecretGetter) Concurrency() int {
	return 1
}

func TestNewAuditor(t *testing.T) {
	ctx := context.Background()
	s := &mockSecretGetter{}
	a := New(ctx, s)

	assert.NotNil(t, a)
	assert.Equal(t, s, a.s)
	assert.NotNil(t, a.r)
	assert.NotNil(t, a.v)
}

func TestBatch(t *testing.T) {
	ctx := context.Background()
	s := &mockSecretGetter{}
	a := New(ctx, s)

	secrets := []string{"secret1", "secret2"}
	report, err := a.Batch(ctx, secrets)

	require.NoError(t, err)
	assert.NotNil(t, report)
	assert.Equal(t, len(secrets), len(report.Secrets))
}

func TestAuditSecret(t *testing.T) {
	ctx := context.Background()
	s := &mockSecretGetter{}
	a := New(ctx, s)

	secret := "secret1"
	a.auditSecret(ctx, secret)

	assert.Contains(t, a.r.secrets, secret)
}

func TestCheckHIBP(t *testing.T) {
	ctx := context.Background()
	s := &mockSecretGetter{}
	a := New(ctx, s)

	err := a.checkHIBP(ctx)

	require.NoError(t, err)
}
