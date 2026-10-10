package gogit

import (
	"bytes"
	"context"
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
	return authForURL(ctx, g.remoteURL(remote))
}

// authForURL returns an auth method suitable for the given remote URL.
func authForURL(ctx context.Context, url string) transport.AuthMethod {
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
		return httpsAuth(ctx, url)
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
func httpsAuth(ctx context.Context, url string) transport.AuthMethod {
	// URL userinfo (https://user:pass@host/...) takes precedence.
	if u, p, ok := urlUserInfo(url); ok {
		if u == "" {
			return &http.TokenAuth{Token: p}
		}

		return &http.BasicAuth{Username: u, Password: p}
	}

	if a := credentialHelperAuth(ctx, url); a != nil {
		return a
	}

	debug.Log("no HTTPS credentials found for %s; go-git does not read git credential helpers", url)

	return nil
}

// urlUserInfo extracts userinfo from an http(s) URL if present.
func urlUserInfo(raw string) (string, string, bool) {
	i := strings.Index(raw, "://")
	if i < 0 {
		return "", "", false
	}
	rest := raw[i+3:]
	at := strings.Index(rest, "@")
	if at < 0 {
		return "", "", false
	}
	userinfo := rest[:at]
	slash := strings.Index(userinfo, "/")
	if slash >= 0 {
		return "", "", false
	}
	user, pass, found := strings.Cut(userinfo, ":")
	if !found {
		return user, "", true
	}

	return user, pass, true
}

// credentialHelperAuth shells out to `git credential fill` when a git binary is
// available, returning nil otherwise. The secret is never logged.
func credentialHelperAuth(ctx context.Context, url string) transport.AuthMethod {
	if _, err := exec.LookPath("git"); err != nil {
		return nil
	}

	protocol, host, path := splitHTTPURL(url)
	if host == "" {
		return nil
	}

	input := bytes.NewBufferString("protocol=" + protocol + "\nhost=" + host + "\n")
	if path != "" {
		input.WriteString("path=" + path + "\n")
	}
	input.WriteString("\n")

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "git", "credential", "fill")
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

// splitHTTPURL splits an HTTP(S) URL into protocol, host and path.
func splitHTTPURL(raw string) (string, string, string) {
	i := strings.Index(raw, "://")
	if i < 0 {
		return "", "", ""
	}
	protocol := raw[:i]
	rest := raw[i+3:]
	host := rest
	path := ""
	if slash := strings.Index(rest, "/"); slash >= 0 {
		host = rest[:slash]
		path = rest[slash+1:]
	}

	return protocol, host, path
}
