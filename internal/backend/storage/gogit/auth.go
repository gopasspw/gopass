package gogit

import (
	"bytes"
	"context"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"github.com/gopasspw/gopass/pkg/debug"
)

// This file is the single seam for all transport and authentication concerns.
// Every call to CloneOptions.Auth, FetchOptions.Auth, PushOptions.Auth and
// PullOptions.Auth must go through g.auth (or authForURL). When go-git is later
// bumped to /v6 — which rewrites the transport layer — this is the only file
// that needs to be rewritten.

// auth returns a transport.AuthMethod for the named remote, or nil when no
// credentials are required (e.g. local file paths).
func (g *GoGit) auth(ctx context.Context, remote string) transport.AuthMethod {
	return authForURL(ctx, g.remoteURL(remote), g.fs.Path())
}

// authForURL returns an auth method suitable for the given remote URL. dir is
// the directory `git credential fill` runs in, so credential helpers resolve
// against the store's config rather than the process working directory.
func authForURL(ctx context.Context, url, dir string) transport.AuthMethod {
	switch {
	case isSSHURL(url):
		user := userFromURL(url)
		if a, err := ssh.NewSSHAgentAuth(user); err == nil {
			return a
		}
		if key := sshKeyPath(); key != "" {
			if a, err := ssh.NewPublicKeysFromFile(user, key, ""); err == nil {
				return a
			}
		}

		return nil
	case strings.HasPrefix(url, "https://"), strings.HasPrefix(url, "http://"):
		return httpsAuth(ctx, url, dir)
	default:
		// local file paths need no auth.
		return nil
	}
}

// isSSHURL reports whether url describes an SSH remote.
func isSSHURL(url string) bool {
	return strings.HasPrefix(url, "ssh://") ||
		strings.HasPrefix(url, "git@") ||
		(strings.Contains(url, "@") && strings.Contains(url, ":") && !strings.Contains(url, "://"))
}

// userFromURL extracts the SSH user from a remote URL, defaulting to "git".
func userFromURL(url string) string {
	u := url
	if i := strings.Index(u, "://"); i >= 0 {
		u = u[i+3:]
	}
	if i := strings.Index(u, "@"); i >= 0 {
		return u[:i]
	}

	return "git"
}

// sshKeyPath returns a candidate SSH private key path from the environment or
// the default ~/.ssh locations, or "" if none is found.
func sshKeyPath() string {
	if k := os.Getenv("GIT_SSH_KEY"); k != "" {
		return k
	}
	if c := os.Getenv("GIT_SSH_COMMAND"); c != "" {
		if k := keyFromSSHCommand(c); k != "" {
			return k
		}
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	for _, name := range []string{"id_ed25519", "id_rsa", "id_ecdsa"} {
		p := filepath.Join(home, ".ssh", name)
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}

	return ""
}

// keyFromSSHCommand extracts an identity file (-i) from a GIT_SSH_COMMAND value.
func keyFromSSHCommand(cmd string) string {
	fields := strings.Fields(cmd)
	for i, f := range fields {
		if f == "-i" && i+1 < len(fields) {
			return fields[i+1]
		}
		if strings.HasPrefix(f, "-i") && len(f) > 2 {
			return f[2:]
		}
	}

	return ""
}

// httpsAuth resolves HTTP(S) credentials. go-git ignores git's credential
// helpers, so when a git binary is available we shell out to
// `git credential fill`. Otherwise we fall back to URL userinfo and finally to
// GIT_ASKPASS/.netrc being honoured by the caller providing explicit config.
func httpsAuth(ctx context.Context, url, dir string) transport.AuthMethod {
	// URL userinfo (https://user:pass@host/...) takes precedence, but only
	// when it actually carries a password. A username-only URL (e.g. Azure
	// DevOps https://org@dev.azure.com/...) must still consult the credential
	// helper.
	if u, p, ok := urlUserInfo(url); ok && p != "" {
		return &http.BasicAuth{Username: u, Password: p}
	}

	if a := credentialHelperAuth(ctx, url, dir); a != nil {
		return a
	}

	debug.Log("no HTTPS credentials found for %s; go-git does not read git credential helpers", url)

	return nil
}

// urlUserInfo extracts userinfo from an http(s) URL if present. The user and
// password are percent-decoded. ok is true only when a password is present, so
// a username-only URL does not short-circuit the credential helper.
func urlUserInfo(raw string) (string, string, bool) {
	u, err := url.Parse(raw)
	if err != nil || u.User == nil {
		return "", "", false
	}

	user := u.User.Username()
	pass, hasPass := u.User.Password()
	if !hasPass {
		return user, "", false
	}

	return user, pass, true
}

// credentialHelperAuth shells out to `git credential fill` when a git binary is
// available, returning nil otherwise. The secret is never logged. dir is the
// directory the command runs in so credential helpers resolve against the
// store's config.
func credentialHelperAuth(ctx context.Context, url, dir string) transport.AuthMethod {
	if _, err := exec.LookPath("git"); err != nil {
		return nil
	}

	protocol, host, path, username := splitHTTPURL(url)
	if host == "" {
		return nil
	}

	input := bytes.NewBufferString("protocol=" + protocol + "\nhost=" + host + "\n")
	if path != "" {
		input.WriteString("path=" + path + "\n")
	}
	if username != "" {
		input.WriteString("username=" + username + "\n")
	}
	input.WriteString("\n")

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "git", "credential", "fill")
	if dir != "" {
		cmd.Dir = dir
	}
	// Never let git prompt on the TTY: with no helper configured it would block
	// until the timeout and then be killed.
	cmd.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0")
	cmd.Stdin = input
	out, err := cmd.Output()
	if err != nil {
		debug.Log("git credential fill failed: %s", err)

		return nil
	}

	var user, pass string
	for line := range strings.SplitSeq(string(out), "\n") {
		switch {
		case strings.HasPrefix(line, "username="):
			user = strings.TrimPrefix(line, "username=")
		case strings.HasPrefix(line, "password="):
			pass = strings.TrimPrefix(line, "password=")
		}
	}

	if user == "" && pass == "" {
		return nil
	}
	if user == "" {
		return &http.TokenAuth{Token: pass}
	}

	return &http.BasicAuth{Username: user, Password: pass}
}

// splitHTTPURL splits an HTTP(S) URL into protocol, host and path. The host is
// returned without any userinfo, and a username present in the URL is returned
// separately so credential helpers keyed on the host (e.g. dev.azure.com) still
// match a username-only URL such as https://org@dev.azure.com/org/proj/_git/repo.
func splitHTTPURL(raw string) (string, string, string, string) {
	u, err := url.Parse(raw)
	if err != nil {
		return "", "", "", ""
	}

	username := ""
	if u.User != nil {
		username = u.User.Username()
	}

	return u.Scheme, u.Host, strings.TrimPrefix(u.Path, "/"), username
}

// stripUserInfo removes any userinfo from a URL so it can be handed to go-git
// without the credentials ending up in go-git's error messages. Credentials are
// passed via Auth instead.
func stripUserInfo(raw string) string {
	u, err := url.Parse(raw)
	if err != nil || u.User == nil {
		return raw
	}

	u.User = nil

	return u.String()
}

// redactURL removes any password from a URL so it can be safely logged or
// embedded in an error message.
func redactURL(raw string) string {
	u, err := url.Parse(raw)
	if err != nil || u.User == nil {
		return raw
	}
	if _, hasPass := u.User.Password(); !hasPass {
		return raw
	}

	u.User = url.UserPassword(u.User.Username(), "REDACTED")

	return u.String()
}
