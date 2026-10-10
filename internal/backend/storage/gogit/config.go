package gogit

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	git "github.com/go-git/go-git/v5"
	formatconfig "github.com/go-git/go-git/v5/plumbing/format/config"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/gopasspw/gopass/internal/backend"
	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/internal/store"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/pkg/debug"
)

const (
	fileMode = 0o600
)

// fixConfig sets up the git config for the password store in a way that
// simplifies some of the quirks of git. Unlike the gitfs backend this does NOT
// set core.sshCommand: go-git does not read it and setting it would only
// affect a later switch back to the CLI backend.
func (g *GoGit) fixConfig(ctx context.Context) error {
	defaults := []struct{ key, value string }{
		{"push.default", "matching"},
		{"pull.rebase", "false"},
		{"diff.gpg.binary", "true"},
		{"diff.gpg.textconv", "gpg --no-tty --decrypt"},
	}

	for _, d := range defaults {
		if err := g.ConfigSet(ctx, d.key, d.value); err != nil {
			debug.Log("failed to set git config %q: %s", d.key, err)
		}
	}

	return nil
}

// InitConfig initializes and prepares the git config.
func (g *GoGit) InitConfig(ctx context.Context, userName, userEmail string) error {
	if userName != "" {
		if err := g.ConfigSet(ctx, "user.name", userName); err != nil {
			return fmt.Errorf("failed to set git config user.name: %w", err)
		}
	} else {
		out.Printf(ctx, "Git Username not set")
	}

	if userEmail != "" && strings.Contains(userEmail, "@") {
		if err := g.ConfigSet(ctx, "user.email", userEmail); err != nil {
			return fmt.Errorf("failed to set git config user.email: %w", err)
		}
	} else {
		out.Printf(ctx, "Git Email not set")
	}

	if err := g.fixConfig(ctx); err != nil {
		return fmt.Errorf("failed to fix git config: %w", err)
	}

	gitattributes, commitMsg := gitattributesForBackend(ctx)
	if err := os.WriteFile(filepath.Join(g.fs.Path(), ".gitattributes"), []byte(gitattributes), fileMode); err != nil {
		return fmt.Errorf("failed to initialize git: %w", err)
	}
	if err := g.Add(ctx, filepath.Join(g.fs.Path(), ".gitattributes")); err != nil {
		out.Warningf(ctx, "Failed to add .gitattributes to git")
	}
	if err := g.Commit(ctx, commitMsg); err != nil {
		out.Warningf(ctx, "Failed to commit .gitattributes to git")
	}

	return nil
}

// ConfigSet sets a local config value.
//
// go-git's typed config does not expose a generic string setter, so the value
// is written into the raw config and the structured fields that the marshaler
// reads back (user.name/user.email) are kept in sync explicitly.
func (g *GoGit) ConfigSet(_ context.Context, key, value string) error {
	cfg, err := g.repo.Config()
	if err != nil {
		return err
	}
	if cfg.Raw == nil {
		cfg.Raw = formatconfig.New()
	}

	sec, sub, opt := splitKey(key)
	if sub == "" {
		cfg.Raw.Section(sec).SetOption(opt, value)
	} else {
		cfg.Raw.Section(sec).Subsection(sub).SetOption(opt, value)
	}

	switch key {
	case "user.name":
		cfg.User.Name = value
	case "user.email":
		cfg.User.Email = value
	}

	return g.repo.Storer.SetConfig(cfg)
}

// ConfigGet returns a given config value.
func (g *GoGit) ConfigGet(_ context.Context, key string) (string, error) {
	if !g.IsInitialized() {
		return "", store.ErrGitNotInit
	}

	cfg, err := g.repo.Config()
	if err != nil {
		return "", err
	}
	if cfg.Raw == nil {
		return "", nil
	}

	sec, sub, opt := splitKey(key)
	if sub != "" {
		return cfg.Raw.Section(sec).Subsection(sub).Option(opt), nil
	}

	return cfg.Raw.Section(sec).Option(opt), nil
}

// ConfigList returns all git config settings.
func (g *GoGit) ConfigList(_ context.Context) (map[string]string, error) {
	if !g.IsInitialized() {
		return nil, store.ErrGitNotInit
	}

	cfg, err := g.repo.Config()
	if err != nil {
		return nil, err
	}

	return flattenRawConfig(cfg.Raw), nil
}

