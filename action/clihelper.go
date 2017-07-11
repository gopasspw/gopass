package action

import (
	"bufio"
	"fmt"
	"os"
	"os/signal"
	"sort"
	"strconv"
	"strings"

	"golang.org/x/crypto/ssh/terminal"
)

const (
	maxTries = 42
)

// confirmRecipients asks the user to confirm a given set of recipients
func (s *Action) confirmRecipients(name string, recipients []string) ([]string, error) {
	if s.Store.NoConfirm() || !s.isTerm {
		return recipients, nil
	}

	fmt.Printf("gopass: Encrypting %s for these recipients:\n", name)
	sort.Strings(recipients)
	for _, r := range recipients {
		kl, err := s.gpg.FindPublicKeys(r)
		if err != nil {
			fmt.Println(err)
			continue
		}
		if len(kl) < 1 {
			fmt.Println("key not found", r)
			continue
		}
		fmt.Printf(" - %s\n", kl[0].OneLine())
	}
	fmt.Println("")

	yes, err := s.askForBool("Do you want to continue?", true)
	if err != nil {
		return recipients, err
	}

	if yes {
		return recipients, nil
	}

	return recipients, fmt.Errorf("user aborted")
}

// askForConfirmation asks a yes/no question until the user
// replies yes or no
func (s *Action) askForConfirmation(text string) bool {
	for i := 0; i < maxTries; i++ {
		if choice, err := s.askForBool(text, false); err == nil {
			return choice
		}
	}
	return false
}

// askForBool ask for a bool (yes or no) exactly once.
// The empty answer uses the specified default, any other answer
// is an error.
func (s *Action) askForBool(text string, def bool) (bool, error) {
	choices := "y/N"
	if def {
		choices = "Y/n"
	}

	str, err := s.askForString(text, choices)
	if err != nil {
		return false, err
	}
	switch str {
	case "Y/n":
		return true, nil
	case "y/N":
		return false, nil
	}

	str = strings.ToLower(string(str[0]))
	switch str {
	case "y":
		return true, nil
	case "n":
		return false, nil
	default:
		return false, fmt.Errorf("Unknown answer: %s", str)
	}
}

// askForString asks for a string once, using the default if the
// anser is empty. Errors are only returned on I/O errors
func (s *Action) askForString(text, def string) (string, error) {
	reader := bufio.NewReader(os.Stdin)

	fmt.Printf("%s [%s]: ", text, def)
	input, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	input = strings.TrimSpace(input)
	if input == "" {
		input = def
	}
	return input, nil
}

// askForInt asks for an valid interger once. If the input
// can not be converted to an int it returns an error
func (s *Action) askForInt(text string, def int) (int, error) {
	str, err := s.askForString(text, strconv.Itoa(def))
	if err != nil {
		return 0, err
	}
	intVal, err := strconv.Atoi(str)
	if err != nil {
		return 0, err
	}
	return intVal, nil
}

// askForPassword prompts for a password twice until both match
func (s *Action) askForPassword(name string, askFn func(string) (string, error)) (string, error) {
	if !s.isTerm {
		return "", fmt.Errorf("impossible without terminal")
	}
	if askFn == nil {
		askFn = s.promptPass
	}
	for i := 0; i < maxTries; i++ {
		pass, err := askFn(fmt.Sprintf("Enter password for %s", name))
		if err != nil {
			return "", err
		}

		passAgain, err := askFn(fmt.Sprintf("Retype password for %s", name))
		if err != nil {
			return "", err
		}

		if pass == passAgain {
			return strings.TrimSpace(pass), nil
		}

		fmt.Println("Error: the entered password do not match")
	}
	return "", fmt.Errorf("no valid user input")
}

// askForKeyImport asks for permissions to import the named key
func (s *Action) askForKeyImport(key string) bool {
	if !s.isTerm {
		return false
	}
	ok, err := s.askForBool("Do you want to import the public key '%s' into your keyring?", false)
	if err != nil {
		return false
	}
	return ok
}

// askforPrivateKey promts the user to select from a list of private keys
func (s *Action) askForPrivateKey(prompt string) (string, error) {
	if !s.isTerm {
		return "", fmt.Errorf("no interaction without terminal")
	}
	kl, err := s.gpg.ListPrivateKeys()
	if err != nil {
		return "", err
	}
	kl = kl.UseableKeys()
	if len(kl) < 1 {
		return "", fmt.Errorf("No useable private keys found")
	}
	for i := 0; i < maxTries; i++ {
		fmt.Println(prompt)
		for i, k := range kl {
			fmt.Printf("[%d] %s\n", i, k.OneLine())
		}
		iv, err := s.askForInt(fmt.Sprintf("Please enter the number of a key (0-%d)", len(kl)-1), 0)
		if err != nil {
			continue
		}
		if iv >= 0 && iv < len(kl) {
			return kl[iv].Fingerprint, nil
		}
	}
	return "", fmt.Errorf("no valid user input")
}

// askForGitConfigUser will iterate over GPG private key identities and return one identity's name
// and/or one identity's email address if selected by the user for being used as values for
// git config user.name and git config user.email, respectively.
func (s *Action) askForGitConfigUser() (string, string, error) {
	var (
		name string
		email string
	)

	if !s.isTerm {
		return "", "", fmt.Errorf("no interaction without terminal")
	}
	keyList, err := s.gpg.ListPrivateKeys()
	if err != nil {
		return "", "", err
	}
	keyList = keyList.UseableKeys()
	if len(keyList) < 1 {
		return "", "", fmt.Errorf("No usable private keys found")
	}

	for _, key := range keyList {
		for _, identity := range key.Identities {
			ok, err := s.askForBool(fmt.Sprintf("Use %q as user name for password store git config?", identity.Name), false)
			if err != nil {
				return "", "", err
			}
			if ok {
				name = identity.Name
			}
			ok, err = s.askForBool(fmt.Sprintf("Use %q as email address for password store git config?", identity.Email), false)
			if err != nil {
				return "", "", err
			}
			if ok {
				email = identity.Email
			}
			if name != "" && email != "" {
				break
			}
		}
	}

	return name, email, nil
}

// promptPass will prompt user's for a password by terminal.
func (s *Action) promptPass(prompt string) (pass string, err error) {
	if !s.isTerm {
		return
	}
	// Make a copy of STDIN's state to restore afterward
	fd := int(os.Stdin.Fd())
	oldState, err := terminal.GetState(fd)
	if err != nil {
		return "", fmt.Errorf("Could not get state of terminal: %s", err)
	}
	defer func() {
		if err := terminal.Restore(fd, oldState); err != nil {
			fmt.Printf("Failed to restore terminal: %s\n", err)
		}
	}()

	// Restore STDIN in the event of a signal interruption
	sigch := make(chan os.Signal, 1)
	signal.Notify(sigch, os.Interrupt)
	go func() {
		for range sigch {
			if err := terminal.Restore(fd, oldState); err != nil {
				fmt.Printf("Failed to restore terminal: %s\n", err)
			}
			os.Exit(1)
		}
	}()

	fmt.Printf("%s: ", prompt)
	passBytes, err := terminal.ReadPassword(fd)
	fmt.Println("")
	return string(passBytes), err
}
