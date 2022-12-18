package gitconfig

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/exp/maps"
)

// https://mirrors.edge.kernel.org/pub/software/scm/git/docs/git-config.html#EXAMPLES
var configSampleDocs = `#
# This is the config file, and
# a '#' or ';' character indicates
# a comment
#

; core variables
[core]
	; Don't trust file modes
	filemode = false

; Our diff algorithm
[diff]
	external = /usr/local/bin/diff-wrapper
	renames = true

; Proxy settings
[core]
	gitproxy = default-proxy ; default proxy

; HTTP
[http]
    sslVerify

[http "https://weak.example.com"]
	sslVerify = false
	cookieFile = /tmp/cookie.txt
`

var configSampleComplex = `
[alias]
   
   # add
   a = add                                   # add
   aa = add --all
   all = add -A
   chunkyadd = add --patch    # stage commits chunk by chunk

   # branch
   b  = branch -v                            # branch (verbose)
   branches = branch -a
   recent = branch --sort=-committerdate

   # commit
   c = commit -m
   ca = commit -am
   ci = commit
   credit = "!f() { git commit --amend --author \"$1 <$2>\" -C HEAD; }; f" # Credit an author on the latest commit
   credit = commit --amend --author "$1 <$2>" -C HEAD
   amend = commit --amend
   commend = commit --amend --no-edit

   # checkout
   co = checkout                             # checkout
   nb = checkout -b                          # create and switch to a new branch (mnemonic: "git new branch branchname...")
   go = checkout -B                          # Switch to a branch, creating it if necessary

   # clone
   cr = clone --recursive                     # Clone a repository including all submodules

   # cherry-pick
   cp = cherry-pick -x               # grab a change from a branch

   # diff
   d = !"git diff-index --quiet HEAD -- || clear; git diff --patch-with-stat" # Show the diff between the latest commit and the current state
   di = diff
   dc = diff --cached
   div = divergence                        # Divergence (commits we added and commits remote added)
   ds = diff --stat=160,120
   gn = goodness                             # Goodness (summary of diff lines added/removed/total)
   gnc   = goodness --cached
   last = diff HEAD^

   # log
   l = log --pretty=oneline -n 20 --graph    # View the SHA, description, and history graph of the latest 20 commits
   changes = log --pretty=format:\"%h %cr %cn %Cgreen%s%Creset\" --name-status
   short = log --pretty=format:\"%h %cr %cn %Cgreen%s%Creset\"
   changelog = log --pretty=format:\" * %s\"
   shortnocolor = log --pretty=format:\"%h %cr %cn %s\"
   show-graph = log --graph --abbrev-commit --pretty=oneline

   # pull
   please = push --force-with-lease
   pl = pull
   fa = fetch --all
   ff = merge --ff-only
   noff  = merge --no-ff
   pullff   = pull --ff-only
   p = !"git pull; git submodule foreach git pull origin master"  # Pull in remote changes for the current repository and all its submodules
   pp = !"git pull ; git push origin master"
   ppd = !"git pull origin develop ; git push origin develop"
   mdm = !"git checkout master ; git pull origin master ; git push origin master ; git merge develop ; git push origin master ; git checkout develop"

   # push
   ps = push
   pom   = pull origin master
   pum = push origin master

   # rebase
   rc = rebase --continue # continue rebase
   rs = rebase --skip # skip rebase
   reb = "!r() { git rebase -i HEAD~$1; }; r"   # Interactive rebase with the given number of latest commits

   # remote
   r = remote -v
   remotes = remote -v

   # reset
   unstage = reset HEAD # remove files from index (tracking)
   uncommit = reset --soft HEAD^ # go back before last commit, with files in uncommitted state
   undo = reset --soft HEAD^
   filelog = log -u # show changes to a file
   mt = mergetool # fire up the merge tool

   # stash
   ss = stash # stash changes
   sl = stash list # list stashes
   sa = stash apply # apply stash (restore changes)
   sd = stash drop # drop stashes (destroy changes)
   stsh = stash --keep-index
   staash = stash --include-untracked
   staaash = stahs --all

   # status
   s = status -s                             # View the current working tree status using the short format
   st = status # status
   stat = status # status
   shorty = status --short --branch

   # tag
   t = tag -n                                # show tags with <n> lines of each tag message
   tags = tag -l                             # Show verbose output about tags, branches or remotes

  # init
  it = !"git init && git commit -m "root" --allow-empty"

  # merge
  merc = merge --no-ff
	grog = log --graph --abbrev-commit --decorate --all --format=format:\"%C(bold blue)%h%C(reset) - %C(bold cyan)%aD%C(dim white) - %an%C(reset) %C(bold green)(%ar)%C(reset)%C(bold yellow)%d%C(reset)%n %C(white)%s%C(reset)\"
   
[apply]
   # Detect whitespace errors when applying a patch
   whitespace = fix

[branch]
   autosetupmerge = true

[core]
  editor = vim
  excludesfile = ~/.gitignore
  attributesfile = ~/.gitattributes
  # Treat spaces before tabs, lines that are indented with 8 or more spaces, and all kinds of trailing whitespace as an error
  whitespace = space-before-tab,indent-with-non-tab,trailing-space
  autocrlf = input
  protectHFS = true
  protectNTFS = true
  sshCommand = ssh -oControlMaster=auto -oControlPersist=600 -oControlPath=/tmp/.ssh-%C

[receive]
  fsckObjects = true
  quotepath = false

[color]
   # Use colors in Git commands that are capable of colored output when outputting to the terminal
   ui = auto

[diff] 

[format]
   pretty = format:%C(blue)%ad%Creset %C(yellow)%h%C(green)%d%Creset %C(blue)%s %C(magenta) [%an]%Creset

[mergetool]
   prompt = false

#[mergetool "mvimdiff"]
#  cmd="mvim -c 'Gdiff' $MERGED" # use fugitive.vim for 3-way merge
#  keepbackup=false

[merge]
   # Include summaries of merged commits in newly created merge commit messages
   log = true
   summary = true
   verbosity = 1
#   tool = mvimdiff

# Use origin as the default remote on the master branch in all cases
[branch "master \""]
   remote = origin
   merge = refs/heads/master

[github]
   user = johndoe

# URL shorthands
[url "git@github.com:"]
   insteadOf = "gh:"
   pushInsteadOf = "github:"
   pushInsteadOf = "git://github.com/"
	 insteadOf = https://github.com

[url "git://github.com/"]
   insteadOf = "github:"

[url "git@gist.github.com:"]
   insteadOf = "gst:"
   pushInsteadOf = "gist:"
   pushInsteadOf = "git://gist.github.com/"

[url "git://gist.github.com/"]
   insteadOf = "gist:"

[url "git@gitlab.com:"]
	insteadOf = https://gitlab.com/

[url "git@gitlab.com:"]
	insteadOf = http://gitlab.com/

[user]
   email = john.doe@gmail.com
   name = John Doe
	 signingkey = DEADBEEF

[push]
	default = simple

[gc]
	auto = 64
	autopacklimit = 64
[pull]
	rebase = false
[init]
	defaultBranch = master
[fetch]
	prune = true

[credential]
	helper = osxkeychain

`

