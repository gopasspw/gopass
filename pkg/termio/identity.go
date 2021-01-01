package termio

import (
	"context"
	"os"

	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/urfave/cli/v2"
)

var (
	// NameVars are the env vars checked for a valid name
	NameVars = []string{
		"GIT_AUTHOR_NAME",
		"DEBFULLNAME",
		"USER",
	}
	// EmailVars are the env vars checked for a valid email
	EmailVars = []string{
		"GIT_AUTHOR_EMAIL",
		"DEBEMAIL",
		"EMAIL",
	}
)

// DetectName tries to guess the name of the logged in user
func DetectName(ctx context.Context, c *cli.Context) string {
	cand := make([]string, 0, 5)
	cand = append(cand, ctxutil.GetUsername(ctx))
	if c != nil {
		cand = append(cand, c.String("name"))
	}
	for _, k := range NameVars {
		cand = append(cand, os.Getenv(k))
	}
	for _, e := range cand {
		if e != "" {
			return e
		}
	}
	return ""
}

// DetectEmail tries to guess the email of the logged in user
func DetectEmail(ctx context.Context, c *cli.Context) string {
	cand := make([]string, 0, 5)
	cand = append(cand, ctxutil.GetEmail(ctx))
	if c != nil {
		cand = append(cand, c.String("email"))
	}
	for _, k := range EmailVars {
		cand = append(cand, os.Getenv(k))
	}
	for _, e := range cand {
		if e != "" {
			return e
		}
	}
	return ""
}
