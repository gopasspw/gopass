package action

import (
	"context"
	"errors"
	"fmt"

	"github.com/fatih/color"
	"github.com/muesli/crunchy"
	"github.com/muesli/goprogressbar"
	"github.com/urfave/cli"
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

// Audit validates passwords against common flaws
func (s *Action) Audit(ctx context.Context, c *cli.Context) error {
	t, err := s.Store.Tree()
	if err != nil {
		return s.exitError(ctx, ExitList, err, "failed to get store tree: %s", err)
	}
	list := t.List(0)

	fmt.Printf("Checking %d secrets. This may take some time ...\n", len(list))

	// Secrets that still need auditing.
	secrets := make(chan string)

	// Secrets that have been audited.
	checked := make(chan auditedSecret)

	// Spawn workers that run the auditing of all secrets concurrently.
	validator := crunchy.NewValidator()
	for jobs := 0; jobs < c.Int("jobs"); jobs++ {
		go s.audit(ctx, validator, secrets, checked)
	}

	go func() {
		for _, secret := range list {
			secrets <- secret
		}
		close(secrets)
	}()

	duplicates := make(map[string][]string)
	messages := make(map[string][]string)
	errors := make(map[string][]string)

	bar := &goprogressbar.ProgressBar{
		Total: int64(len(list)),
		Width: 120,
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

		if i == len(list) {
			break
		}
	}
	close(checked)
	fmt.Println() // Print empty line after the progressbar.

	foundDuplicates := false
	for _, secrets := range duplicates {
		if len(secrets) > 1 {
			foundDuplicates = true

			fmt.Println(color.CyanString("Detected a shared secret for:"))
			for _, secret := range secrets {
				fmt.Println(color.CyanString("\t- %s", secret))
			}
		}
	}
	if !foundDuplicates {
		fmt.Println(color.GreenString("No shared secrets found."))
	}

	foundWeakPasswords := printAuditResults(ctx, messages, "%s:\n", color.CyanString)
	if !foundWeakPasswords {
		fmt.Println(color.GreenString("No weak secrets detected."))
	}
	foundErrors := printAuditResults(ctx, errors, "%s:\n", color.RedString)

	if foundWeakPasswords || foundDuplicates || foundErrors {
		return s.exitError(ctx, ExitAudit, nil, "found weak passwords or duplicates")
	}

	return nil
}

func (s *Action) audit(ctx context.Context, validator *crunchy.Validator, secrets <-chan string, checked chan<- auditedSecret) {
	for secret := range secrets {
		// check for context cancelation
		select {
		case <-ctx.Done():
			checked <- auditedSecret{name: secret, content: "", err: errors.New("user aborted")}
			continue
		default:
		}

		sec, err := s.Store.Get(ctx, secret)
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
}

func printAuditResults(ctx context.Context, m map[string][]string, format string, color func(format string, a ...interface{}) string) bool {
	b := false

	for msg, secrets := range m {
		b = true
		fmt.Print(color(format, msg))
		for _, secret := range secrets {
			fmt.Print(color("\t- %s\n", secret))
		}
	}

	return b
}

func printAuditResult(pw string) {
	validator := crunchy.NewValidator()
	if err := validator.Check(pw); err != nil {
		fmt.Println(color.CyanString(fmt.Sprintf("Warning: %s", err)))
	}
}
