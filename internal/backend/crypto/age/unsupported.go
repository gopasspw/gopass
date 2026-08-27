package age

import (
	"context"
	"fmt"
)

// FormatKey returns the key id.
func (a *Age) FormatKey(ctx context.Context, id, tpl string) string {
	return id
}

// FormatKeys returns the IDs.
func (a *Age) FormatKeys(ctx context.Context, ids []string) map[string]string {
	formatted := make(map[string]string, len(ids))
	for _, id := range ids {
		formatted[id] = id
	}

	return formatted
}

// Fingerprint returns the id.
func (a *Age) Fingerprint(ctx context.Context, id string) string {
	return id
}

// ListRecipients is not supported for the age backend.
func (a *Age) ListRecipients(context.Context) ([]string, error) {
	return nil, fmt.Errorf("not implemented")
}

// ReadNamesFromKey is not supported for the age backend.
func (a *Age) ReadNamesFromKey(ctx context.Context, buf []byte) ([]string, error) {
	return nil, fmt.Errorf("not implemented")
}

// RecipientIDs is not supported for the age backend.
func (a *Age) RecipientIDs(ctx context.Context, buf []byte) ([]string, error) {
	return nil, fmt.Errorf("reading recipient IDs is not supported by the age backend by design")
}
