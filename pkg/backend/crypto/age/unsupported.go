package age

import (
	"context"
	"fmt"
)

// EmailFromKey is TODO
func (a *Age) EmailFromKey(context.Context, string) string {
	return ""
}

// NameFromKey is TODO
func (a *Age) NameFromKey(context.Context, string) string {
	return ""
}

// FormatKey is TODO
func (a *Age) FormatKey(ctx context.Context, id string) string {
	return id
}

// Fingerprint is TODO
func (a *Age) Fingerprint(ctx context.Context, id string) string {
	return id
}

// ImportPublicKey is TODO
func (a *Age) ImportPublicKey(ctx context.Context, buf []byte) error {
	return nil
}

// Sign is TODO
func (a *Age) Sign(ctx context.Context, in string, sigf string) error {
	return fmt.Errorf("not implemented")
}

// ListPublicKeyIDs is TODO
func (a *Age) ListPublicKeyIDs(context.Context) ([]string, error) {
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
