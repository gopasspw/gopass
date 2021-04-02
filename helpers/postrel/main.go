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
	ctx := context.Background()

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

	ghCl, err := newGHClient(ctx)
	if err != nil {
		panic(err)
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
	if err := ghCl.createMilestones(ctx, nextVer); err != nil {
		fmt.Printf("Failed to create GitHub milestones: %s\n", err)
	}

	// send PRs to update gopass ports
	ghFork := os.Getenv("GITHUB_FORK")
	if ghFork == "" {
		panic("Please set GITHUB_FORK")
	}

	upd, err := newRepoUpdater(ghCl.client, curVer, ghFork)
	if err != nil {
		fmt.Printf("Failed to create repo updater: %s\n", err)
	} else {
		upd.update(ctx)
	}

	// TODO tweet about the new release

	fmt.Println("💎🙌 Done 🚀🚀🚀🚀🚀🚀")
}

type ghClient struct {
	client *github.Client
	org    string
	repo   string
}

func newGHClient(ctx context.Context) (*ghClient, error) {
	pat := os.Getenv("GITHUB_TOKEN")
	if pat == "" {
		return nil, fmt.Errorf("❌ Please set GITHUB_TOKEN")
	}

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: pat},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	return &ghClient{
		client: client,
		org:    "gopasspw",
		repo:   "gopass",
	}, nil
}

func (g *ghClient) createMilestones(ctx context.Context, v semver.Version) error {
	ms, _, err := g.client.Issues.ListMilestones(ctx, g.org, g.repo, nil)
	if err != nil {
		return err
	}

	if err := g.createMilestone(ctx, v.String(), 1, ms); err != nil {
		return err
	}

	v.IncrementPatch()
	return g.createMilestone(ctx, v.String(), 2, ms)
}