var configSampleGopass = `
# This is a gopass config file

[core]
  autoclip = true
  autoimport = true
  cliptimeout = 45
  editor = vim
  exportkeys = true
  pager = false
  notifications = true
  showsafecontent = false

[mounts]
  path = /home/johndoe/.password-store

[mounts "foo/sub"]
  path = /home/johndoe/.password-store-foo-sub

[mounts "work"]
  path = /home/johndoe/.password-store-work

[domain-alias "foo.com"]
  insteadOf = foo.de

[domain-alias "foo.com"]
  insteadOf = foo.it
`

func TestGopass(t *testing.T) {
	t.Parallel()

	c := &Configs{
		global: ParseConfig(strings.NewReader(configSampleGopass)),
	}
	c.global.noWrites = true

	assert.Equal(t, "true", c.Get("core.autoclip"))
	assert.Equal(t, "true", c.Get("core.autoimport"))
	assert.Equal(t, "45", c.Get("core.cliptimeout"))
	assert.Equal(t, "vim", c.Get("core.editor"))
	assert.Equal(t, "true", c.Get("core.exportkeys"))
	assert.Equal(t, "false", c.Get("core.pager"))
	assert.Equal(t, "true", c.Get("core.notifications"))
	assert.Equal(t, "false", c.Get("core.showsafecontent"))
	assert.Equal(t, "foo.it", c.Get("domain-alias.foo.com.insteadOf"))
	// TODO: support multivars
	// foo.de should be part of a multi-var get

	assert.Equal(t, "/home/johndoe/.password-store", c.Get("mounts.path"))
	assert.Equal(t, "/home/johndoe/.password-store-foo-sub", c.Get("mounts.foo/sub.path"))
	assert.Equal(t, "/home/johndoe/.password-store-work", c.Get("mounts.work.path"))

	t.Logf("Raw:\n%s\n", c.global.raw.String())
	t.Logf("Vars:\n%+v\n", c.global.vars)
}

