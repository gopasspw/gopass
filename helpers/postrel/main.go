package main

import (
	"context"
	"fmt"
	"html/template"
	"io/ioutil"
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
		panic(err)
	}

	// create a new GitHub milestone
	if err := createGHMilestone(ghPat, nextVer); err != nil {
		panic(err)
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

	ver := v.String()
	for _, m := range ms {
		if *m.Title == ver {
			fmt.Printf("Milestone %s exists\n", ver)
			return nil
		}
	}

	due := time.Now().Add(30 * 24 * time.Hour)
	_, _, err = client.Issues.CreateMilestone(ctx, "gopasspw", "gopass", &github.Milestone{
		Title: &ver,
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
