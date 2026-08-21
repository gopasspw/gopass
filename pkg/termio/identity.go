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

// DetectName tries to get the name of the user by checking
// (1) context, (2) command line flag, (3) GIT_AUTHOR_NAME,
// (4) git config, (5) DEBFULLNAME, and (6) USER.
func DetectName(ctx context.Context, cmd *cli.Command) string {
	if n := ctxutil.GetUsername(ctx); n != "" {
		return n
	}
	if cmd != nil && cmd.String("name") != "" {
		return cmd.String("name")
	}

	for _, envVar := range NameVars {
		if n := os.Getenv(envVar); n != "" {
			return n
		}

		// Check git config right after GIT_AUTHOR_NAME env var
		if envVar == "GIT_AUTHOR_NAME" {
			if n := gitconfig.New().LoadAll(GetWorkdir(ctx)).Get("user.name"); n != "" {
				return n
			}
		}
	}

	return ""
}

// DetectEmail tries to get the email of the user by checking
// (1) context, (2) command line flag, (3) GIT_AUTHOR_EMAIL,
// (4) git config, (5) DEBEMAIL, and (6) EMAIL.
func DetectEmail(ctx context.Context, cmd *cli.Command) string {
	if e := ctxutil.GetEmail(ctx); e != "" {
		return e
	}
	if cmd != nil && cmd.String("email") != "" {
		return cmd.String("email")
	}

	for _, envVar := range EmailVars {
		if e := os.Getenv(envVar); e != "" {
			return e
		}

		// Check git config right after GIT_AUTHOR_EMAIL env var
		if envVar == "GIT_AUTHOR_EMAIL" {
			if e := gitconfig.New().LoadAll(GetWorkdir(ctx)).Get("user.email"); e != "" {
				return e
			}
		}
	}

	return ""
}
