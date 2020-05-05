package age

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/gopasspw/gopass/pkg/fsutil"
)

type ghCache struct{}

func (g *ghCache) cacheDir() string {
	ucd, err := os.UserCacheDir()
	if err != nil {
		return ""
	}
	d := filepath.Join(ucd, "gopass", "age", "github")
	if err := os.MkdirAll(d, 0644); err != nil {
		return ""
	}
	return d
}

func (g *ghCache) Get(key string) []string {
	key = fsutil.CleanFilename(key)
	buf, err := ioutil.ReadFile(filepath.Join(g.cacheDir(), key))
	if err != nil {
		return nil
	}
	return strings.Split(string(buf), "\n")
}

func (g *ghCache) Set(key string, value []string) {
	key = fsutil.CleanFilename(key)
	ioutil.WriteFile(filepath.Join(g.cacheDir(), key), []byte(strings.Join(value, "\n")), 0644)
}
