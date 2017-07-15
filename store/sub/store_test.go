package sub

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func createStore(dir string, recipients, entries []string) ([]string, []string, error) {
	if recipients == nil {
		recipients = []string{
			"0xDEADBEEF",
			"0xFEEDBEEF",
		}
	}
	if entries == nil {
		entries = []string{
			"foo/bar/baz",
			"baz/ing/a",
		}
	}
	sort.Strings(entries)
	for _, file := range entries {
		filename := filepath.Join(dir, file+".gpg")
		if err := os.MkdirAll(filepath.Dir(filename), 0700); err != nil {
			return recipients, entries, err
		}
		if err := ioutil.WriteFile(filename, []byte{}, 0644); err != nil {
			return recipients, entries, err
		}
	}
	err := ioutil.WriteFile(filepath.Join(dir, GPGID), []byte(strings.Join(recipients, "\n")), 0600)
	return recipients, entries, err
}