func TestParseSimple(t *testing.T) {
	t.Parallel()

	c := ParseConfig(strings.NewReader(configSampleDocs))

	for k, v := range c.vars {
		t.Logf("%s => %s\n", k, v)
	}

	want := map[string]string{
		"core.filemode": "false",
		"diff.external": "/usr/local/bin/diff-wrapper",
		"diff.renames":  "true",
		"core.gitproxy": "default-proxy",
		// TODO(gitconfig): "http.sslVerify": "", // not supported, yet
		"http.https://weak.example.com.sslVerify":  "false",
		"http.https://weak.example.com.cookieFile": "/tmp/cookie.txt",
	}

	assert.Equal(t, want, c.vars)
}

func TestParseComplex(t *testing.T) {
	t.Parallel()

	c := ParseConfig(strings.NewReader(configSampleComplex))

	assert.Contains(t, maps.Keys(c.vars), "core.sshCommand")
	assert.Equal(t, "ssh -oControlMaster=auto -oControlPersist=600 -oControlPath=/tmp/.ssh-%C", c.vars["core.sshCommand"])
}

func TestParseDocs(t *testing.T) {
	t.Parallel()

	c := ParseConfig(strings.NewReader(configSampleComplex))

	// TODO(#2479) - fix parsing
	t.Skip("TODO - broken")

	assert.Equal(t, "ssh -oControlMaster=auto -oControlPersist=600 -oControlPath=/tmp/.ssh-%C", c.vars["core.sshCommand"])
}

func TestGitBinary(t *testing.T) {
	t.Skip("not ready, yet") // TODO(gitconfig) make tests pass

	cfgs := New()
	cfgs.LoadAll(".")

	cmd := exec.Command("git", "config", "--list")
	buf, err := cmd.Output()
	require.NoError(t, err)
	lines := strings.Split(string(buf), "\n")
	for _, line := range lines {
		p := strings.SplitN(line, "=", 2)
		if len(p) < 2 {
			continue
		}
		key := p[0]
		want := p[1]

		assert.Equal(t, want, cfgs.Get(key), key)
	}
}

func TestSet(t *testing.T) {
	t.Parallel()

	c := ParseConfig(strings.NewReader(configSampleDocs))
	c.noWrites = true
	require.NoError(t, c.Set("core.gitproxy", "foobar"))
	want := strings.ReplaceAll(configSampleDocs, "default-proxy", "foobar")
	assert.Equal(t, want, c.raw.String())
}

func TestUnset(t *testing.T) {
	t.Parallel()

	c := ParseConfig(strings.NewReader(configSampleDocs))
	c.noWrites = true
	require.NoError(t, c.Unset("core.filemode"))
	want := `#
# This is the config file, and
# a '#' or ';' character indicates
# a comment
#

; core variables
[core]
	; Don't trust file modes

; Our diff algorithm
[diff]
	external = /usr/local/bin/diff-wrapper
	renames = true

; Proxy settings
[core]
	gitproxy = default-proxy ; default proxy

; HTTP
[http]
    sslVerify

[http "https://weak.example.com"]
	sslVerify = false
	cookieFile = /tmp/cookie.txt
`
	assert.Equal(t, want, c.raw.String())
}

func TestSetEmptyConfig(t *testing.T) {
	t.Parallel()

	td := t.TempDir()
	c := &Config{
		path:     filepath.Join(td, "config"),
		noWrites: false,
	}
	assert.Error(t, c.Set("foobar", "baz"))
	assert.NoError(t, c.Set("foo.bar", "baz"))
	assert.Equal(t, "baz", c.vars["foo.bar"])
	buf, err := os.ReadFile(c.path)
	require.NoError(t, err)
	assert.Equal(t, "[foo]\n\tbar = baz\n", string(buf))
}

func TestList(t *testing.T) {
	t.Parallel()

	c := &Configs{
		global: ParseConfig(strings.NewReader(configSampleGopass)),
	}
	c.global.noWrites = true
	assert.Equal(t, []string{
		"mounts.foo/sub.path",
		"mounts.path",
		"mounts.work.path",
	}, c.List("mounts."))
}

func TestListSections(t *testing.T) {
	t.Parallel()

	c := &Configs{global: ParseConfig(strings.NewReader(configSampleGopass))}
	c.global.noWrites = true
	assert.Equal(t, []string{"core", "domain-alias", "mounts"}, c.ListSections())
}

func TestListSubsections(t *testing.T) {
	t.Parallel()

	c := &Configs{global: ParseConfig(strings.NewReader(configSampleGopass))}
	c.global.noWrites = true
	assert.Equal(t, []string{"foo/sub", "work"}, c.ListSubsections("mounts"))
}
