package config

import (
	"context"
	"testing"

	"github.com/justwatchcom/gopass/utils/ctxutil"
)

func TestContext(t *testing.T) {
	ctx := context.Background()

	sc := StoreConfig{
		AskForMore: true,
		NoConfirm:  false,
	}

	// should return the default value from the store config
	if ctxutil.IsNoConfirm(sc.WithContext(ctx)) {
		t.Errorf("NoConfirm should be false")
	}

	// after overwriting the noconfirm value in the context,
	// it should not be overwritten by the store config value
	ctx = ctxutil.WithNoConfirm(ctx, true)

	if !ctxutil.IsNoConfirm(sc.WithContext(ctx)) {
		t.Errorf("NoConfirm should be true")
	}
}
