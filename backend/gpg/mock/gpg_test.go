package mock

import (
	"context"
	"testing"
)

func TestMock(t *testing.T) {
	ctx := context.Background()

	m := New()
	kl, err := m.ListPrivateKeys(ctx)
	if err != nil {
		t.Fatalf("Error: %s", err)
	}
	if kl[0].ID() != "0xDEADBEEF" {
		t.Errorf("Wrong Key: %s", kl[0])
	}
}
