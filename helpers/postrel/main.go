package main

import (
	"bufio"
	"context"
	"crypto/sha256"
	"crypto/sha512"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/blang/semver/v4"
	"github.com/google/go-github/v33/github"
	"golang.org/x/oauth2"
)

const logo = `
   __     _    _ _      _ _   ___   ___
 /'_ '\ /'_'\ ( '_'\  /'_' )/',__)/',__)
( (_) |( (_) )| (_) )( (_| |\__, \\__, \
'\__  |'\___/'| ,__/''\__,_)(____/(____/
( )_) |       | |
 \___/'       (_)
`

func main() {
	fmt.Println(logo)
	fmt.Println()
	fmt.Println("🌟 Performing post-release cleanup.")

	curVer, err := versionFile()
	if err != nil {
		panic(err)
	}
	nextVer := curVer
	nextVer.IncrementPatch()

	htmlDir := "../gopasspw.github.io"
	if h := os.Getenv("GOPASS_HTMLDIR"); h != "" {
		htmlDir = h
	}

	ghPat := os.Getenv("GITHUB_TOKEN")
	if ghPat == "" {
		panic("❌ Please set GITHUB_TOKEN")
	}

	fmt.Println()
	fmt.Printf("✅ Current version is: %s\n", curVer.String())
	fmt.Printf("✅ New version milestone will be: %s\n", nextVer.String())
	fmt.Printf("✅ Expecting HTML in: %s\n", htmlDir)
	fmt.Println()
	fmt.Println("❓ Do you want to continue? (press any key to continue or Ctrl+C to abort)")
	fmt.Scanln()

	// update gopass.pw
	if err := updateGopasspw(htmlDir, curVer); err != nil {
		fmt.Printf("Failed to update gopasspw.github.io: %s\n", err)
	}

	// create a new GitHub milestone
	if err := createGHMilestone(ghPat, nextVer); err != nil {
		fmt.Printf("Failed to create GitHub milestones: %s\n", err)
	}

	// send PRs to update gopass ports
	if err := updateRepos(curVer); err != nil {
		fmt.Printf("Failed to update repos: %s\n", err)
	}

	// TODO tweet about the new release

	fmt.Println("💎🙌 Done 🚀🚀🚀🚀🚀🚀")
}

func createGHMilestone(pat string, v semver.Version) error {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: pat},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	ms, _, err := client.Issues.ListMilestones(ctx, "gopasspw", "gopass", nil)
	if err != nil {
		return err
	}

	if err := createSingleGHMilestone(ctx, client, v.String(), 1, ms); err != nil {
		return err
	}

	v.IncrementPatch()
	return createSingleGHMilestone(ctx, client, v.String(), 2, ms)
}

func createSingleGHMilestone(ctx context.Context, client *github.Client, title string, offset int, ms []*github.Milestone) error {
	for _, m := range ms {
		if *m.Title == title {
			fmt.Printf("Milestone %s exists\n", title)
			return nil
		}
	}

	due := time.Now().Add(time.Duration(offset) * 30 * 24 * time.Hour)
	_, _, err := client.Issues.CreateMilestone(ctx, "gopasspw", "gopass", &github.Milestone{
		Title: &title,
		DueOn: &due,
	})

	return err
}

func updateGopasspw(dir string, ver semver.Version) error {
	buf, err := ioutil.ReadFile(filepath.Join(dir, "index.tpl"))
	if err != nil {
		return err
	}

	tmpl, err := template.New("index").Parse(string(buf))
	if err != nil {
		return err
	}

	fh, err := os.Create(filepath.Join(dir, "index.html"))
	if err != nil {
		return err
	}
	defer fh.Close()

	type pl struct {
		Version string
	}

	if err := tmpl.Execute(fh, pl{
		Version: ver.String(),
	}); err != nil {
		return err
	}

	return gitCommitAndPush(dir, ver)
}

func gitCommitAndPush(dir string, v semver.Version) error {
	cmd := exec.Command("git", "add", "index.html")
	cmd.Dir = dir
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}

	cmd = exec.Command("git", "commit", "-s", "-m", "Update to v"+v.String())
	cmd.Dir = dir
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}

	cmd = exec.Command("git", "push", "origin", "master")
	cmd.Dir = dir
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func versionFile() (semver.Version, error) {
	buf, err := os.ReadFile("VERSION")
	if err != nil {
		return semver.Version{}, err
	}
	return semver.Parse(strings.TrimSpace(string(buf)))
}

