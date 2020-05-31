package termio

import (
	"context"
	"os"

	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/urfave/cli/v2"
)

// DetectName tries to guess the name of the logged in user
func DetectName(ctx context.Context, c *cli.Context) string {
	cand := make([]string, 0, 5)
	cand = append(cand, ctxutil.GetUsername(ctx))
	if c != nil {
		cand = append(cand, c.String("name"))
	}
	cand = append(cand,
		os.Getenv("GIT_AUTHOR_NAME"),
		os.Getenv("DEBFULLNAME"),
		os.Getenv("USER"),
	)
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
	cand = append(cand,
		os.Getenv("GIT_AUTHOR_EMAIL"),
		os.Getenv("DEBEMAIL"),
		os.Getenv("EMAIL"),
	)
	for _, e := range cand {
		if e != "" {
			return e
		}
	}
	return ""
}
