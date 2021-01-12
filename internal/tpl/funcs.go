package tpl

import (
	"context"
	"crypto/md5"
	"crypto/sha1"
	"fmt"
	"strconv"
	"text/template"

	"github.com/gopasspw/gopass/internal/debug"
	"github.com/gopasspw/gopass/internal/pwschemes/argon2i"
	"github.com/gopasspw/gopass/internal/pwschemes/argon2id"
	"github.com/gopasspw/gopass/internal/pwschemes/bcrypt"
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
	FuncArgon2i     = "argon2i"
	FuncArgon2id    = "argon2id"
	FuncBcrypt      = "bcrypt"
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

// saltLen tries to parse the given string into a numeric salt length.
// NOTE: This is on of the rare cases where I think named returns
// are useful.
func saltLen(s []string) (saltLen int) {
	defer func() {
		debug.Log("using saltLen %d", saltLen)
	}()

	// default should be 32bit
	saltLen = 32

	if len(s) < 2 {
		return
	}

	i, err := strconv.Atoi(s[0])
	if err == nil && i > 0 {
		saltLen = i
	}
	if err != nil {
		debug.Log("failed to parse saltLen %+v: %q", s, err)
	}
	return
}

func md5cryptFunc() func(...string) (string, error) {
	return func(s ...string) (string, error) {
		sl := uint8(saltLen(s))
		if sl > 8 || sl < 1 {
			sl = 4
		}
		return md5crypt.Generate(s[0], sl)
	}
}

func sshaFunc() func(...string) (string, error) {
	return func(s ...string) (string, error) {
		return ssha.Generate(s[0], uint8(saltLen(s)))
	}
}

func ssha256Func() func(...string) (string, error) {
	return func(s ...string) (string, error) {
		return ssha256.Generate(s[0], uint8(saltLen(s)))
	}
}

func ssha512Func() func(...string) (string, error) {
	return func(s ...string) (string, error) {
		return ssha512.Generate(s[0], uint8(saltLen(s)))
	}
}

func argon2iFunc() func(...string) (string, error) {
	return func(s ...string) (string, error) {
		return argon2i.Generate(s[0], uint8(saltLen(s)))
	}
}

func argon2idFunc() func(...string) (string, error) {
	return func(s ...string) (string, error) {
		return argon2id.Generate(s[0], uint8(saltLen(s)))
	}
}

func bcryptFunc() func(...string) (string, error) {
	return func(s ...string) (string, error) {
		return bcrypt.Generate(s[0])
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
		FuncArgon2i:     argon2iFunc(),
		FuncArgon2id:    argon2idFunc(),
		FuncBcrypt:      bcryptFunc(),
	}
}