func updateRepos(v semver.Version) error {
	relURL := fmt.Sprintf("https://github.com/gopasspw/gopass/releases/download/v%s/gopass-%s.tar.gz", v.String(), v.String())
	// fetch https://github.com/gopasspw/gopass/archive/vVER.tar.gz
	// compute sha256, sha512
	_, relSHA512s, err := checksum(relURL)
	if err != nil {
		return err
	}
	arcURL := fmt.Sprintf("https://github.com/gopasspw/gopass/archive/v%s.tar.gz", v.String())
	// fetch https://github.com/gopasspw/gopass/archive/vVER.tar.gz
	// compute sha256, sha512
	arcSHA256s, arcSHA512s, err := checksum(arcURL)
	if err != nil {
		return err
	}

	for _, upd := range []struct {
		Distro string
		UpFn   func() error
	}{
		{
			Distro: "AlpineLinux",
			UpFn: func() error {
				return updateAlpine(arcURL, v, arcSHA512s)
			},
		},
		{
			Distro: "Homebrew",
			UpFn: func() error {
				return updateHomebrew(relURL, v, relSHA512s)
			},
		},
		{
			Distro: "Termux",
			UpFn: func() error {
				return updateTermux(arcURL, v, arcSHA256s)
			},
		},
		{
			Distro: "VoidLinux",
			UpFn: func() error {
				return updateVoid(arcURL, v, arcSHA256s)
			},
		},
	} {
		fmt.Println("------------------------------")
		fmt.Printf("Updating: %s ...\n", upd.Distro)
		if err := upd.UpFn(); err != nil {
			fmt.Printf("ERROR: %s\n", err)
			continue
		}
		fmt.Println("OK")
	}

	return nil
}

func updateAlpine(url string, v semver.Version, sha512 string) error {
	dir := "../repos/alpine/"
	if d := os.Getenv("GOPASS_ALPINE_PKG_DIR"); d != "" {
		dir = d
	}

	r := &repo{
		ver: v,
		url: url,
		dir: dir,
	}

	if err := r.updatePrepare(); err != nil {
		return err
	}

	// update community/gopass/APKBUILD
	buildFn := "community/gopass/APKBUILD"
	buildPath := filepath.Join(dir, buildFn)

	repl := map[string]string{
		"pkgver=":     "pkgver=" + v.String(),
		"sha512sums=": "sha512sums=\"" + sha512 + "  gopass-" + v.String() + ".tar.gz\"",
		"source=":     `source="$pkgname-$pkgver.tar.gz::https://github.com/gopasspw/gopass/archive/v$pkgver.tar.gz"`,
	}
	if err := updateBuild(buildPath, repl); err != nil {
		return err
	}

	if err := r.updateFinalize("community/gopass: upgrade to "+v.String(), buildFn); err != nil {
		return err
	}

	// TODO could open an MR: https://docs.gitlab.com/ce/api/merge_requests.html#create-mhttps://docs.gitlab.com/ce/api/merge_requests.html#comments-on-merge-requestsr
	return nil
}

func updateHomebrew(url string, v semver.Version, sha256 string) error {
	dir := "../repos/homebrew/"
	if d := os.Getenv("GOPASS_HOMEBREW_PKG_DIR"); d != "" {
		dir = d
	}

	r := &repo{
		ver: v,
		url: url,
		dir: dir,
	}

	if err := r.updatePrepare(); err != nil {
		return err
	}

	// update Formula/gopass.rb
	buildFn := "Formula/gopass.rb"
	buildPath := filepath.Join(dir, buildFn)

	repl := map[string]string{
		"url \"https://github.com/": "url \"" + url + "\"",
		"sha256 \"":                 "sha256 \"" + sha256 + "\"",
	}
	if err := updateBuild(
		buildPath,
		repl,
	); err != nil {
		return err
	}
	if err := r.updateFinalize("", buildFn); err != nil {
		return err
	}
	// TODO could open a PR: https://pkg.go.dev/github.com/google/go-github/v33@v33.0.0/github#PullRequestsService.Create
	return nil
}

func updateTermux(url string, v semver.Version, sha256 string) error {
	dir := "../repos/termux/"
	if d := os.Getenv("GOPASS_TERMUX_PKG_DIR"); d != "" {
		dir = d
	}

	r := &repo{
		ver: v,
		url: url,
		dir: dir,
	}

	if err := r.updatePrepare(); err != nil {
		return err
	}

	// update packages/gopass/build.sh
	buildFn := "packages/gopass/build.sh"
	buildPath := filepath.Join(dir, buildFn)

	repl := map[string]string{
		"TERMUX_PKG_VERSION": "TERMUX_PKG_VERSION=" + v.String(),
		"TERMUX_PKG_SHA256":  "TERMUX_PKG_SHA256=" + sha256,
		"TERMUX_PKG_REVISON": "TERMUX_PKG_REVISION=1",
		"TERMUX_PKG_SRCURL":  `TERMUX_PKG_SRCURL=https://github.com/gopasspw/gopass/archive/v$TERMUX_PKG_VERSION.tar.gz`,
	}
	if err := updateBuild(
		buildPath,
		repl,
	); err != nil {
		return err
	}
	if err := r.updateFinalize("", buildFn); err != nil {
		return err
	}

	// TODO could open a PR: https://pkg.go.dev/github.com/google/go-github/v33@v33.0.0/github#PullRequestsService.Create
	return nil
}

