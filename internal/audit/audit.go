// Package audit contains the password-strength auditing implementation. It reads all decrypted
// passwords and applies different heuristics and external password strength checks to determine
// the quality of the password (i.e. the first line of the secret - only!).
package audit

import (
	"context"
	"fmt"
	"path"
	"sync"
	"time"

	"github.com/gopasspw/gopass-hibp/pkg/hibp/api"
	"github.com/gopasspw/gopass-hibp/pkg/hibp/dump"
	"github.com/gopasspw/gopass/internal/backend"
	"github.com/gopasspw/gopass/internal/config"
	"github.com/gopasspw/gopass/internal/hashsum"
	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/pkg/debug"
	"github.com/gopasspw/gopass/pkg/fsutil"
	"github.com/gopasspw/gopass/pkg/gopass"
	"github.com/gopasspw/gopass/pkg/termio"
	"github.com/muesli/crunchy"
	"github.com/nbutton23/zxcvbn-go"
	"golang.org/x/exp/maps"
)

type secretGetter interface {
	Get(context.Context, string) (gopass.Secret, error)
	ListRevisions(context.Context, string) ([]backend.Revision, error)
	Concurrency() int
	IsSymlink(string) bool
}

type validator struct {
	Name        string
	Description string
	Validate    func(string, gopass.Secret) error
}

// DefaultExpiration is the default expiration time for secrets.
var DefaultExpiration = time.Hour * 24 * 365

type Auditor struct {
	s      secretGetter
	r      *ReportBuilder
	expiry time.Duration
	pcb    func()
	v      []validator
}

func New(ctx context.Context, s secretGetter) *Auditor {
	a := &Auditor{
		s:   s,
		r:   newReport(),
		pcb: func() {},
	}

	cv := crunchy.NewValidator()
	a.v = []validator{
		{
			Name:        "crunchy",
			Description: "github.com/muesli/crunchy",
			Validate: func(_ string, sec gopass.Secret) error {
				return cv.Check(sec.Password())
			},
		},
		{
			Name:        "zxcvbn",
			Description: "github.com/nbutton23/zxcvbn-go",
			Validate: func(name string, sec gopass.Secret) error {
				ui := make([]string, 0, len(sec.Keys())+1)
				for _, k := range sec.Keys() {
					pw, found := sec.Get(k)
					if !found {
						continue
					}
					ui = append(ui, pw)
				}
				ui = append(ui, name)
				match := zxcvbn.PasswordStrength(sec.Password(), ui)
				if match.Score < 3 {
					return fmt.Errorf("weak password (%d / 4)", match.Score)
				}

				return nil
			},
		},
		{
			Name:        "equals-name",
			Description: "Checks for passwords the match the secret name",
			Validate: func(name string, sec gopass.Secret) error {
				if name == sec.Password() || path.Base(name) == sec.Password() {
					return fmt.Errorf("password equals name")
				}

				return nil
			},
		},
	}

	if config.Bool(ctx, "audit.hibp-use-api") {
		a.v = append(a.v, validator{
			Name:        "hibp",
			Description: "Checks passwords against the HIBPv2 API. See https://haveibeenpwned.com/",
			Validate: func(_ string, sec gopass.Secret) error {
				if sec.Password() == "" {
					return nil
				}

				numFound, err := api.Lookup(hashsum.SHA1Hex(sec.Password()))
				if err != nil {
					return fmt.Errorf("can't check HIBPv2 API: %w", err)
				}

				if numFound > 0 {
					return fmt.Errorf("password contained in at least %d public data breaches (HIBP API)", numFound)
				}

				return nil
			},
		})
	}

	return a
}

