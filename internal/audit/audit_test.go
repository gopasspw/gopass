package audit

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gopasspw/gopass/internal/backend"
	"github.com/gopasspw/gopass/internal/config"
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
	ctx := t.Context()
	s := &mockSecretGetter{}
	a := New(ctx, s)

	assert.NotNil(t, a)
	assert.Equal(t, s, a.s)
	assert.NotNil(t, a.r)
	assert.NotNil(t, a.v)
}

func TestBatch(t *testing.T) {
	ctx := t.Context()
	s := &mockSecretGetter{}
	a := New(ctx, s)

	secrets := []string{"secret1", "secret2"}
	report, err := a.Batch(ctx, secrets)

	require.NoError(t, err)
	assert.NotNil(t, report)
	assert.Len(t, secrets, len(report.Secrets))
}

func TestAuditSecret(t *testing.T) {
	ctx := t.Context()
	s := &mockSecretGetter{}
	a := New(ctx, s)

	secret := "secret1"
	a.auditSecret(ctx, secret)

	assert.Contains(t, a.r.secrets, secret)
}

type mockLinkedSecretGetter struct {
	mockSecretGetter

	links map[string]string
}

func (m *mockLinkedSecretGetter) LinkTarget(ctx context.Context, name string) (string, bool, error) {
	target, found := m.links[name]

	return target, found, nil
}

func TestBatchSkipsLinks(t *testing.T) {
	ctx := t.Context()
	s := &mockLinkedSecretGetter{
		links: map[string]string{"secret2": "secret1"},
	}
	a := New(ctx, s)

	report, err := a.Batch(ctx, []string{"secret1", "secret2", "secret3"})
	require.NoError(t, err)

	assert.Contains(t, report.Secrets, "secret1")
	assert.NotContains(t, report.Secrets, "secret2")
	assert.Equal(t, "Duplicates detected. Shared with: {secret3}", report.Secrets["secret1"].Findings["duplicates"].Message)
}

func TestCheckHIBP(t *testing.T) {
	ctx := t.Context()
	s := &mockSecretGetter{}
	a := New(ctx, s)

	err := a.checkHIBP(ctx)

	require.NoError(t, err)
}

type concurrencyTrackingSecretGetter struct {
	mockSecretGetter

	concurrency int
	inFlight    atomic.Int32
	maxInFlight atomic.Int32
}

func (m *concurrencyTrackingSecretGetter) Get(ctx context.Context, name string) (gopass.Secret, error) {
	cur := m.inFlight.Add(1)
	for {
		prev := m.maxInFlight.Load()
		if cur <= prev || m.maxInFlight.CompareAndSwap(prev, cur) {
			break
		}
	}
	time.Sleep(5 * time.Millisecond)
	m.inFlight.Add(-1)

	return m.mockSecretGetter.Get(ctx, name)
}

func (m *concurrencyTrackingSecretGetter) Concurrency() int {
	return m.concurrency
}

func TestBatchRespectsConcurrency(t *testing.T) {
	t.Run("store concurrency 1", func(t *testing.T) {
		ctx := config.NewContextInMemory()
		s := &concurrencyTrackingSecretGetter{concurrency: 1}
		a := New(ctx, s)

		report, err := a.Batch(ctx, []string{"s1", "s2", "s3", "s4"})
		require.NoError(t, err)
		assert.Len(t, report.Secrets, 4)
		assert.Equal(t, int32(1), s.maxInFlight.Load())
	})

	t.Run("audit.concurrency caps store concurrency", func(t *testing.T) {
		cfg := config.NewInMemory()
		require.NoError(t, cfg.Set("", "audit.concurrency", "1"))
		ctx := cfg.WithConfig(t.Context())

		s := &concurrencyTrackingSecretGetter{concurrency: 4}
		a := New(ctx, s)

		report, err := a.Batch(ctx, []string{"s1", "s2", "s3", "s4"})
		require.NoError(t, err)
		assert.Len(t, report.Secrets, 4)
		assert.Equal(t, int32(1), s.maxInFlight.Load())
	})

	t.Run("zero concurrency clamped to 1", func(t *testing.T) {
		ctx := config.NewContextInMemory()
		s := &concurrencyTrackingSecretGetter{concurrency: 0}
		a := New(ctx, s)

		report, err := a.Batch(ctx, []string{"s1", "s2"})
		require.NoError(t, err)
		assert.Len(t, report.Secrets, 2)
		assert.Equal(t, int32(1), s.maxInFlight.Load())
	})
}
