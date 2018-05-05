package config

import (
	"context"
	"testing"

	"github.com/justwatchcom/gopass/pkg/ctxutil"

	"github.com/stretchr/testify/assert"
)

func TestContext(t *testing.T) {
	ctx := context.Background()

	sc := StoreConfig{
		AskForMore: true,
		NoConfirm:  false,
	}

	// should return the default value from the store config
	assert.Equal(t, false, ctxutil.IsNoConfirm(sc.WithContext(ctx)))

	// after overwriting the noconfirm value in the context,
	// it should not be overwritten by the store config value
	ctx = ctxutil.WithNoConfirm(ctx, true)
	assert.Equal(t, true, ctxutil.IsNoConfirm(sc.WithContext(ctx)))
}