// Batch runs a password strength audit on multiple secrets. Expiration is in days.
func (a *Auditor) Batch(ctx context.Context, secrets []string) (*Report, error) {
	out.Printf(ctx, "Checking %d secrets. This may take some time ...\n", len(secrets))

	a.r = newReport()
	pending := make(chan string, 1024)

	// It would be nice to parallelize this operation and limit the maxJobs to
	// runtime.NumCPU(), but sadly this causes various problems with multiple
	// gnupg jobs running in parallel. See the entire discussion here:
	//
	// https://github.com/gopasspw/gopass/pull/245
	//
	maxJobs := a.s.Concurrency()
	if max := config.Int(ctx, "audit.concurrency"); max > 0 {
		if maxJobs > max {
			maxJobs = max
		}
	}

	// Spawn workers that run the auditing of all secrets concurrently.
	debug.Log("launching %d audit workers", maxJobs)

	done := make(chan struct{}, maxJobs)
	for jobs := 0; jobs < maxJobs; jobs++ {
		go a.audit(ctx, pending, done)
	}

	go func() {
		for _, secret := range secrets {
			pending <- secret
		}
		close(pending)
	}()

	bar := termio.NewProgressBar(int64(len(secrets)))
	bar.Hidden = ctxutil.IsHidden(ctx)
	a.pcb = func() {
		bar.Inc()
	}

	for i := 0; i < maxJobs; i++ {
		<-done
	}
	bar.Done()

	if err := a.checkHIBP(ctx); err != nil {
		return nil, err
	}

	return a.r.Finalize(), nil
}

func (a *Auditor) audit(ctx context.Context, secrets <-chan string, done chan struct{}) {
	for secret := range secrets {
		// check for context cancelation.
		select {
		case <-ctx.Done():
			continue
		default:
		}

		a.auditSecret(ctx, secret)
		a.pcb()
	}
	done <- struct{}{}
}

func (a *Auditor) auditSecret(ctx context.Context, secret string) {
	debug.Log("Auditing %q", secret)

	// handle old passwords
	revs, err := a.s.ListRevisions(ctx, secret)
	if err != nil {
		a.r.AddFinding(secret, "error-revisions", err.Error(), "error")
	}
	if len(revs) > 0 {
		a.r.SetAge(secret, time.Since(revs[0].Date))
	}

	sec, err := a.s.Get(ctx, secret)
	if err != nil {
		debug.Log("Failed to check %s: %s", secret, err)

		a.r.AddFinding(secret, "error-read", err.Error(), "error")
		if sec != nil {
			a.r.AddPassword(secret, sec.Password())
		}

		return
	}

	// do not check empty secrets.
	if sec.Password() == "" {
		return
	}

	// add the password for the duplicate check
	if !a.s.IsSymlink(secret) {
		a.r.AddPassword(secret, sec.Password())
	}

	// pass the secret to all validators.
	var wg sync.WaitGroup
	for _, v := range a.v {
		v := v

		wg.Add(1)
		go func() {
			defer wg.Done()

			if err := v.Validate(secret, sec); err != nil {
				a.r.AddFinding(secret, v.Name, err.Error(), "warning")

				return
			}

			a.r.AddFinding(secret, v.Name, "ok", "none")
		}()
	}
	wg.Wait()
}

func (a *Auditor) checkHIBP(ctx context.Context) error {
	if config.Bool(ctx, "audit.hibp-use-api") {
		// no need to check the dumps if we already checked the API
		return nil
	}

	// if the user has set up the path to an HIBP dump we can continue.
	fn := config.String(ctx, "audit.hibp-dump-file")
	if fn == "" || !fsutil.IsFile(fn) {
		debug.Log("audit.hibp-dump-file not pointing to a valid dump file")

		return nil
	}

	// if creating the scanner fails the dump file is most likely invalid.
	scanner, err := dump.New(fn)
	if err != nil {
		return err
	}

	out.Notice(ctx, "Starting HIBP check (slow) ...")

	// look up all known sha1sums. The LookupBatch method will sort the
	// input so we don't need to.
	matches := scanner.LookupBatch(ctx, maps.Keys(a.r.sha1sums))
	for _, m := range matches {
		// map any match back to the secret(s).
		secs, found := a.r.sha1sums[m]
		if !found {
			// should not happen
			continue
		}

		// add a breach warning to each of these secrets.
		for _, sec := range secs.Elements() {
			a.r.AddFinding(sec, "hibp", "Found in at least one public data breach (HIBP Dump)", "warning")
		}
	}

	for name, sr := range a.r.secrets {
		if sr.Findings == nil {
			sr.Findings = make(map[string]Finding, 1)
		}
		if _, found := sr.Findings["hibp"]; !found {
			sr.Findings["hibp"] = Finding{
				Severity: "none",
				Message:  "ok",
			}
			a.r.secrets[name] = sr
		}
	}

	return nil
}
