package action

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/justwatchcom/gopass/pkg/store/sub"

	"github.com/justwatchcom/gopass/pkg/store/secret"

	"github.com/justwatchcom/gopass/pkg/fsutil"

	"github.com/justwatchcom/gopass/pkg/out"

	"github.com/justwatchcom/gopass/pkg/ctxutil"

	"github.com/urfave/cli"
)

type gitCredentials struct {
	Protocol string
	Host     string
	Path     string
	Username string
	Password string
}

func (c *gitCredentials) WriteTo(w io.Writer) (int64, error) {
	var n int64
	if c.Protocol != "" {
		i, err := io.WriteString(w, "protocol="+c.Protocol+"\n")
		n += int64(i)
		if err != nil {
			return n, err
		}
	}
	if c.Host != "" {
		i, err := io.WriteString(w, "host="+c.Host+"\n")
		n += int64(i)
		if err != nil {
			return n, err
		}
	}
	if c.Path != "" {
		i, err := io.WriteString(w, "path="+c.Path+"\n")
		n += int64(i)
		if err != nil {
			return n, err
		}
	}
	if c.Username != "" {
		i, err := io.WriteString(w, "username="+c.Username+"\n")
		n += int64(i)
		if err != nil {
			return n, err
		}
	}
	if c.Password != "" {
		i, err := io.WriteString(w, "password="+c.Password+"\n")
		n += int64(i)
		if err != nil {
			return n, err
		}
	}
	return n, nil
}

func parseGitCredentials(r io.Reader) (*gitCredentials, error) {
	rd := bufio.NewReader(r)
	c := &gitCredentials{}
	for {
		key, err := rd.ReadString('=')
		if err == io.EOF {
			return c, nil
		} else if err != nil {
			return nil, err
		}
		key = strings.TrimSuffix(key, "=")
		val, err := rd.ReadString('\n')
		if err == io.EOF {
			err = io.ErrUnexpectedEOF
		}
		if err != nil {
			return nil, err
		}
		val = strings.TrimSuffix(val, "\n")
		switch key {
		case "protocol":
			c.Protocol = val
		case "host":
			c.Host = val
		case "path":
			c.Path = val
		case "username":
			c.Username = val
		case "password":
			c.Password = val
		}
	}
}

// GitCredentialBefore is executed before another git-credential command
func (s *Action) GitCredentialBefore(ctx context.Context, c *cli.Context) error {
	if !ctxutil.IsStdin(ctx) {
		return ExitError(ctx, ExitUsage, nil, "missing stdin from git")
	}
	return s.Initialized(ctx, c)
}

// GitCredentialGet returns a credential to git
func (s *Action) GitCredentialGet(ctx context.Context, c *cli.Context) error {
	cred, err := parseGitCredentials(os.Stdin)
	if err != nil {
		return ExitError(ctx, ExitUnsupported, err, "Error: %v while parsing git-credential", err)
	}
	path := "git/" + fsutil.CleanFilename(cred.Host) + "/" + fsutil.CleanFilename(cred.Username)
	if !s.Store.Exists(ctx, path) {
		if s.Store.IsDir(ctx, path) {
			tree, err := s.Store.Tree(ctx)
			if err != nil {
				return ExitError(ctx, ExitDecrypt, err, "Error: %v while listing the storage", err)
			}
			sub, err := tree.FindFolder(path)
			if err != nil {
				return ExitError(ctx, ExitDecrypt, err, "Error: %v while listing the storage", err)
			}
			entries := sub.List(0)
			if len(entries) == 1 {
				path = "git/" + entries[0]
			} else {
				fmt.Fprintln(os.Stderr, "gopass error: too many entries")
			}
		} else {
			return nil
		}
	}
	secret, err := s.Store.Get(ctx, path)
	if err != nil {
		return ExitError(ctx, ExitDecrypt, err, "")
	}
	cred.Password = secret.Password()
	cred.Username, err = secret.Value("login")
	if err != nil {
		log.Println(err)
	}

	_, err = cred.WriteTo(out.Stdout)
	if err != nil {
		return ExitError(ctx, ExitIO, err, "Could not write to stdout")
	}
	return nil
}

// GitCredentialStore stores a credential got from git
func (s *Action) GitCredentialStore(ctx context.Context, c *cli.Context) error {
	cred, err := parseGitCredentials(os.Stdin)
	if err != nil {
		return ExitError(ctx, ExitUnsupported, err, "Error: %v while parsing git-credential", err)
	}
	path := "git/" + fsutil.CleanFilename(cred.Host) + "/" + fsutil.CleanFilename(cred.Username)
	secret := secret.New(cred.Password, "")
	if cred.Username != "" {
		_ = secret.SetValue("login", cred.Username)
	}
	err = s.Store.Set(sub.WithAutoSync(ctx, false), path, secret)
	if err != nil {
		fmt.Fprintln(os.Stderr, "gopass error: error while writing to store")
	}
	return nil
}

// GitCredentialErase removes a credential got from git
func (s *Action) GitCredentialErase(ctx context.Context, c *cli.Context) error {
	cred, err := parseGitCredentials(os.Stdin)
	if err != nil {
		return ExitError(ctx, ExitUnsupported, err, "Error: %v while parsing git-credential", err)
	}
	path := "git/" + fsutil.CleanFilename(cred.Host) + "/" + fsutil.CleanFilename(cred.Username)
	err = s.Store.Delete(sub.WithAutoSync(ctx, false), path)
	if err != nil {
		fmt.Fprintln(os.Stderr, "gopass error: error while writing to store")
	}
	return nil
}
