package action

import (
	"fmt"
	"os"
	"runtime"

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
func (s *Action) Audit(c *cli.Context) error {
	t, err := s.Store.Tree()
	if err != nil {
		return err
	}
	list := t.List(0)

	fmt.Printf("Checking %d secrets. This may take some time ...\n", len(list))

	// Jobs are the secrets that still need auditing.
	jobs := make(chan string)

	// Secrets that have been audited.
	checked := make(chan auditedSecret)

	// Spawn workers that run the auditing of all secrets concurrently.
	validator := crunchy.NewValidator()
	for worker := 0; worker < runtime.NumCPU(); worker++ {
		go s.audit(validator, jobs, checked)
	}

	go func() {
		for _, secret := range list {
			jobs <- secret
		}
		close(jobs)
	}()

	duplicates := make(map[string][]string)
	messages := make(map[string][]string)

	bar := &goprogressbar.ProgressBar{
		Total: int64(len(list)),
		Width: 120,
	}

	i := 0
	for secret := range checked {
		if secret.err == nil {
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
	fmt.Println("") // Print empty line after the progressbar.

	foundDuplicates := false
	for _, secrets := range duplicates {
		if len(secrets) > 1 {
			foundDuplicates = true

			fmt.Printf(color.CyanString("Detected a shared secret for:\n"))
			for _, secret := range secrets {
				fmt.Printf(color.CyanString("\t- %s\n", secret))
			}
		}
	}

	if !foundDuplicates {
		fmt.Println(color.GreenString("No shared secrets found."))
	}

	foundWeakPasswords := false
	for msg, secrets := range messages {
		foundWeakPasswords = true
		fmt.Printf(color.CyanString("%s:\n", msg))
		for _, secret := range secrets {
			fmt.Printf(color.CyanString("\t- %s\n", secret))
		}
	}

	if !foundWeakPasswords {
		fmt.Println(color.GreenString("No weak secrets detected."))
	}

	if foundWeakPasswords || foundDuplicates {
		os.Exit(1)
	}

	return nil
}

func (s *Action) audit(validator *crunchy.Validator, secrets <-chan string, checked chan<- auditedSecret) {
	for secret := range secrets {
		content, err := s.Store.GetFirstLine(secret)
		if err != nil {
			checked <- auditedSecret{name: secret, content: string(content), err: err, message: err.Error()}
			continue
		}

		if err := validator.Check(string(content)); err != nil {
			checked <- auditedSecret{name: secret, content: string(content), message: err.Error()}
			continue
		}

		checked <- auditedSecret{name: secret, content: string(content)}
	}
}