func updateVoid(url string, v semver.Version, sha256 string) error {
	dir := "../repos/void/"
	if d := os.Getenv("GOPASS_VOID_PKG_DIR"); d != "" {
		dir = d
	}

	r := &repo{
		ver: v,
		url: url,
		dir: dir,
	}

	if err := r.updatePrepare(); err != nil {
		return err
	}

	// update srcpkgs/gopass/template
	buildFn := "srcpkgs/gopass/template"
	buildPath := filepath.Join(dir, buildFn)

	repl := map[string]string{
		"version=":   "version=" + v.String(),
		"checksum=":  "checksum=" + sha256,
		"revision=":  "revision=1",
		"distfiles=": `distfiles="https://github.com/gopasspw/gopass/archive/v${version}.tar.gz"`,
	}
	if err := updateBuild(
		buildPath,
		repl,
	); err != nil {
		return err
	}

	if err := r.updateFinalize("", buildFn); err != nil {
		return err
	}

	// TODO could open a PR: https://pkg.go.dev/github.com/google/go-github/v33@v33.0.0/github#PullRequestsService.Create
	return nil
}

func checksum(url string) (string, string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	s2 := sha256.New()
	s5 := sha512.New()
	w := io.MultiWriter(s2, s5)

	_, err = io.Copy(w, resp.Body)
	if err != nil {
		return "", "", err
	}

	return fmt.Sprintf("%x", s2.Sum(nil)), fmt.Sprintf("%x", s5.Sum(nil)), nil
}

func updateBuild(path string, m map[string]string) error {
	fin, err := os.Open(path)
	if err != nil {
		return err
	}
	defer fin.Close()

	npath := path + ".new"
	fout, err := os.Create(npath)
	if err != nil {
		return err
	}
	defer fout.Close()

	s := bufio.NewScanner(fin)
SCAN:
	for s.Scan() {
		line := s.Text()
		for match, repl := range m {
			if strings.HasPrefix(line, match) {
				fmt.Fprintln(fout, repl)
				continue SCAN
			}
		}
		fmt.Fprintln(fout, line)
	}

	return os.Rename(npath, path)
}

type repo struct {
	ver semver.Version // gopass version
	url string         // gopass download url
	dir string         // repo dir
}

func (r *repo) updatePrepare() error {
	// git co master
	if err := r.gitCoMaster(); err != nil {
		return err
	}
	if !r.isGitClean() {
		return fmt.Errorf("git is dirty")
	}
	// git pull origin master
	if err := r.gitPom(); err != nil {
		return err
	}
	// git co -b gopass-VER
	return r.gitBranch()
}

func (r *repo) updateFinalize(msg, path string) error {
	// git commit -m 'gopass: update to VER'
	if err := r.gitCommit(msg, path); err != nil {
		return err
	}
	// git push myfork gopass-VER
	return r.gitPush("myfork")

}

func (r *repo) gitCoMaster() error {
	cmd := exec.Command("git", "checkout", "master")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = r.dir
	return cmd.Run()
}

func (r *repo) gitBranch() error {
	cmd := exec.Command("git", "checkout", "-b", "gopass-"+r.ver.String())
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = r.dir
	return cmd.Run()
}

func (r *repo) gitPom() error {
	cmd := exec.Command("git", "pull", "origin", "master")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = r.dir
	return cmd.Run()
}

func (r *repo) gitPush(remote string) error {
	cmd := exec.Command("git", "push", remote, "gopass-"+r.ver.String())
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = r.dir
	return cmd.Run()
}

func (r *repo) gitCommit(msg string, files ...string) error {
	args := []string{"add"}
	args = append(args, files...)

	cmd := exec.Command("git", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = r.dir
	if err := cmd.Run(); err != nil {
		return err
	}

	if msg == "" {
		msg = "gopass: update to " + r.ver.String()
	}
	cmd = exec.Command("git", "commit", "-s", "-m", msg)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = r.dir
	return cmd.Run()
}

func (r *repo) isGitClean() bool {
	cmd := exec.Command("git", "diff", "--stat")
	cmd.Dir = r.dir

	buf, err := cmd.CombinedOutput()
	if err != nil {
		panic(err)
	}
	return strings.TrimSpace(string(buf)) == ""
}
