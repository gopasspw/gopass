package backend

import "sort"

var (
	cryptoNameToBackendMap = map[string]CryptoBackend{
		"plain":   0,
		"gpgcli":  1,
		"xc":      2,
		"openpgp": 3,
		"vault":   4,
	}
	rcsNameToBackendMap = map[string]RCSBackend{
		"noop":   0,
		"gitcli": 1,
		"gogit":  2,
	}
	storageNameToBackendMap = map[string]StorageBackend{
		"fs":     0,
		"inmem":  1,
		"consul": 2,
	}

	cryptoBackendToNameMap  = map[CryptoBackend]string{}
	rcsBackendToNameMap     = map[RCSBackend]string{}
	storageBackendToNameMap = map[StorageBackend]string{}
)

func init() {
	for k, v := range cryptoNameToBackendMap {
		cryptoBackendToNameMap[v] = k
	}
	for k, v := range rcsNameToBackendMap {
		rcsBackendToNameMap[v] = k
	}
	for k, v := range storageNameToBackendMap {
		storageBackendToNameMap[v] = k
	}
}

// CryptoBackends returns the list of registered crypto backends.
func CryptoBackends() []string {
	bes := make([]string, 0, len(cryptoNameToBackendMap))
	for k := range cryptoNameToBackendMap {
		bes = append(bes, k)
	}
	sort.Strings(bes)
	return bes
}

// RCSBackends returns the list of registered RCS backends.
func RCSBackends() []string {
	bes := make([]string, 0, len(rcsNameToBackendMap))
	for k := range rcsNameToBackendMap {
		bes = append(bes, k)
	}
	sort.Strings(bes)
	return bes
}

// StorageBackends returns the list of registered storage backends.
func StorageBackends() []string {
	bes := make([]string, 0, len(storageNameToBackendMap))
	for k := range storageNameToBackendMap {
		bes = append(bes, k)
	}
	sort.Strings(bes)
	return bes
}

func cryptoBackendFromName(name string) CryptoBackend {
	if b, found := cryptoNameToBackendMap[name]; found {
		return b
	}
	return -1
}

func cryptoNameFromBackend(be CryptoBackend) string {
	if b, found := cryptoBackendToNameMap[be]; found {
		return b
	}
	return ""
}

func rcsBackendFromName(name string) RCSBackend {
	if b, found := rcsNameToBackendMap[name]; found {
		return b
	}
	return -1
}

func rcsNameFromBackend(be RCSBackend) string {
	if b, found := rcsBackendToNameMap[be]; found {
		return b
	}
	return ""
}

func storageBackendFromName(name string) StorageBackend {
	if b, found := storageNameToBackendMap[name]; found {
		return b
	}
	return FS
}

func storageNameFromBackend(be StorageBackend) string {
	if b, found := storageBackendToNameMap[be]; found {
		return b
	}
	return ""
}
