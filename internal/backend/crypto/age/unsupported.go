package age

import (
	"context"
	"fmt"
)

// FormatKey is TODO
func (a *Age) FormatKey(ctx context.Context, id, tpl string) string {
	return id
}

// Fingerprint return the id
func (a *Age) Fingerprint(ctx context.Context, id string) string {
	return id
}

// ImportPublicKey is TODO
func (a *Age) ImportPublicKey(ctx context.Context, buf []byte) error {
	return nil
}

// ListRecipients is TODO
func (a *Age) ListRecipients(context.Context) ([]string, error) {
	return nil, fmt.Errorf("not implemented")
}

// ReadNamesFromKey is TODO
func (a *Age) ReadNamesFromKey(ctx context.Context, buf []byte) ([]string, error) {
	return nil, fmt.Errorf("not implemented")
}

// RecipientIDs is TODO
func (a *Age) RecipientIDs(ctx context.Context, buf []byte) ([]string, error) {
	return nil, fmt.Errorf("not supported by backend")
}
