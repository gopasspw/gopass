// +build windows
package cli

import (
	"errors"
	"path/filepath"

	"github.com/justwatchcom/gopass/utils/fsutil"

	"golang.org/x/sys/windows/registry"
)

func (g *GPG) detectBinary(bin string) error {
	// set default
	g.binary = "gpg.exe"

	// try to detect location
	k, err := registry.OpenKey(registry.LOCAL_MACHINE, `SOFTWARE\GnuPG`, registry.QUERY_VALUE|registry.WOW64_32KEY)
	if err != nil {
		return err
	}

	v, _, err := k.GetStringValue("Install Directory")
	if err != nil {
		return err
	}

	// gpg.exe for GPG4Win 3.0.0; would be gpg2.exe for 2.x
	for _, b := range []string{bin, "gpg.exe", "gpg2.exe"} {
		gpgPath := filepath.Join(v, "bin", b)
		if fsutil.IsFile(gpgPath) {
			g.binary = gpgPath
			return nil
		}
	}
	return errors.New("gpg.exe not found")
}
