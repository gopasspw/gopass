package config

import "testing"

func TestConfigs(t *testing.T) {
	for _, cfg := range []string{
		`root:
  askformore: false
  autoimport: false
  autosync: false
  cliptimeout: 45
  noconfirm: false
  nopager: false
  path: /home/johndoe/.password-store
  safecontent: false
mounts:
  foo/sub:
    askformore: false
    autoimport: false
    autosync: false
    cliptimeout: 45
    noconfirm: false
    nopager: false
    path: /home/johndoe/.password-store-foo-sub
    safecontent: false
  work:
    askformore: false
    autoimport: false
    autosync: false
    cliptimeout: 45
    noconfirm: false
    nopager: false
    path: /home/johndoe/.password-store-work
    safecontent: false
version: 1.4.0`,
		`askformore: false
autoimport: true
autosync: false
cliptimeout: 45
mounts:
  dev: /Users/johndoe/.password-store-dev
  ops: /Users/johndoe/.password-store-ops
  personal: /Users/johndoe/secrets
  teststore: /Users/johndoe/tmp/teststore
noconfirm: false
path: /home/tex/.password-store
safecontent: true
version: "1.3.0"`,
		`alwaystrust: true
askformore: false
autoimport: true
autopull: true
autopush: true
cliptimeout: 45
debug: false
loadkeys: true
mounts:
  dev: /Users/johndoe/.password-store-dev
  ops: /Users/johndoe/.password-store-ops
  personal: /Users/johndoe/secrets
  teststore: /Users/johndoe/tmp/teststore
nocolor: false
noconfirm: false
path: /home/tex/.password-store
persistkeys: true
safecontent: true
version: "1.2.0"`,
		`alwaystrust: false
autoimport: false
autopull: true
autopush: true
cliptimeout: 45
loadkeys: false
mounts:
  dev: /home/johndoe/.password-store-dev
  ops: /home/johndoe/.password-store-ops
  personal: /home/johndoe/secrets
  teststore: /home/johndoe/tmp/teststore
nocolor: false
noconfirm: false
path: /home/johndoe/.password-store
persistkeys: true
safecontent: false
version: 1.1.0`,
		`alwaystrust: false
autoimport: false
autopull: true
autopush: false
cliptimeout: 45
loadkeys: false
mounts:
  dev: /Users/johndoe/.password-store-dev
  ops: /Users/johndoe/.password-store-ops
  personal: /Users/johndoe/secrets
  teststore: /Users/johndoe/tmp/teststore
noconfirm: false
path: /home/tex/.password-store
persistkeys: false
version: "1.0.0"`,
	} {
		if _, err := decode([]byte(cfg)); err != nil {
			t.Errorf("Failed to load config: %s\n%s", err, cfg)
		}
	}
}
