// Package audit contains the password-strength auditing implementation. It reads all decrypted
// passwords and applies different heuristics and external password strength checks to determine
// the quality of the password (i.e. the first line of the secret - only!).
package audit

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/gopasspw/gopass/internal/backend"
	"github.com/gopasspw/gopass/internal/config"
	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/pkg/debug"
	"github.com/gopasspw/gopass/pkg/gopass"
	"github.com/gopasspw/gopass/pkg/termio"
	"github.com/muesli/crunchy"
	"github.com/nbutton23/zxcvbn-go"
)

type secretGetter interface {
	Get(context.Context, string) (gopass.Secret, error)
	ListRevisions(context.Context, string) ([]backend.Revision, error)
	Concurrency() int
}

type validator func(string, gopass.Secret) error

// DefaultExpiration is the default expiration time for secrets.
var DefaultExpiration = time.Hour * 24 * 365

type Auditor struct {
	s      secretGetter
	r      *Report
	expiry time.Duration
	pcb    func()
	v      []validator
}

func New(s secretGetter) *Auditor {
	a := &Auditor{
		s:   s,
		r:   newReport(),
		pcb: func() {},
	}

	cv := crunchy.NewValidator()
	a.v = []validator{
		func(_ string, sec gopass.Secret) error {
			return cv.Check(sec.Password())
		},
		func(name string, sec gopass.Secret) error {
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
		func(name string, sec gopass.Secret) error {
			if name == sec.Password() {
				return fmt.Errorf("password equals name")
			}

			return nil
		},
		// TODO add HIBP validator
	}

	return a
}

// Batch runs a password strength audit on multiple secrets. Expiration is in days.
func (a *Auditor) Batch(ctx context.Context, secrets []string) error {
	out.Printf(ctx, "Checking %d secrets. This may take some time ...\n", len(secrets))

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

	return nil
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
	}
	done <- struct{}{}
}

func (a *Auditor) auditSecret(ctx context.Context, secret string) {
	debug.Log("Auditing %q", secret)

	// handle old passwords
	revs, err := a.s.ListRevisions(ctx, secret)
	if err != nil {
		a.r.AddError(secret, err)
	}
	if len(revs) > 0 {
		a.r.SetAge(secret, time.Since(revs[0].Date))
	}

	sec, err := a.s.Get(ctx, secret)
	if err != nil {
		debug.Log("Failed to check %s: %s", secret, err)

		a.r.AddError(secret, err)
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
	a.r.AddPassword(secret, sec.Password())

	// pass the secret to all validators.
	var wg sync.WaitGroup
	for _, v := range a.v {
		wg.Add(1)
		go func() {
			defer wg.Done()

			if err := v(secret, sec); err != nil {
				a.r.AddWarning(secret, err.Error())
			}
		}()
	}
	wg.Wait()
}
