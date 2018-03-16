package tpl

import (
	"context"
	"crypto/md5"
	"crypto/sha1"
	"fmt"
	"text/template"

	"github.com/pkg/errors"
)

// These constants defined the template function names used
const (
	FuncMd5sum      = "md5sum"
	FuncSha1sum     = "sha1sum"
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
		buf, err := sec.Bytes()
		if err != nil {
			return err.Error(), nil
		}
		return string(buf), nil
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
		val, err := sec.Value(s[1])
		if err != nil {
			return err.Error(), nil
		}
		return val, nil
	}
}

func funcMap(ctx context.Context, kv kvstore) template.FuncMap {
	return template.FuncMap{
		FuncGet:         get(ctx, kv),
		FuncGetPassword: getPassword(ctx, kv),
		FuncGetValue:    getValue(ctx, kv),
		FuncMd5sum:      md5sum(),
		FuncSha1sum:     sha1sum(),
	}
}
