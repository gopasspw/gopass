package tpl

import (
	"crypto/md5"
	"crypto/sha1"
	"fmt"
	"text/template"
)

// These constants defined the template function names used
const (
	FuncMd5sum  = "md5sum"
	FuncSha1sum = "sha1sum"
	FuncGet     = "get"
)

func md5sum() func(...string) (string, error) {
	return func(s ...string) (string, error) {
		return fmt.Sprintf("%x", md5.Sum([]byte(s[0]))), nil
	}
}

func sha1sum() func(...string) (string, error) {
	return func(s ...string) (string, error) {
		if len(s) < 2 {
			return "", nil
		}
		return fmt.Sprintf("%x", sha1.Sum([]byte(s[1]))), nil
	}
}

func get(kv kvstore) func(...string) (string, error) {
	return func(s ...string) (string, error) {
		if len(s) < 2 {
			return "", nil
		}
		buf, err := kv.Get(s[1])
		return string(buf), err
	}
}

func funcMap(kv kvstore) template.FuncMap {
	return template.FuncMap{
		FuncGet:     get(kv),
		FuncMd5sum:  md5sum(),
		FuncSha1sum: sha1sum(),
	}
}
