package pwrules

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"

	"github.com/gopasspw/gopass/pkg/appdir"
	"github.com/gopasspw/gopass/pkg/debug"
	"github.com/gopasspw/gopass/pkg/fsutil"
)

var (
	customAliases = map[string][]string{}
)

// LookupAliases looks up known aliases for the given domain
func LookupAliases(domain string) []string {
	aliases := make([]string, 0, len(genAliases[domain])+len(customAliases[domain]))
	aliases = append(aliases, genAliases[domain]...)
	aliases = append(aliases, customAliases[domain]...)
	sort.Strings(aliases)
	return aliases
}

// AllAliases returns all aliases
func AllAliases() map[string][]string {
	all := make(map[string][]string, len(genAliases)+len(customAliases))
	for k, v := range genAliases {
		all[k] = append(all[k], v...)
	}
	for k, v := range customAliases {
		all[k] = append(all[k], v...)
	}
	return all
}

func init() {
	if err := loadCustomAliases(); err != nil {
		debug.Log("failed to load custom aliases: %s", err)
	}
}

func filename() string {
	return filepath.Join(appdir.UserConfig(), "domain-aliases.json")
}

func loadCustomAliases() error {
	fn := filename()
	if !fsutil.IsFile(fn) {
		debug.Log("no custom aliases found at %s", fn)
		return nil
	}
	fh, err := os.Open(fn)
	if err != nil {
		return err
	}
	defer fh.Close()
	return json.NewDecoder(fh).Decode(&customAliases)
}

func saveCustomAliases() error {
	fn := filename()
	if err := os.MkdirAll(filepath.Dir(fn), 0700); err != nil {
		return err
	}
	fh, err := os.OpenFile(fn, os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		return err
	}
	defer fh.Close()
	return json.NewEncoder(fh).Encode(customAliases)
}

// AddCustomAlias adds a custom alias
func AddCustomAlias(domain, alias string) error {
	if len(customAliases) < 1 {
		loadCustomAliases()
	}
	v := make([]string, 0, 1)
	if ev, found := customAliases[domain]; found {
		v = ev
	}
	for _, k := range v {
		if k == alias {
			return nil
		}
	}
	v = append(v, alias)
	sort.Strings(v)
	customAliases[domain] = v

	return saveCustomAliases()
}

// RemoveCustomAlias removes a custom alias
func RemoveCustomAlias(domain, alias string) error {
	if len(customAliases) < 1 {
		loadCustomAliases()
	}
	ev, found := customAliases[domain]
	if !found {
		return nil
	}
	nv := make([]string, 0, len(ev)-1)
	for _, a := range ev {
		if alias == a {
			continue
		}
		nv = append(nv, a)
	}
	customAliases[domain] = nv

	return saveCustomAliases()
}

// DeleteCustomAlias removes a whole domain
func DeleteCustomAlias(domain string) error {
	if len(customAliases) < 1 {
		loadCustomAliases()
	}
	delete(customAliases, domain)

	return saveCustomAliases()
}
