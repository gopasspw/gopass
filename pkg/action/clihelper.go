package action

import (
	"context"

	"github.com/justwatchcom/gopass/pkg/cui"
)

// ConfirmRecipients asks the user to confirm a given set of recipients
func (s *Action) ConfirmRecipients(ctx context.Context, name string, recipients []string) ([]string, error) {
	return cui.ConfirmRecipients(ctx, s.Store.Crypto(ctx, name), name, recipients)
}
