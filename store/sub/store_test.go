package sub

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func createStore(dir string) ([]string, []string, error) {
	recipients := []string{
		"0xDEADBEEF",
		"0xFEEDBEEF",
	}
	list := []string{
		"foo/bar/baz",
		"baz/ing/a",
	}
	sort.Strings(list)
	for _, file := range list {
		filename := filepath.Join(dir, file+".gpg")
		if err := os.MkdirAll(filepath.Dir(filename), 0700); err != nil {
			return recipients, list, err
		}
		if err := ioutil.WriteFile(filename, []byte{}, 0644); err != nil {
			return recipients, list, err
		}
	}
	err := ioutil.WriteFile(filepath.Join(dir, GPGID), []byte(strings.Join(recipients, "\n")), 0600)
	return recipients, list, err
}