// commitSignature returns the author/committer signature for a commit, honoring
// the configured user and an optional fixed commit timestamp.
func (g *GoGit) commitSignature(ctx context.Context) *object.Signature {
	name := ctxutil.GetUsername(ctx)
	email := ctxutil.GetEmail(ctx)

	if name == "" || email == "" {
		cfgName, cfgEmail := g.configUser()
		if name == "" {
			name = cfgName
		}
		if email == "" {
			email = cfgEmail
		}
	}

	if name == "" {
		name = "gopass"
	}
	if email == "" {
		email = "gopass@localhost"
	}

	when := time.Now()
	if ctxutil.HasCommitTimestamp(ctx) {
		when = ctxutil.GetCommitTimestamp(ctx)
	}

	return &object.Signature{Name: name, Email: email, When: when}
}

// configUser returns the user.name and user.email from the local git config.
func (g *GoGit) configUser() (string, string) {
	cfg, err := g.repo.Config()
	if err != nil {
		return "", ""
	}

	return cfg.User.Name, cfg.User.Email
}

// remoteURL returns the URL of the named remote, or "" if it does not exist.
func (g *GoGit) remoteURL(remote string) string {
	if remote == "" {
		remote = git.DefaultRemoteName
	}
	r, err := g.repo.Remote(remote)
	if err != nil {
		return ""
	}
	urls := r.Config().URLs
	if len(urls) == 0 {
		return ""
	}

	return urls[0]
}

// defaultRemote returns the default remote for the given branch, falling back
// to origin.
func (g *GoGit) defaultRemote(branch string) string {
	cfg, err := g.repo.Config()
	if err != nil {
		return git.DefaultRemoteName
	}
	if b, ok := cfg.Branches[branch]; ok && b.Remote != "" {
		if _, err := g.repo.Remote(b.Remote); err == nil {
			return b.Remote
		}
	}

	return git.DefaultRemoteName
}

// defaultBranch returns the current branch name, defaulting to master when HEAD
// is unborn.
func (g *GoGit) defaultBranch() string {
	head, err := g.repo.Head()
	if err != nil {
		return "master"
	}
	if !head.Name().IsBranch() {
		return "master"
	}

	return head.Name().Short()
}

// resolveRemoteBranch fills in the default remote and branch if they are empty.
func (g *GoGit) resolveRemoteBranch(remote, branch string) (string, string) {
	if branch == "" {
		branch = g.defaultBranch()
	}
	if remote == "" {
		remote = g.defaultRemote(branch)
	}

	return remote, branch
}

// gitattributesForBackend returns the .gitattributes content and commit message
// appropriate for the active crypto backend.
func gitattributesForBackend(ctx context.Context) (string, string) {
	if backend.GetCryptoBackend(ctx) == backend.Age {
		return "*.age binary\n", "Configure git repository for age file handling."
	}

	return "*.gpg diff=gpg\n", "Configure git repository for gpg file diff."
}

// splitKey splits a git config key like "branch.master.remote" into its
// section, subsection and name parts. For a two-part key ("user.name") the
// subsection is empty. Only the dotted form is supported here, which is all the
// gopass backend uses.
func splitKey(key string) (string, string, string) {
	parts := strings.Split(key, ".")
	switch len(parts) {
	case 0:
		return "", "", ""
	case 1:
		return parts[0], "", ""
	case 2:
		return parts[0], "", parts[1]
	default:
		return parts[0], strings.Join(parts[1:len(parts)-1], "."), parts[len(parts)-1]
	}
}

// flattenRawConfig flattens a raw git config into a dotted key/value map.
func flattenRawConfig(raw *formatconfig.Config) map[string]string {
	kv := make(map[string]string, 23)
	if raw == nil {
		return kv
	}
	for _, s := range raw.Sections {
		for _, sub := range s.Subsections {
			for _, opt := range sub.Options {
				kv[s.Name+"."+sub.Name+"."+opt.Key] = opt.Value
			}
		}
		for _, opt := range s.Options {
			kv[s.Name+"."+opt.Key] = opt.Value
		}
	}

	return kv
}
