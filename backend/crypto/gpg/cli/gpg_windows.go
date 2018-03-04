// +build windows

package cli

import (
	"os/exec"
	"path/filepath"

	"github.com/justwatchcom/gopass/utils/fsutil"

	"golang.org/x/sys/windows/registry"
)

func detectBinaryCandidates(bin string) ([]string, error) {
	// gpg.exe for GPG4Win 3.0.0; would be gpg2.exe for 2.x
	bins := make([]string, 0, 4)

	bins, err := searchRegistry(bin, bins)
	if err != nil {
		return bins, err
	}

	bins, err = searchPath(bin, bins)
	if err != nil {
		return bins, err
	}

	return bins, nil
}

func searchRegistry(bin string, bins []string) ([]string, error) {
	// try to detect location of installed GPG4Win
	k, err := registry.OpenKey(registry.LOCAL_MACHINE, `SOFTWARE\GnuPG`, registry.QUERY_VALUE|registry.WOW64_32KEY)
	if err != nil {
		return bins, nil
	}

	if v, _, err := k.GetStringValue("Install Directory"); err == nil && v != "" {
		for _, b := range []string{bin, "gpg2.exe", "gpg.exe"} {
			gpgPath := filepath.Join(v, "bin", b)
			if fsutil.IsFile(gpgPath) {
				bins = append(bins, gpgPath)
			}
		}
	}

	return bins, nil
}

func searchPath(bin string, bins []string) ([]string, error) {
	// try to detect location for GPG installed somewhere on the PATH
	for _, b := range []string{bin, "gpg2.exe", "gpg.exe"} {
		gpgPath, err := exec.LookPath(b)
		if err != nil {
			continue
		}
		if fsutil.IsFile(gpgPath) {
			bins = append(bins, gpgPath)
		}
	}

	return bins, nil
}
