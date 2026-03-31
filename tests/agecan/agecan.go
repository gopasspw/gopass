package agecan

import (
	"embed"
	"os"
	"path/filepath"
)

//go:embed age-identity.txt age-recipient.txt
var ageFS embed.FS

const TestPin = "test"

func Identity() string {
	data, err := ageFS.ReadFile("age-identity.txt")
	if err != nil {
		return ""
	}

	return string(data)
}

func Recipient() string {
	data, err := ageFS.ReadFile("age-recipient.txt")
	if err != nil {
		return ""
	}

	return string(data)
}

func Setup(homedir string) (string, error) {
	idPath := filepath.Join(homedir, ".config", "gopass", "age", "identities")
	if err := os.MkdirAll(filepath.Dir(idPath), 0o700); err != nil {
		return "", err
	}

	if err := os.WriteFile(idPath, []byte(Identity()), 0o600); err != nil {
		return "", err
	}

	return Recipient(), nil
}