func (g *ghClient) createMilestone(ctx context.Context, title string, offset int, ms []*github.Milestone) error {
	for _, m := range ms {
		if *m.Title == title {
			fmt.Printf("Milestone %s exists\n", title)
			return nil
		}
	}

	due := time.Now().Add(time.Duration(offset) * 30 * 24 * time.Hour)
	_, _, err := g.client.Issues.CreateMilestone(ctx, g.org, g.repo, &github.Milestone{
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

type repoUpdater struct {
	github    *github.Client
	ghFork    string
	v         semver.Version
	relURL    string
	arcURL    string
	relSHA256 string
	relSHA512 string
	arcSHA256 string
	arcSHA512 string
}

func newRepoUpdater(client *github.Client, v semver.Version, fork string) (*repoUpdater, error) {
	relURL := fmt.Sprintf("https://github.com/gopasspw/gopass/releases/download/v%s/gopass-%s.tar.gz", v.String(), v.String())
	// fetch https://github.com/gopasspw/gopass/archive/vVER.tar.gz
	// compute sha256, sha512
	relSHA256, relSHA512, err := checksum(relURL)
	if err != nil {
		return nil, err
	}
	arcURL := fmt.Sprintf("https://github.com/gopasspw/gopass/archive/v%s.tar.gz", v.String())
	// fetch https://github.com/gopasspw/gopass/archive/vVER.tar.gz
	// compute sha256, sha512
	arcSHA256, arcSHA512, err := checksum(arcURL)
	if err != nil {
		return nil, err
	}

	return &repoUpdater{
		github:    client,
		ghFork:    fork,
		v:         v,
		relURL:    relURL,
		arcURL:    arcURL,
		relSHA256: relSHA256,
		relSHA512: relSHA512,
		arcSHA256: arcSHA256,
		arcSHA512: arcSHA512,
	}, nil
}

func (u *repoUpdater) update(ctx context.Context) {
	for _, upd := range []struct {
		Distro string
		UpFn   func(context.Context) error
	}{
		{
			Distro: "AlpineLinux",
			UpFn:   u.updateAlpine,
		},
		{
			Distro: "Homebrew",
			UpFn:   u.updateHomebrew,
		},
		{
			Distro: "Termux",
			UpFn:   u.updateTermux,
		},
		{
			Distro: "VoidLinux",
			UpFn:   u.updateVoid,
		},
	} {
		fmt.Println("------------------------------")
		fmt.Printf("Updating: %s ...\n", upd.Distro)
		if err := upd.UpFn(ctx); err != nil {
			fmt.Printf("\tERROR: %s\n", err)
			continue
		}
		fmt.Println("\tOK")
	}
}

func (u *repoUpdater) updateAlpine(ctx context.Context) error {
	dir := "../repos/alpine/"
	if d := os.Getenv("GOPASS_ALPINE_PKG_DIR"); d != "" {
		dir = d
	}

	r := &repo{
		ver: u.v,
		url: u.arcURL,
		dir: dir,
		msg: "community/gopass: upgrade to " + u.v.String(),
	}

	if err := r.updatePrepare(); err != nil {
		return err
	}

	// update community/gopass/APKBUILD
	buildFn := "community/gopass/APKBUILD"
	buildPath := filepath.Join(dir, buildFn)

	repl := map[string]*string{
		"pkgver=":     strp("pkgver=" + u.v.String()),
		"sha512sums=": strp("sha512sums=\"" + u.arcSHA512 + "  gopass-" + u.v.String() + ".tar.gz\""),
		"source=":     strp(`source="$pkgname-$pkgver.tar.gz::https://github.com/gopasspw/gopass/archive/v$pkgver.tar.gz"`),
	}
	if err := updateBuild(buildPath, repl); err != nil {
		return err
	}

	if err := r.updateFinalize(buildFn); err != nil {
		return err
	}

	// TODO could open an MR: https://docs.gitlab.com/ce/api/merge_requests.html#create-mhttps://docs.gitlab.com/ce/api/merge_requests.html#comments-on-merge-requestsr
	return nil
}

func (u *repoUpdater) updateHomebrew(ctx context.Context) error {
	dir := "../repos/homebrew/"
	if d := os.Getenv("GOPASS_HOMEBREW_PKG_DIR"); d != "" {
		dir = d
	}

	r := &repo{
		ver: u.v,
		url: u.relURL,
		dir: dir,
	}

	if err := r.updatePrepare(); err != nil {
		return err
	}

	// update Formula/gopass.rb
	buildFn := "Formula/gopass.rb"
	buildPath := filepath.Join(dir, buildFn)

	repl := map[string]*string{
		"url \"https://github.com/": strp("url \"" + u.relURL + "\""),
		"sha256 \"":                 strp("sha256 \"" + u.relSHA256 + "\""),
	}
	if err := updateBuild(
		buildPath,
		repl,
	); err != nil {
		return err
	}
	if err := r.updateFinalize(buildFn); err != nil {
		return err
	}

	return u.createPR(ctx, r.commitMsg(), u.ghFork+":"+r.branch(), "Homebrew", "homebrew-core")
}

func (u *repoUpdater) updateTermux(ctx context.Context) error {
	dir := "../repos/termux/"
	if d := os.Getenv("GOPASS_TERMUX_PKG_DIR"); d != "" {
		dir = d
	}

	r := &repo{
		ver: u.v,
		url: u.arcURL,
		dir: dir,
	}

	if err := r.updatePrepare(); err != nil {
		return err
	}

	// update packages/gopass/build.sh
	buildFn := "packages/gopass/build.sh"
	buildPath := filepath.Join(dir, buildFn)

	repl := map[string]*string{
		"TERMUX_PKG_VERSION": strp("TERMUX_PKG_VERSION=" + u.v.String()),
		"TERMUX_PKG_SHA256":  strp("TERMUX_PKG_SHA256=" + u.arcSHA256),
		"TERMUX_PKG_REVISON": nil, // a new release shouldn't have a revision
		"TERMUX_PKG_SRCURL":  strp(`TERMUX_PKG_SRCURL=https://github.com/gopasspw/gopass/archive/v$TERMUX_PKG_VERSION.tar.gz`),
	}
	if err := updateBuild(
		buildPath,
		repl,
	); err != nil {
		return err
	}
	if err := r.updateFinalize(buildFn); err != nil {
		return err
	}

	return u.createPR(ctx, r.commitMsg(), u.ghFork+":"+r.branch(), "termux", "termux-packages")
}

func (u *repoUpdater) updateVoid(ctx context.Context) error {
	dir := "../repos/void/"
	if d := os.Getenv("GOPASS_VOID_PKG_DIR"); d != "" {
		dir = d
	}

	r := &repo{
		ver: u.v,
		url: u.arcURL,
		dir: dir,
	}

	if err := r.updatePrepare(); err != nil {
		return err
	}

	// update srcpkgs/gopass/template
	buildFn := "srcpkgs/gopass/template"
	buildPath := filepath.Join(dir, buildFn)

	repl := map[string]*string{
		"version=":   strp("version=" + u.v.String()),
		"checksum=":  strp("checksum=" + u.arcSHA256),
		"revision=":  nil,
		"distfiles=": strp(`distfiles="https://github.com/gopasspw/gopass/archive/v${version}.tar.gz"`),
	}
	if err := updateBuild(
		buildPath,
		repl,
	); err != nil {
		return err
	}

	if err := r.updateFinalize(buildFn); err != nil {
		return err
	}

	return u.createPR(ctx, r.commitMsg(), u.ghFork+":"+r.branch(), "void-linux", "void-packages")
}

func (u *repoUpdater) createPR(ctx context.Context, title, from, toOrg, toRepo string) error {
	newPR := &github.NewPullRequest{
		Title:               github.String(title),
		Head:                github.String(from),
		Base:                github.String("master"),
		Body:                github.String(title),
		MaintainerCanModify: github.Bool(true),
	}

	pr, _, err := u.github.PullRequests.Create(ctx, toOrg, toRepo, newPR)
	fmt.Println(pr)
	return err
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

func updateBuild(path string, m map[string]*string) error {
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
				if repl != nil {
					fmt.Fprintln(fout, *repl)
				}
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
	msg string
}

func (r *repo) branch() string {
	return fmt.Sprintf("gopass-%s", r.ver.String())
}

func (r *repo) commitMsg() string {
	if r.msg != "" {
		return r.msg
	}
	return "gopass: update to " + r.ver.String()
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

func (r *repo) updateFinalize(path string) error {
	// git commit -m 'gopass: update to VER'
	if err := r.gitCommit(path); err != nil {
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
	cmd := exec.Command("git", "checkout", "-b", r.branch())
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
	cmd := exec.Command("git", "push", remote, r.branch())
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = r.dir
	return cmd.Run()
}

func (r *repo) gitCommit(files ...string) error {
	args := []string{"add"}
	args = append(args, files...)

	cmd := exec.Command("git", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = r.dir
	if err := cmd.Run(); err != nil {
		return err
	}

	cmd = exec.Command("git", "commit", "-s", "-m", r.commitMsg())
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

func strp(s string) *string {
	return &s
}
