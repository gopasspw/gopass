package backend

import "sort"

var (
	cryptoNameToBackendMap  = map[string]CryptoBackend{}
	cryptoBackendToNameMap  = map[CryptoBackend]string{}
	storageNameToBackendMap = map[string]StorageBackend{}
	storageBackendToNameMap = map[StorageBackend]string{}
)

func init() {
	for k, v := range cryptoNameToBackendMap {
		cryptoBackendToNameMap[v] = k
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

// StorageBackends returns the list of registered storage backends.
func StorageBackends() []string {
	bes := make([]string, 0, len(storageNameToBackendMap))
	for k := range storageNameToBackendMap {
		bes = append(bes, k)
	}
	sort.Strings(bes)
	return bes
}

// CryptoBackendFromName parses the identifier into a crypto backend
func CryptoBackendFromName(name string) CryptoBackend {
	if name == "gpg" {
		name = "gpgcli"
	}
	if b, found := cryptoNameToBackendMap[name]; found {
		return b
	}
	return -1
}

// CryptoNameFromBackend returns the name of a given crypto backend
func CryptoNameFromBackend(be CryptoBackend) string {
	if b, found := cryptoBackendToNameMap[be]; found {
		return b
	}
	return ""
}

// StorageBackendFromName parses the identifier into a storage backend
func StorageBackendFromName(name string) StorageBackend {
	if b, found := storageNameToBackendMap[name]; found {
		return b
	}
	return FS
}

// StorageNameFromBackend returns the name of a given storage backend
func StorageNameFromBackend(be StorageBackend) string {
	if b, found := storageBackendToNameMap[be]; found {
		return b
	}
	return ""
}
