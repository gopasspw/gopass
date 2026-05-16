package termio

import (
	"context"
	"os"

	"github.com/gopasspw/gitconfig"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/urfave/cli/v3"
)

var (
	// NameVars are the env vars checked for a valid name.
	NameVars = []string{
		"GIT_AUTHOR_NAME",
		"DEBFULLNAME",
		"USER",
	}
	// EmailVars are the env vars checked for a valid email.
	EmailVars = []string{
		"GIT_AUTHOR_EMAIL",
		"DEBEMAIL",
		"EMAIL",
	}
)

// DetectName tries to guess the name of the logged in user.
// It checks the context, the command line flags, environment variables,
// and the git config.
func DetectName(ctx context.Context, cmd *cli.Command) string {
	cand := make([]string, 0, 10)
	cand = append(cand, ctxutil.GetUsername(ctx))

	if cmd != nil {
		cand = append(cand, cmd.String("name"))
	}

	for _, k := range NameVars {
		cand = append(cand, os.Getenv(k))
	}

	cfg := gitconfig.New().LoadAll(GetWorkdir(ctx))
	cand = append(cand, cfg.Get("user.name"))

	for _, e := range cand {
		if e != "" {
			return e
		}
	}

	return ""
}

// DetectEmail tries to guess the email of the logged in user.
// It checks the context, the command line flags, environment variables,
// and the git config.
func DetectEmail(ctx context.Context, cmd *cli.Command) string {
	cand := make([]string, 0, 10)
	cand = append(cand, ctxutil.GetEmail(ctx))

	if cmd != nil {
		cand = append(cand, cmd.String("email"))
	}

	for _, k := range EmailVars {
		cand = append(cand, os.Getenv(k))
	}

	cfg := gitconfig.New().LoadAll(GetWorkdir(ctx))
	cand = append(cand, cfg.Get("user.email"))

	for _, e := range cand {
		if e != "" {
			return e
		}
	}

	return ""
}
