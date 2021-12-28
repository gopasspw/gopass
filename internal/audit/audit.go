// Package audit contains the password-strength auditing implementation. It reads all decrypted
// passwords and applies different heuristics and external password strength checks to determine
// the quality of the password (i.e. the first line of the secret - only!).
package audit

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/fatih/color"
	"github.com/gopasspw/gopass/internal/backend"
	"github.com/gopasspw/gopass/internal/notify"
	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/pkg/debug"
	"github.com/gopasspw/gopass/pkg/gopass"
	"github.com/gopasspw/gopass/pkg/termio"
	"github.com/muesli/crunchy"
	"github.com/nbutton23/zxcvbn-go"
)

// auditedSecret with its name, content a warning message and a pipeline error.
type auditedSecret struct {
	name string

	// the secret's content as a string. Needed for checking for duplicates.
	content string

	// message to the user about some flaw in the secret
	messages []string

	// real error that something in the pipeline went wrong
	err error
}

type secretGetter interface {
	Get(context.Context, string) (gopass.Secret, error)
	ListRevisions(context.Context, string) ([]backend.Revision, error)
}

type validator func(string, gopass.Secret) error

// Batch runs a password strength audit on multiple secrets
func Batch(ctx context.Context, secrets []string, secStore secretGetter) error {
	out.Printf(ctx, "Checking %d secrets. This may take some time ...\n", len(secrets))

	// Secrets that still need auditing.
	pending := make(chan string, 100)

	// Secrets that have been audited.
	checked := make(chan auditedSecret, 100)

	// Spawn workers that run the auditing of all secrets concurrently.
	cv := crunchy.NewValidator()
	validators := []validator{
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
	}

	// It would be nice to parallelize this operation and limit the maxJobs to
	// runtime.NumCPU(), but sadly this causes various problems with multiple
	// gnupg jobs running in parallel. See the entire discussion here:
	//
	// https://github.com/gopasspw/gopass/pull/245

	maxJobs := 1 // do not change
	done := make(chan struct{}, maxJobs)
	for jobs := 0; jobs < maxJobs; jobs++ {
		go audit(ctx, secStore, validators, pending, checked, done)
	}

	go func() {
		for _, secret := range secrets {
			pending <- secret
		}
		close(pending)
	}()
	go func() {
		for i := 0; i < maxJobs; i++ {
			<-done
		}
		close(checked)
	}()

	duplicates := make(map[string][]string)
	messages := make(map[string][]string)
	errors := make(map[string][]string)

	bar := termio.NewProgressBar(int64(len(secrets)))
	bar.Hidden = ctxutil.IsHidden(ctx)

	i := 0
	for secret := range checked {
		if secret.err != nil {
			en := secret.err.Error()
			errors[en] = append(errors[en], secret.name)
		} else if secret.content != "" {
			duplicates[secret.content] = append(duplicates[secret.content], secret.name)
		}
		for _, m := range secret.messages {
			messages[m] = append(messages[m], secret.name)
		}

		bar.Inc()
		i++
		if i == len(secrets) {
			break
		}
	}
	bar.Done()

	return auditPrintResults(ctx, duplicates, messages, errors)
}

func audit(ctx context.Context, secStore secretGetter, validators []validator, secrets <-chan string, checked chan<- auditedSecret, done chan struct{}) {
	for secret := range secrets {
		as := auditedSecret{
			name: secret,
		}
		// check for context cancelation
		select {
		case <-ctx.Done():
			as.err = errors.New("user aborted")
			checked <- as
			continue
		default:
		}

		debug.Log("Checking %s", secret)
		sec, err := secStore.Get(ctx, secret)
		if err != nil {
			debug.Log("Failed to check %s: %s", secret, err)
			as.err = err
			if sec != nil {
				as.content = sec.Password()
			}
			// failed to properly retrieve the secret
			checked <- as
			continue
		}

		as.content = sec.Password()

		// do not check empty secrets
		if as.content == "" {
			checked <- as
			continue
		}

		// handle password validation errors
		if errs := allValid(validators, secret, sec); len(errs) > 0 {
			for _, e := range errs {
				as.messages = append(as.messages, e.Error())
			}
			checked <- as
			continue
		}

		// handle old passwords
		revs, err := secStore.ListRevisions(ctx, secret)
		if err != nil {
			as.messages = append(as.messages, err.Error())
		} else {
			if len(revs) > 0 && time.Since(revs[0].Date) > 90*24*time.Hour {
				as.messages = append(as.messages, "Password too old (90d)")
			}
		}

		// record every password for possible duplicates
		checked <- as
	}
	done <- struct{}{}
}

func allValid(vs []validator, name string, sec gopass.Secret) []error {
	errs := make([]error, 0, len(vs))
	for _, v := range vs {
		if err := v(name, sec); err != nil {
			errs = append(errs, err)
		}
	}
	return errs
}

func printAuditResults(m map[string][]string, format string, color func(format string, a ...any) string) bool {
	b := false

	for msg, secrets := range m {
		b = true
		fmt.Fprint(out.Stdout, color(format, msg))
		for _, secret := range secrets {
			fmt.Fprint(out.Stdout, color("\t- %s\n", secret))
		}
	}

	return b
}

// Single runs a password strength audit on a single password
func Single(ctx context.Context, password string) {
	validator := crunchy.NewValidator()
	if err := validator.Check(password); err != nil {
		out.Printf(ctx, fmt.Sprintf("Warning: %s", err))
	}
}

func auditPrintResults(ctx context.Context, duplicates, messages, errors map[string][]string) error {
	foundDuplicates := false
	for _, secrets := range duplicates {
		if len(secrets) > 1 {
			foundDuplicates = true

			out.Printf(ctx, "Detected a shared secret for:")
			for _, secret := range secrets {
				out.Printf(ctx, "\t- %s", secret)
			}
		}
	}
	if !foundDuplicates {
		out.Printf(ctx, "No shared secrets found.")
	}

	foundWeakPasswords := printAuditResults(messages, "%s:\n", color.CyanString)
	if !foundWeakPasswords {
		out.Printf(ctx, "No weak secrets detected.")
	}
	foundErrors := printAuditResults(errors, "%s:\n", color.RedString)

	if foundWeakPasswords || foundDuplicates || foundErrors {
		_ = notify.Notify(ctx, "gopass - audit", "Finished. Found weak passwords and/or duplicates")
		return fmt.Errorf("found weak passwords or duplicates")
	}

	_ = notify.Notify(ctx, "gopass - audit", "Finished. No weak passwords or duplicates found!")
	return nil
}
