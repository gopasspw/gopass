package tpl

import (
	"context"
	"crypto/md5"
	"crypto/sha1"
	"fmt"
	"text/template"

	"github.com/jsimonetti/pwscheme/md5crypt"
	"github.com/jsimonetti/pwscheme/ssha"
	"github.com/jsimonetti/pwscheme/ssha256"
	"github.com/jsimonetti/pwscheme/ssha512"
	"github.com/pkg/errors"
)

// These constants defined the template function names used
const (
	FuncMd5sum      = "md5sum"
	FuncSha1sum     = "sha1sum"
	FuncMd5Crypt    = "md5crypt"
	FuncSSHA        = "ssha"
	FuncSSHA256     = "ssha256"
	FuncSSHA512     = "ssha512"
	FuncGet         = "get"
	FuncGetPassword = "getpw"
	FuncGetValue    = "getval"
)

func md5sum() func(...string) (string, error) {
	return func(s ...string) (string, error) {
		return fmt.Sprintf("%x", md5.Sum([]byte(s[0]))), nil
	}
}

func sha1sum() func(...string) (string, error) {
	return func(s ...string) (string, error) {
		return fmt.Sprintf("%x", sha1.Sum([]byte(s[0]))), nil
	}
}

func md5cryptFunc() func(...string) (string, error) {
	return func(s ...string) (string, error) {
		return md5crypt.Generate(s[0], 4)
	}
}

func sshaFunc() func(...string) (string, error) {
	return func(s ...string) (string, error) {
		return ssha.Generate(s[0], 4)
	}
}

func ssha256Func() func(...string) (string, error) {
	return func(s ...string) (string, error) {
		return ssha256.Generate(s[0], 4)
	}
}

func ssha512Func() func(...string) (string, error) {
	return func(s ...string) (string, error) {
		return ssha512.Generate(s[0], 4)
	}
}

func get(ctx context.Context, kv kvstore) func(...string) (string, error) {
	return func(s ...string) (string, error) {
		if len(s) < 1 {
			return "", nil
		}
		if kv == nil {
			return "", errors.Errorf("KV is nil")
		}
		sec, err := kv.Get(ctx, s[0])
		if err != nil {
			return err.Error(), nil
		}
		return string(sec.Bytes()), nil
	}
}

func getPassword(ctx context.Context, kv kvstore) func(...string) (string, error) {
	return func(s ...string) (string, error) {
		if len(s) < 1 {
			return "", nil
		}
		if kv == nil {
			return "", errors.Errorf("KV is nil")
		}
		sec, err := kv.Get(ctx, s[0])
		if err != nil {
			return err.Error(), nil
		}
		return sec.Password(), nil
	}
}

func getValue(ctx context.Context, kv kvstore) func(...string) (string, error) {
	return func(s ...string) (string, error) {
		if len(s) < 2 {
			return "", nil
		}
		if kv == nil {
			return "", errors.Errorf("KV is nil")
		}
		sec, err := kv.Get(ctx, s[0])
		if err != nil {
			return err.Error(), nil
		}
		sv, found := sec.Get(s[1])
		if !found {
			return "", fmt.Errorf("key %q not found", s[1])
		}
		return sv, nil
	}
}

func funcMap(ctx context.Context, kv kvstore) template.FuncMap {
	return template.FuncMap{
		FuncGet:         get(ctx, kv),
		FuncGetPassword: getPassword(ctx, kv),
		FuncGetValue:    getValue(ctx, kv),
		FuncMd5sum:      md5sum(),
		FuncSha1sum:     sha1sum(),
		FuncMd5Crypt:    md5cryptFunc(),
		FuncSSHA:        sshaFunc(),
		FuncSSHA256:     ssha256Func(),
		FuncSSHA512:     ssha512Func(),
	}
}
