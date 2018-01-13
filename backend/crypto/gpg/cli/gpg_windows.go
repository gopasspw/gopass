// +build windows

package cli

import (
	"path/filepath"

	"github.com/justwatchcom/gopass/utils/fsutil"

	"golang.org/x/sys/windows/registry"
)

func (g *GPG) detectBinaryCandidates(bin string) ([]string, error) {
	// gpg.exe for GPG4Win 3.0.0; would be gpg2.exe for 2.x
	bins := make([]string, 0, 4)

	// try to detect location
	k, err := registry.OpenKey(registry.LOCAL_MACHINE, `SOFTWARE\GnuPG`, registry.QUERY_VALUE|registry.WOW64_32KEY)
	if err != nil {
		return bins, err
	}

	v, _, err := k.GetStringValue("Install Directory")
	if err != nil {
		return bins, err
	}

	for _, b := range []string{bin, "gpg2.exe", "gpg.exe"} {
		gpgPath := filepath.Join(v, "bin", b)
		if fsutil.IsFile(gpgPath) {
			bins = append(bins, gpgPath)
		}
	}
	return bins, nil
}
