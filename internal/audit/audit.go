// Package audit contains the password-strength auditing implementation
package audit

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"strings"
	"time"

	"github.com/gopasspw/gopass/internal/backend"
	"github.com/gopasspw/gopass/internal/debug"
	"github.com/gopasspw/gopass/internal/notify"
	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/internal/store"

	"github.com/fatih/color"
	"github.com/muesli/crunchy"
	"github.com/muesli/goprogressbar"
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
	Get(context.Context, string) (store.Secret, error)
	ListRevisions(context.Context, string) ([]backend.Revision, error)
}

// Batch runs a password strength audit on multiple secrets
func Batch(ctx context.Context, secrets []string, secStore secretGetter) error {
	out.Print(ctx, "Checking %d secrets. This may take some time ...\n", len(secrets))

	// Secrets that still need auditing.
	pending := make(chan string, 100)

	// Secrets that have been audited.
	checked := make(chan auditedSecret, 100)

	// Spawn workers that run the auditing of all secrets concurrently.
	validator := crunchy.NewValidator()

	// It would be nice to parallelize this operation and limit the maxJobs to
	// runtime.NumCPU(), but sadly this causes various problems with multiple
	// gnupg jobs running parallelly. See the entire discussion here:
	//
	// https://github.com/gopasspw/gopass/pull/245

	maxJobs := 1 // do not change
	done := make(chan struct{}, maxJobs)
	for jobs := 0; jobs < maxJobs; jobs++ {
		go audit(ctx, secStore, validator, pending, checked, done)
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

	bar := &goprogressbar.ProgressBar{
		Total: int64(len(secrets)),
		Width: 120,
	}
	if out.IsHidden(ctx) {
		old := goprogressbar.Stdout
		goprogressbar.Stdout = ioutil.Discard
		defer func() {
			goprogressbar.Stdout = old
		}()
	}

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

		i++
		bar.Current = int64(i)
		if bar.Current > bar.Total {
			bar.Total = bar.Current
		}
		bar.Text = fmt.Sprintf("%d of %d secrets checked", bar.Current, bar.Total)
		bar.LazyPrint()

		if i == len(secrets) {
			break
		}
	}
	fmt.Fprintln(goprogressbar.Stdout) // Print empty line after the progressbar.

	return auditPrintResults(ctx, duplicates, messages, errors)
}

func audit(ctx context.Context, secStore secretGetter, validator *crunchy.Validator, secrets <-chan string, checked chan<- auditedSecret, done chan struct{}) {
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
			as.err = err
			if sec != nil {
				as.content = sec.Password()
			}
			// failed to properly retrieve the secret
			checked <- as
			continue
		}

		as.content = sec.Password()

		// do not check binary secrets
		if as.content == "" || strings.HasSuffix(secret, ".b64") {
			checked <- as
			continue
		}

		// handle password validation errors
		if err := validator.Check(as.content); err != nil {
			as.messages = append(as.messages, err.Error())
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

func printAuditResults(m map[string][]string, format string, color func(format string, a ...interface{}) string) bool {
	b := false

	for msg, secrets := range m {
		b = true
		fmt.Fprint(goprogressbar.Stdout, color(format, msg))
		for _, secret := range secrets {
			fmt.Fprint(goprogressbar.Stdout, color("\t- %s\n", secret))
		}
	}

	return b
}

// Single runs a password strength audit on a single password
func Single(ctx context.Context, password string) {
	validator := crunchy.NewValidator()
	if err := validator.Check(password); err != nil {
		out.Cyan(ctx, fmt.Sprintf("Warning: %s", err))
	}
}

func auditPrintResults(ctx context.Context, duplicates, messages, errors map[string][]string) error {
	foundDuplicates := false
	for _, secrets := range duplicates {
		if len(secrets) > 1 {
			foundDuplicates = true

			out.Cyan(ctx, "Detected a shared secret for:")
			for _, secret := range secrets {
				out.Cyan(ctx, "\t- %s", secret)
			}
		}
	}
	if !foundDuplicates {
		out.Green(ctx, "No shared secrets found.")
	}

	foundWeakPasswords := printAuditResults(messages, "%s:\n", color.CyanString)
	if !foundWeakPasswords {
		out.Green(ctx, "No weak secrets detected.")
	}
	foundErrors := printAuditResults(errors, "%s:\n", color.RedString)

	if foundWeakPasswords || foundDuplicates || foundErrors {
		_ = notify.Notify(ctx, "gopass - audit", "Finished. Found weak passwords and/or duplicates")
		return fmt.Errorf("found weak passwords or duplicates")
	}

	_ = notify.Notify(ctx, "gopass - audit", "Finished. No weak passwords or duplicates found!")
	return nil
}
