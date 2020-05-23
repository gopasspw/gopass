package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/internal/store/secret"
	"github.com/gopasspw/gopass/internal/store/sub"
	"github.com/gopasspw/gopass/internal/termio"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/pkg/fsutil"
	"github.com/gopasspw/gopass/pkg/gopass"
	"github.com/urfave/cli/v2"
)

type gitCredentials struct {
	Protocol string
	Host     string
	Path     string
	Username string
	Password string
}

// WriteTo writes the given credentials to the given io.Writer in the git-credential format
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
		if err != nil {
			if err == io.EOF {
				if key == "" {
					return c, nil
				}
				return nil, io.ErrUnexpectedEOF
			}
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

type gc struct {
	gp gopass.Store
}

// Before is executed before another git-credential command
func (s *gc) Before(c *cli.Context) error {
	ctx := ctxutil.WithGlobalFlags(c)
	ctx = ctxutil.WithInteractive(ctx, false)
	if !ctxutil.IsStdin(ctx) {
		return fmt.Errorf("missing stdin from git")
	}
	return nil
}

func filter(ls []string, prefix string) []string {
	out := make([]string, 0, len(ls))
	for _, e := range ls {
		if !strings.HasPrefix(e, prefix) {
			continue
		}
		out = append(out, e)
	}
	return out
}

// Get returns a credential to git
func (s *gc) Get(c *cli.Context) error {
	ctx := ctxutil.WithGlobalFlags(c)
	ctx = sub.WithAutoSync(ctx, false)
	cred, err := parseGitCredentials(termio.Stdin)
	if err != nil {
		return fmt.Errorf("error: %v while parsing git-credential", err)
	}
	// try git/host/username... If username is empty, simply try git/host
	path := "git/" + fsutil.CleanFilename(cred.Host) + "/" + fsutil.CleanFilename(cred.Username)
	if _, err := s.gp.Get(ctx, path); err != nil {
		// if the looked up path is a directory with only one entry (e.g. one user per host), take the subentry instead
		ls, err := s.gp.List(ctx)
		if err != nil {
			return fmt.Errorf("error: %v while listing the storage", err)
		}
		entries := filter(ls, path)
		if len(entries) < 1 {
			// no entry found, this is not an error
			return nil
		}
		if len(entries) > 1 {
			fmt.Fprintln(os.Stderr, "gopass error: too many entries")
			return nil
		}
		path = entries[0]
	}
	secret, err := s.gp.Get(ctx, path)
	if err != nil {
		return err
	}
	cred.Password = secret.Password()
	username, err := secret.Value("login")
	if err == nil {
		// leave the username as is otherwise
		cred.Username = username
	}

	_, err = cred.WriteTo(out.Stdout)
	if err != nil {
		return fmt.Errorf("could not write to stdout: %s", err)
	}
	return nil
}

// Store stores a credential got from git
func (s *gc) Store(c *cli.Context) error {
	ctx := ctxutil.WithGlobalFlags(c)
	cred, err := parseGitCredentials(termio.Stdin)
	if err != nil {
		return fmt.Errorf("error: %v while parsing git-credential", err)
	}
	path := "git/" + fsutil.CleanFilename(cred.Host) + "/" + fsutil.CleanFilename(cred.Username)
	// This should never really be an issue because git automatically removes invalid credentials first
	if _, err := s.gp.Get(ctx, path); err == nil {
		out.Debug(ctx, ""+
			"gopass: did not store \"%s\" because it already exists. "+
			"If you want to overwrite it, delete it first by doing: "+
			"\"gopass rm %s\"\n",
			path, path,
		)
		return nil
	}
	secret := secret.New(cred.Password, "")
	if cred.Username != "" {
		_ = secret.SetValue("login", cred.Username)
	}

	if err := s.gp.Set(ctx, path, secret); err != nil {
		fmt.Fprintf(os.Stderr, "gopass error: error while writing to store: %v\n", err)
	}
	return nil
}

// Erase removes a credential got from git
func (s *gc) Erase(c *cli.Context) error {
	ctx := ctxutil.WithGlobalFlags(c)
	cred, err := parseGitCredentials(termio.Stdin)
	if err != nil {
		return fmt.Errorf("error: %v while parsing git-credential", err)
	}

	path := "git/" + fsutil.CleanFilename(cred.Host) + "/" + fsutil.CleanFilename(cred.Username)
	if err := s.gp.Remove(ctx, path); err != nil {
		fmt.Fprintln(os.Stderr, "gopass error: error while writing to store")
	}
	return nil
}

// Configure configures gopass as git's credential.helper
func (s *gc) Configure(c *cli.Context) error {
	ctx := ctxutil.WithGlobalFlags(c)
	flags := 0
	flag := "--global"
	if c.Bool("local") {
		flag = "--local"
		flags++
	}
	if c.Bool("global") {
		flag = "--global"
		flags++
	}
	if c.Bool("system") {
		flag = "--system"
		flags++
	}
	if flags >= 2 {
		return fmt.Errorf("only specify one target of installation")
	}
	if flags == 0 {
		log.Println("No target given, assuming --global.")
	}
	cmd := exec.CommandContext(ctx, "git", "config", flag, "credential.helper", `"!gopass-git-credentials $@"`)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
