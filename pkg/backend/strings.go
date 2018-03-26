package backend

var (
	cryptoNameToBackendMap = map[string]CryptoBackend{
		"plain":   Plain,
		"gpgcli":  GPGCLI,
		"xc":      XC,
		"openpgp": OpenPGP,
		"vault":   Vault,
	}
	cryptoBackendToNameMap = map[CryptoBackend]string{}
	rcsNameToBackendMap    = map[string]RCSBackend{
		"gitcli": GitCLI,
		"noop":   Noop,
		"gogit":  GoGit,
	}
	rcsBackendToNameMap     = map[RCSBackend]string{}
	storageNameToBackendMap = map[string]StorageBackend{
		"inmem":  InMem,
		"fs":     FS,
		"consul": Consul,
	}
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
