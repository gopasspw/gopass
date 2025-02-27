package tpl

import (
	"context"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/gopasspw/gopass/internal/hashsum"
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
	FuncMd5sum        = "md5sum"
	FuncSha1sum       = "sha1sum"
	FuncSha256sum     = "sha256sum"
	FuncSha512sum     = "sha512sum"
	FuncBlake3        = "blake3"
	FuncMd5Crypt      = "md5crypt"
	FuncSSHA          = "ssha"
	FuncSSHA256       = "ssha256"
	FuncSSHA512       = "ssha512"
	FuncGet           = "get"
	FuncGetPassword   = "getpw"
	FuncGetValue      = "getval"
	FuncGetValues     = "getvals"
	FuncArgon2i       = "argon2i"
	FuncArgon2id      = "argon2id"
	FuncBcrypt        = "bcrypt"
	FuncJoin          = "join"
	FuncRoundDuration = "roundDuration"
	FuncDate          = "date"
	FuncTruncate      = "truncate"
)

func md5sum() func(...string) (string, error) {
	return func(s ...string) (string, error) {
		return hashsum.MD5Hex(s[0]), nil
	}
}

func sha1sum() func(...string) (string, error) {
	return func(s ...string) (string, error) {
		return hashsum.SHA1Hex(s[0]), nil
	}
}

func sha256sum() func(...string) (string, error) {
	return func(s ...string) (string, error) {
		return hashsum.SHA256Hex(s[0]), nil
	}
}

func sha512sum() func(...string) (string, error) {
	return func(s ...string) (string, error) {
		return hashsum.SHA512Hex(s[0]), nil
	}
}

func blake3sum() func(...string) (string, error) {
	return func(s ...string) (string, error) {
		return hashsum.Blake3Hex(s[0]), nil
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

func roundDuration(duration any) string {
	var d time.Duration
	switch duration := duration.(type) {
	case string:
		d, _ = time.ParseDuration(duration)
	case int64:
		d = time.Duration(duration)
	case time.Time:
		d = time.Since(duration)
	case time.Duration:
		d = duration
	default:
		d = 0
	}

	u := uint64(d)
	year := uint64(time.Hour) * 24 * 365
	month := uint64(time.Hour) * 24 * 30
	day := uint64(time.Hour) * 24
	hour := uint64(time.Hour)
	minute := uint64(time.Minute)
	second := uint64(time.Second)

	switch {
	case u >= year:
		return strconv.FormatUint(u/year, 10) + "y"
	case u >= month:
		return strconv.FormatUint(u/month, 10) + "mo"
	case u >= day:
		return strconv.FormatUint(u/day, 10) + "d"
	case u >= hour:
		return strconv.FormatUint(u/hour, 10) + "h"
	case u >= minute:
		return strconv.FormatUint(u/minute, 10) + "m"
	case u >= second:
		return strconv.FormatUint(u/second, 10) + "s"
	default:
		return "0s"
	}
}

func date(ts time.Time) string {
	return ts.Format("2006-01-02")
}

func truncate(length int, v any) string {
	sv := strval(v)
	if len(sv) < length {
		return sv
	}

	return sv[:length-3] + "..."
}

func join(sep string, v any) string {
	return strings.Join(stringslice(v), sep)
}

func stringslice(v any) []string {
	switch v := v.(type) {
	case []string:
		return v
	case []interface{}:
		res := make([]string, 0, len(v))
		for _, s := range v {
			if s == nil {
				continue
			}
			res = append(res, strval(s))
		}

		return res
	default:
		val := reflect.ValueOf(v)
		switch val.Kind() { //nolint:exhaustive
		case reflect.Array, reflect.Slice:
			l := val.Len()
			res := make([]string, 0, l)
			for i := range l {
				value := val.Index(i).Interface()
				if value == nil {
					continue
				}
				res = append(res, strval(value))
			}

			return res
		default:
			if v == nil {
				return []string{}
			}

			return []string{strval(v)}
		}
	}
}

func strval(v any) string {
	switch v := v.(type) {
	case string:
		return v
	case []byte:
		return string(v)
	case error:
		return v.Error()
	case fmt.Stringer:
		return v.String()
	default:
		return fmt.Sprintf("%v", v)
	}
}

func funcMap(ctx context.Context, kv kvstore) template.FuncMap {
	return template.FuncMap{
		FuncGet:           get(ctx, kv),
		FuncGetPassword:   getPassword(ctx, kv),
		FuncGetValue:      getValue(ctx, kv),
		FuncGetValues:     getValues(ctx, kv),
		FuncMd5sum:        md5sum(),
		FuncSha1sum:       sha1sum(),
		FuncSha256sum:     sha256sum(),
		FuncSha512sum:     sha512sum(),
		FuncBlake3:        blake3sum(),
		FuncMd5Crypt:      md5cryptFunc(),
		FuncSSHA:          sshaFunc(),
		FuncSSHA256:       ssha256Func(),
		FuncSSHA512:       ssha512Func(),
		FuncArgon2i:       argon2iFunc(),
		FuncArgon2id:      argon2idFunc(),
		FuncBcrypt:        bcryptFunc(),
		FuncJoin:          join,
		FuncRoundDuration: roundDuration,
		FuncDate:          date,
		FuncTruncate:      truncate,
	}
}

// PublicFuncMap returns a template.FuncMap with useful template functions.
func PublicFuncMap() template.FuncMap {
	return template.FuncMap{
		FuncMd5sum:        md5sum(),
		FuncSha1sum:       sha1sum(),
		FuncSha256sum:     sha256sum(),
		FuncSha512sum:     sha512sum(),
		FuncBlake3:        blake3sum(),
		FuncMd5Crypt:      md5cryptFunc(),
		FuncSSHA:          sshaFunc(),
		FuncSSHA256:       ssha256Func(),
		FuncSSHA512:       ssha512Func(),
		FuncArgon2i:       argon2iFunc(),
		FuncArgon2id:      argon2idFunc(),
		FuncBcrypt:        bcryptFunc(),
		FuncJoin:          join,
		FuncRoundDuration: roundDuration,
		FuncDate:          date,
		FuncTruncate:      truncate,
	}
}
