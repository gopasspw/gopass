// Package audit contains the password-strength auditing implementation
package audit

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"runtime"

	"github.com/fatih/color"
	"github.com/justwatchcom/gopass/pkg/notify"
	"github.com/justwatchcom/gopass/pkg/out"
	"github.com/justwatchcom/gopass/pkg/store"
	"github.com/muesli/crunchy"
	"github.com/muesli/goprogressbar"
)

// auditedSecret with its name, content a warning message and a pipeline error.
type auditedSecret struct {
	name string

	// the secret's content as a string. Needed for checking for duplicates.
	content string

	// message to the user about some flaw in the secret
	message string

	// real error that something in the pipeline went wrong
	err error
}

type secretGetter interface {
	Get(context.Context, string) (store.Secret, error)
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
	maxJobs := runtime.NumCPU()
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
			errors[secret.err.Error()] = append(errors[secret.err.Error()], secret.name)
		} else {
			duplicates[secret.content] = append(duplicates[secret.content], secret.name)
		}
		if secret.message != "" {
			messages[secret.message] = append(messages[secret.message], secret.name)
		}

		i++
		bar.Current = int64(i)
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
		// check for context cancelation
		select {
		case <-ctx.Done():
			checked <- auditedSecret{name: secret, content: "", err: errors.New("user aborted")}
			continue
		default:
		}

		sec, err := secStore.Get(ctx, secret)
		if err != nil {
			pw := ""
			if sec != nil {
				pw = sec.Password()
			}
			checked <- auditedSecret{name: secret, content: pw, err: err}
			continue
		}

		if err := validator.Check(sec.Password()); err != nil {
			checked <- auditedSecret{name: secret, content: sec.Password(), message: err.Error()}
			continue
		}

		checked <- auditedSecret{name: secret, content: sec.Password()}
	}
	done <- struct{}{}
}

func printAuditResults(ctx context.Context, m map[string][]string, format string, color func(format string, a ...interface{}) string) bool {
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

	foundWeakPasswords := printAuditResults(ctx, messages, "%s:\n", color.CyanString)
	if !foundWeakPasswords {
		out.Green(ctx, "No weak secrets detected.")
	}
	foundErrors := printAuditResults(ctx, errors, "%s:\n", color.RedString)

	if foundWeakPasswords || foundDuplicates || foundErrors {
		_ = notify.Notify(ctx, "gopass - audit", "Finished. Found weak passwords and/or duplicates")
		return fmt.Errorf("found weak passwords or duplicates")
	}

	_ = notify.Notify(ctx, "gopass - audit", "Finished. No weak passwords or duplicates found!")
	return nil
}
