package agecan

import (
	"embed"
	"fmt"
)

//go:embed age-identity.txt age-recipient.txt age-identity-passphrase.txt age-recipient-passphrase.txt
var ageFS embed.FS

const TestPin = "passphrasesecret"

func Identity(passPhrase bool) (string, error) {
	filename := "age-identity.txt"
	if passPhrase {
		filename = "age-identity-passphrase.txt"
	}

	data, err := ageFS.ReadFile(filename)
	if err != nil {
		return "", fmt.Errorf("failed to read embedded %s: %w", filename, err)
	}

	return string(data), nil
}

func Recipient(passPhrase bool) (string, error) {
	filename := "age-recipient.txt"
	if passPhrase {
		filename = "age-recipient-passphrase.txt"
	}

	data, err := ageFS.ReadFile(filename)
	if err != nil {
		return "", fmt.Errorf("failed to read embedded %s: %w", filename, err)
	}

	return string(data), nil
}
