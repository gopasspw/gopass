package backend

var (
	cryptoNameToBackendMap = map[string]CryptoBackend{
		"gpgmock": GPGMock,
		"gpgcli":  GPGCLI,
		"xc":      XC,
		"openpgp": OpenPGP,
	}
	cryptoBackendToNameMap = map[CryptoBackend]string{}
	syncNameToBackendMap   = map[string]SyncBackend{
		"gitcli":  GitCLI,
		"gitmock": GitMock,
		"gogit":   GoGit,
	}
	syncBackendToNameMap  = map[SyncBackend]string{}
	storeNameToBackendMap = map[string]StoreBackend{
		"kvmock": KVMock,
		"fs":     FS,
	}
	storeBackendToNameMap = map[StoreBackend]string{}
)

func init() {
	for k, v := range cryptoNameToBackendMap {
		cryptoBackendToNameMap[v] = k
	}
	for k, v := range syncNameToBackendMap {
		syncBackendToNameMap[v] = k
	}
	for k, v := range storeNameToBackendMap {
		storeBackendToNameMap[v] = k
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

func syncBackendFromName(name string) SyncBackend {
	if b, found := syncNameToBackendMap[name]; found {
		return b
	}
	return -1
}

func syncNameFromBackend(be SyncBackend) string {
	if b, found := syncBackendToNameMap[be]; found {
		return b
	}
	return ""
}

func storeBackendFromName(name string) StoreBackend {
	if b, found := storeNameToBackendMap[name]; found {
		return b
	}
	return FS
}

func storeNameFromBackend(be StoreBackend) string {
	if b, found := storeBackendToNameMap[be]; found {
		return b
	}
	return ""
}
