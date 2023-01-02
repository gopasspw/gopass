package audit

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/gopasspw/gopass/internal/hashsum"
	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/internal/set"
)

type SecretReport struct {
	Name     string
	Errors   []error
	Warnings []string
	Age      time.Duration
}

type Report struct {
	sync.Mutex

	secrets map[string]SecretReport
	// SHA512(password) -> secret name
	duplicates map[string]*set.Set[string]
}

func (r *Report) AddPassword(name, pw string) {
	if name == "" || pw == "" {
		return
	}

	r.Lock()
	defer r.Unlock()

	r.duplicates[hashsum.SHA256Hex(pw)].Add(name)
}

func (r *Report) AddError(name string, e error) {
	if name == "" || e == nil {
		return
	}

	r.Lock()
	defer r.Unlock()

	s := r.secrets[name]
	s.Errors = append(s.Errors, e)
	r.secrets[name] = s
}

func (r *Report) SetAge(name string, age time.Duration) {
	if name == "" {
		return
	}

	r.Lock()
	defer r.Unlock()

	s := r.secrets[name]
	s.Age = age
	r.secrets[name] = s
}

func (r *Report) AddWarning(name, msg string) {
	if name == "" || msg == "" {
		return
	}

	r.Lock()
	defer r.Unlock()

	s := r.secrets[name]
	if s.Warnings == nil {
		s.Warnings = make([]string, 0, 1)
	}
	s.Warnings = append(s.Warnings, msg)
	r.secrets[name] = s
}

func newReport() *Report {
	return &Report{
		secrets:    make(map[string]SecretReport, 512),
		duplicates: make(map[string]*set.Set[string], 512),
	}
}

func (r *Report) finalize() {
	for k, s := range r.secrets {
		for _, secs := range r.duplicates {
			if secs.Contains(k) {
				// TODO: secs.Difference() - s.Warnings = append(s.Warnings, fmt.Sprintf("Duplicates detected. Shared with: %+v", secs.))
				_ = s
			}
		}
	}
}

func (r *Report) PrintResults(ctx context.Context) error {
	if r == nil {
		out.Warning(ctx, "Empty report")

		return nil
	}

	foundDuplicates := false
	for _, secrets := range r.duplicates {
		if secrets.Len() > 1 {
			foundDuplicates = true

			out.Printf(ctx, "Detected a shared secret for:")
			for _, secret := range secrets.Elements() {
				out.Printf(ctx, "\t- %s", secret)
			}
		}
	}
	if !foundDuplicates {
		out.Printf(ctx, "No shared secrets found.")
	}

	// TODO
	// foundWeakPasswords := printAuditResults(r.Warnings, "%s:\n", color.CyanString)
	// if !foundWeakPasswords {
	// 	out.Printf(ctx, "No weak secrets detected.")
	// }
	// foundErrors := printAuditResults(errors, "%s:\n", color.RedString)

	// if foundWeakPasswords || foundDuplicates || foundErrors {
	// 	_ = notify.Notify(ctx, "gopass - audit", "Finished. Found weak passwords and/or duplicates")

	// 	return fmt.Errorf("found weak passwords or duplicates")
	// }

	// _ = notify.Notify(ctx, "gopass - audit", "Finished. No weak passwords or duplicates found!")

	return nil
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
