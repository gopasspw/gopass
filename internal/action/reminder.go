package action

import (
	"context"
	"os"

	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/pkg/ctxutil"
)

func (s *Action) printReminder(ctx context.Context) {
	if !ctxutil.IsInteractive(ctx) {
		return
	}
	if !ctxutil.IsTerminal(ctx) {
		return
	}
	if sv := os.Getenv("GOPASS_NO_REMINDER"); sv != "" {
		return
	}

	// Note: We only want to print one reminder per day (at most).
	// So we intentionally return after printing one, leaving the others
	// for the following days.
	if s.rem.Overdue("update") {
		out.Notice(ctx, "You haven't checked for updates in a while. Run 'gopass version' or 'gopass update' to check.")
		return
	}

	if s.rem.Overdue("fsck") {
		out.Notice(ctx, "You haven't run 'gopass fsck' in a while.")
		return
	}

	if s.rem.Overdue("audit") {
		out.Notice(ctx, "You haven't run 'gopass audit' in a while.")
		return
	}
}
