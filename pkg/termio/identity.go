package termio

import (
	"context"
	"os"
	"strings"

	"github.com/gopasspw/gitconfig"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/urfave/cli/v3"
)

type strFn func(ctx context.Context, cmd *cli.Command) string

var (
	nameVars = []strFn{
		func(ctx context.Context, cmd *cli.Command) string { return ctxutil.GetUsername(ctx) },
		func(ctx context.Context, cmd *cli.Command) string {
			if cmd != nil {
				return cmd.String("name")
			}

			return ""
		},
		func(ctx context.Context, cmd *cli.Command) string { return os.Getenv("GIT_AUTHOR_NAME") },
		func(ctx context.Context, cmd *cli.Command) string {
			cfg := gitconfig.New().LoadAll(GetWorkdir(ctx))

			return cfg.Get("user.name")
		},
		func(ctx context.Context, cmd *cli.Command) string { return os.Getenv("DEBFULLNAME") },
		func(ctx context.Context, cmd *cli.Command) string { return os.Getenv("USER") },
	}
	emailVars = []strFn{
		func(ctx context.Context, cmd *cli.Command) string { return ctxutil.GetEmail(ctx) },
		func(ctx context.Context, cmd *cli.Command) string {
			if cmd != nil {
				return cmd.String("email")
			}

			return ""
		},
		func(ctx context.Context, cmd *cli.Command) string { return os.Getenv("GIT_AUTHOR_EMAIL") },
		func(ctx context.Context, cmd *cli.Command) string {
			cfg := gitconfig.New().LoadAll(GetWorkdir(ctx))

			return cfg.Get("user.email")
		},
		func(ctx context.Context, cmd *cli.Command) string { return os.Getenv("DEBEMAIL") },
		func(ctx context.Context, cmd *cli.Command) string { return os.Getenv("EMAIL") },
	}
)

func detect(ctx context.Context, cmd *cli.Command, vars []strFn) string {
	for _, fn := range vars {
		s := strings.TrimSpace(fn(ctx, cmd))
		if s != "" {
			return s
		}
	}

	return ""
}

// DetectName tries to guess the name of the logged in user.
// It checks the context, the command line flags, environment variables,
// and the git config.
func DetectName(ctx context.Context, cmd *cli.Command) string {
	return detect(ctx, cmd, nameVars)
}

// DetectEmail tries to guess the email of the logged in user.
// It checks the context, the command line flags, environment variables,
// and the git config.
func DetectEmail(ctx context.Context, cmd *cli.Command) string {
	return detect(ctx, cmd, emailVars)
}
