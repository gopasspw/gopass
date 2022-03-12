package tpl

import (
	"context"
	"crypto/md5"
	"crypto/sha1"
	"fmt"
	"strconv"
	"text/template"

	"github.com/gopasspw/gopass/internal/pwschemes/argon2i"
	"github.com/gopasspw/gopass/internal/pwschemes/argon2id"
	"github.com/gopasspw/gopass/internal/pwschemes/bcrypt"
	"github.com/gopasspw/gopass/pkg/debug"
	"github.com/jsimonetti/pwscheme/md5crypt"
	"github.com/jsimonetti/pwscheme/ssha"
	"github.com/jsimonetti/pwscheme/ssha256"
	"github.com/jsimonetti/pwscheme/ssha512"
)

// These constants defined the template function names used.
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
	FuncGetValues   = "getvals"
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
func saltLen(s []string) uint8 {
	if len(s) < 2 {
		debug.Log("no salt length given, using default %d", 32)

		return 32
	}

	i, err := strconv.ParseUint(s[0], 10, 8)
	if err != nil {
		debug.Log("failed to parse saltLen %+v: %q. using default: %d", s, err, 32)

		return 32
	}

	sl := uint8(i)

	debug.Log("using saltLen %d", sl)

	return sl
}

func md5cryptFunc() func(...string) (string, error) {
	// parameters: s[0] = salt, s[-1] = password
	return func(s ...string) (string, error) {
		if len(s) < 1 {
			return "", fmt.Errorf("usage: %s <salt> <password>", FuncMd5Crypt)
		}

		sl := saltLen(s)
		if sl > 8 || sl < 1 {
			sl = 4
		}

		return md5crypt.Generate(s[len(s)-1], sl) //nolint:wrapcheck
	}
}

func sshaFunc() func(...string) (string, error) {
	// parameters: s[0] = salt, s[-1] = password
	return func(s ...string) (string, error) {
		if len(s) < 1 {
			return "", fmt.Errorf("usage: %s <salt> <password>", FuncSSHA)
		}

		return ssha.Generate(s[len(s)-1], saltLen(s)) //nolint:wrapcheck
	}
}

func ssha256Func() func(...string) (string, error) {
	// parameters: s[0] = salt, s[-1] = password
	return func(s ...string) (string, error) {
		if len(s) < 1 {
			return "", fmt.Errorf("usage: %s <salt> <password>", FuncSSHA256)
		}

		return ssha256.Generate(s[len(s)-1], saltLen(s)) //nolint:wrapcheck
	}
}

func ssha512Func() func(...string) (string, error) {
	// parameters: s[0] = salt, s[-1] = password
	return func(s ...string) (string, error) {
		if len(s) < 1 {
			return "", fmt.Errorf("usage: %s <salt> <password>", FuncSSHA512)
		}

		return ssha512.Generate(s[len(s)-1], saltLen(s)) //nolint:wrapcheck
	}
}

func argon2iFunc() func(...string) (string, error) {
	// parameters: s[0] = salt, s[-1] = password
	return func(s ...string) (string, error) {
		if len(s) < 1 {
			return "", fmt.Errorf("usage: %s <salt> <password>", FuncArgon2i)
		}

		return argon2i.Generate(s[len(s)-1], uint32(saltLen(s))) //nolint:wrapcheck
	}
}

func argon2idFunc() func(...string) (string, error) {
	// parameters: s[0] = salt, s[-1] = password
	return func(s ...string) (string, error) {
		if len(s) < 1 {
			return "", fmt.Errorf("usage: %s <salt> <password> or <password>", FuncArgon2id)
		}

		return argon2id.Generate(s[len(s)-1], uint32(saltLen(s))) //nolint:wrapcheck
	}
}

func bcryptFunc() func(...string) (string, error) {
	// parameters: s[0] = salt, s[-1] = password
	return func(s ...string) (string, error) {
		if len(s) < 1 {
			return "", fmt.Errorf("usage: %s <password>", FuncBcrypt)
		}

		return bcrypt.Generate(s[len(s)-1]) //nolint:wrapcheck
	}
}

func get(ctx context.Context, kv kvstore) func(...string) (string, error) {
	return func(s ...string) (string, error) {
		if len(s) < 1 {
			return "", nil
		}

		if kv == nil {
			return "", fmt.Errorf("KV is nil")
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
			return "", fmt.Errorf("KV is nil")
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
			return "", fmt.Errorf("KV is nil")
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

func getValues(ctx context.Context, kv kvstore) func(...string) ([]string, error) {
	return func(s ...string) ([]string, error) {
		if len(s) < 2 {
			return nil, nil
		}

		if kv == nil {
			return nil, fmt.Errorf("KV is nil")
		}

		sec, err := kv.Get(ctx, s[0])
		if err != nil {
			return nil, fmt.Errorf("failed to get %q: %w", s[0], err)
		}

		values, found := sec.Values(s[1])
		if !found {
			return nil, fmt.Errorf("key %q not found", s[1])
		}

		return values, nil
	}
}

func funcMap(ctx context.Context, kv kvstore) template.FuncMap {
	return template.FuncMap{
		FuncGet:         get(ctx, kv),
		FuncGetPassword: getPassword(ctx, kv),
		FuncGetValue:    getValue(ctx, kv),
		FuncGetValues:   getValues(ctx, kv),
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
