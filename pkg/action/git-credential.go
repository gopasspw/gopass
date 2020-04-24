package action

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/pkg/fsutil"
	"github.com/gopasspw/gopass/pkg/out"
	"github.com/gopasspw/gopass/pkg/store/secret"
	"github.com/gopasspw/gopass/pkg/store/sub"
	"github.com/gopasspw/gopass/pkg/termio"
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

// GitCredentialBefore is executed before another git-credential command
func (s *Action) GitCredentialBefore(ctx context.Context, c *cli.Context) error {
	err := s.Initialized(ctx, c)
	if err != nil {
		return err
	}
	if !ctxutil.IsStdin(ctx) {
		return ExitError(ctx, ExitUsage, nil, "missing stdin from git")
	}
	return nil
}

// GitCredentialGet returns a credential to git
func (s *Action) GitCredentialGet(ctx context.Context, c *cli.Context) error {
	ctx = sub.WithAutoSync(ctx, false)
	cred, err := parseGitCredentials(termio.Stdin)
	if err != nil {
		return ExitError(ctx, ExitUnsupported, err, "Error: %v while parsing git-credential", err)
	}
	// try git/host/username... If username is empty, simply try git/host
	path := "git/" + fsutil.CleanFilename(cred.Host) + "/" + fsutil.CleanFilename(cred.Username)
	if !s.Store.Exists(ctx, path) {
		// if the looked up path is a directory with only one entry (e.g. one user per host), take the subentry instead
		if !s.Store.IsDir(ctx, path) {
			return nil
		}
		tree, err := s.Store.Tree(ctx)
		if err != nil {
			return ExitError(ctx, ExitDecrypt, err, "Error: %v while listing the storage", err)
		}
		sub, err := tree.FindFolder(path)
		if err != nil {
			// no entry found... this is not an error
			return nil
		}
		entries := sub.List(0)
		if len(entries) != 1 {
			fmt.Fprintln(os.Stderr, "gopass error: too many entries")
			return nil
		}
		path = "git/" + entries[0]
	}
	secret, err := s.Store.Get(ctx, path)
	if err != nil {
		return ExitError(ctx, ExitDecrypt, err, "")
	}
	cred.Password = secret.Password()
	username, err := secret.Value("login")
	if err == nil {
		// leave the username as is otherwise
		cred.Username = username
	}

	_, err = cred.WriteTo(out.Stdout)
	if err != nil {
		return ExitError(ctx, ExitIO, err, "Could not write to stdout")
	}
	return nil
}

// GitCredentialStore stores a credential got from git
func (s *Action) GitCredentialStore(ctx context.Context, c *cli.Context) error {
	cred, err := parseGitCredentials(termio.Stdin)
	if err != nil {
		return ExitError(ctx, ExitUnsupported, err, "Error: %v while parsing git-credential", err)
	}
	path := "git/" + fsutil.CleanFilename(cred.Host) + "/" + fsutil.CleanFilename(cred.Username)
	// This should never really be an issue because git automatically removes invalid credentials first
	if s.Store.Exists(ctx, path) {
		fmt.Fprintf(os.Stderr, ""+
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
	err = s.Store.Set(ctx, path, secret)
	if err != nil {
		fmt.Fprintf(os.Stderr, "gopass error: error while writing to store: %v\n", err)
	}
	return nil
}

// GitCredentialErase removes a credential got from git
func (s *Action) GitCredentialErase(ctx context.Context, c *cli.Context) error {
	cred, err := parseGitCredentials(termio.Stdin)
	if err != nil {
		return ExitError(ctx, ExitUnsupported, err, "Error: %v while parsing git-credential", err)
	}
	path := "git/" + fsutil.CleanFilename(cred.Host) + "/" + fsutil.CleanFilename(cred.Username)
	err = s.Store.Delete(ctx, path)
	if err != nil {
		fmt.Fprintln(os.Stderr, "gopass error: error while writing to store")
	}
	return nil
}

// GitCredentialConfigure configures gopass as git's credential.helper
func (s *Action) GitCredentialConfigure(ctx context.Context, c *cli.Context) error {
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
		return ExitError(ctx, ExitUnsupported, nil, "Error: only specify one target of installation!")
	}
	if flags == 0 {
		log.Println("No target given, assuming --global.")
	}
	cmd := exec.Command("git", "config", flag, "credential.helper", `"!gopass git-credential $@"`)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
