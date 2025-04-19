// Copyright 2021 The gopass Authors. All rights reserved.
// Use of this source code is governed by the MIT license,
// that can be found in the LICENSE file.

// Release is the first part of the gopass release automation. It's supposed
// to be run by a member of the gopass team. It will ensure that the repository
// is in a clean state and make it trivial to trigger a new release.
// You can run it without any parameters and as long as you pay close attention
// to the output it will be a breeze.
package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"text/template"
	"time"

	"github.com/blang/semver/v4"
	"github.com/gopasspw/gopass/helpers/gitutils"
)

var (
	sleep     = time.Second
	issueRE   = regexp.MustCompile(`#(\d+)\b`)
	subjectRE = regexp.MustCompile(`^(\[\w+\]\s+.*)$`)
	verTmpl   = `package main

import (
	"strings"

	"github.com/blang/semver/v4"
)

func getVersion() semver.Version {
	sv, err := semver.Parse(strings.TrimPrefix(version, "v"))
	if err == nil {
		return sv
	}

	return semver.Version{
		Major: {{ .Major }},
		Minor: {{ .Minor }},
		Patch: {{ .Patch }},
		Pre: []semver.PRVersion{
			{VersionStr: "git"},
		},
		Build: []string{"{{ .Build }}"},
	}
}
`
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
	fmt.Println("🌟 Preparing a new gopass release.")
	fmt.Println("☝  Checking pre-conditions ...")

	prevVer, nextVer := getVersions()

	// - check that workdir is clean
	if !isGitClean() {
		panic("❌ git is dirty")
	}
	fmt.Println("✅ git is clean")

	if len(nextVer.Pre) < 1 {
		// - check out master
		if err := gitCoMaster(); err != nil {
			panic(err)
		}
		fmt.Println("✅ Switched to master branch")
		// - pull from origin
		if err := gitPom(); err != nil {
			panic(err)
		}
		fmt.Println("✅ Fetched changes for master")
	}
	// - check that workdir is clean
	if !isGitClean() {
		panic("git is dirty")
	}
	fmt.Println("✅ git is still clean")

	fmt.Println()
	fmt.Printf("✅ New version will be: %s\n", nextVer.String())
	fmt.Println()
	fmt.Println("❓ Do you want to continue? (press any key to continue or Ctrl+C to abort)")
	fmt.Scanln()

	// - update deps and run tests
	if err := updateDeps(); err != nil {
		panic(err)
	}

	// - update VERSION
	if err := writeVersion(nextVer); err != nil {
		panic(err)
	}
	fmt.Println("✅ Wrote VERSION")
	time.Sleep(sleep)
	// - update version.go
	if err := writeVersionGo(nextVer); err != nil {
		panic(err)
	}
	fmt.Println("✅ Wrote version.go")
	time.Sleep(sleep)
	// - update CHANGELOG.md
	if err := writeChangelog(prevVer, nextVer); err != nil {
		panic(err)
	}
	fmt.Println("✅ Updated CHANGELOG.md")
	time.Sleep(sleep)
	// - update shell completions
	if err := updateCompletion(); err != nil {
		panic(err)
	}
	fmt.Println("✅ Updated shell completions")
	time.Sleep(sleep)
	// - update man page
	if err := updateManpage(); err != nil {
		panic(err)
	}
	fmt.Println("✅ Updated man page")
	time.Sleep(sleep)

	// - create PR
	//   git checkout -b release/vX.Y.Z
	if err := gitCoRel(nextVer); err != nil {
		panic(err)
	}
	fmt.Printf("✅ Created branch release/v%s\n", nextVer.String())
	time.Sleep(sleep)

	// commit changes
	if err := gitCommit(nextVer); err != nil {
		panic(err)
	}
	fmt.Printf("✅ Committed changes to release/v%s\n", nextVer.String())
	time.Sleep(sleep)

	fmt.Println("🏁 Preparation finished")
	time.Sleep(sleep)

	fmt.Printf("⚠ Prepared release of gopass %s.\n", nextVer.String())
	time.Sleep(sleep)

	fmt.Printf("⚠ Run 'git push <remote> release/v%s' to push this branch and open a PR against gopasspw/gopass master.\n", nextVer.String())
	time.Sleep(sleep)

	fmt.Printf("⚠ Get the PR merged and run 'git tag -s v%s && git push origin v%s' to kick off the release process.\n", nextVer.String(), nextVer.String())
	time.Sleep(sleep)
	fmt.Println()

	fmt.Println("💎🙌 Done 🚀🚀🚀🚀🚀🚀")
}

func getVersions() (semver.Version, semver.Version) {
	nextVerFlag := ""
	if len(os.Args) > 1 && !strings.HasPrefix(os.Args[2], "-test.") {
		nextVerFlag = strings.TrimSpace(strings.TrimPrefix(os.Args[1], "v"))
	}
	prevVerFlag := ""
	if len(os.Args) > 2 && !strings.HasPrefix(os.Args[2], "-test.") {
		prevVerFlag = strings.TrimSpace(strings.TrimPrefix(os.Args[2], "v"))
	}

	// obtain the last tagged version from git
	gitVer, err := gitVersion()
	if err != nil {
		panic(err)
	}

	// read the version file to get the last committed version
	vfVer, err := versionFile()
	if err != nil {
		panic(err)
	}

	prevVer := gitVer
	if prevVerFlag != "" {
		prevVer = semver.MustParse(prevVerFlag)
	}

	if gitVer.NE(vfVer) {
		fmt.Printf("git version: %q != VERSION: %q\n", gitVer.String(), vfVer.String())
		if prevVerFlag == "" && len(vfVer.Pre) < 1 {
			usage()
			panic("version mismatch")
		}
	}

	nextVer := prevVer
	if nextVerFlag != "" {
		nextVer = semver.MustParse(nextVerFlag)
		if nextVer.LTE(prevVer) {
			usage()
			panic("next version must be greather than the previous version")
		}
	} else {
		nextVer.IncrementPatch()
		if len(vfVer.Pre) > 0 {
			nextVer = vfVer
			nextVer.Pre = nil
		}
	}

	fmt.Printf(`☝ Version overview
  Git (latest tag):  %q
  VERSION:           %q
  Next version flag: %q
  Prev version flag: %q

Will use
  Previous: %q
  Next:     %q
`,
		gitVer,
		vfVer,
		prevVerFlag,
		nextVerFlag,
		prevVer,
		nextVer)

	return prevVer, nextVer
}

func updateDeps() error {
	cmd := exec.Command("make", "upgrade")
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return err
	}

	if sv := os.Getenv("GOPASS_NOTEST"); sv != "" {
		fmt.Printf("⚠ GOPASS_NOTEST=%v, skipping 'make gha-linux'", sv)

		return nil
	}

	td := os.TempDir()
	fn := filepath.Join(td, "gopass-release.log")
	fh, err := os.OpenFile(fn, os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		return err
	}

	cmd = exec.Command("make", "gha-linux")
	cmd.Stderr = os.Stderr
	cmd.Stdout = fh
	cmd.Env = []string{
		"LANG=en_US.UTF-8",
		"PATH=/usr/local/bin:/usr/local/sbin:/usr/sbin:/usr/bin:/bin:" + os.Getenv("GOBIN"),
	}

	for _, v := range os.Environ() {
		if strings.HasPrefix(v, "GO") {
			cmd.Env = append(cmd.Env, v)
		}
		if strings.HasPrefix(v, "HOME=") {
			cmd.Env = append(cmd.Env, v)
		}
	}

	if err := cmd.Run(); err != nil {
		_ = fh.Close()
		fmt.Printf("⚠ 'make gha-linux' failed. Please see the log at %s!", fn)

		return err
	}

	// remove the log, we don't need it anymore
	_ = fh.Close()
	_ = os.RemoveAll(td)

	return nil
}

func gitCoMaster() error {
	return gitutils.GitCoMaster(".")
}

func gitPom() error {
	return gitutils.GitPom(".")
}

func gitCoRel(v semver.Version) error {
	return gitutils.GitCoBranch(".", "release/v"+v.String())
}

func gitCommit(v semver.Version) error {
	args := []string{
		"CHANGELOG.md",
		"VERSION",
		"version.go",
		"gopass.1",
		"*.completion",
		"go.mod",
		"go.sum",
		"pkg/pwgen/pwrules/pwrules_gen.go",
	}
	return gitutils.GitCommit(".", "Tag v"+v.String(), args...)
}

func writeChangelog(prev, next semver.Version) error {
	cl, err := changelogEntries(prev)
	if err != nil {
		panic(err)
	}

	// prepend the new changelog entries by first writing the
	// new content in a new file ...
	fh, err := os.Create("CHANGELOG.new")
	if err != nil {
		return err
	}
	defer fh.Close()

	ofh, err := os.Open("CHANGELOG.md")
	if err != nil {
		return err
	}
	defer ofh.Close()

	scanner := bufio.NewScanner(ofh)

	var written bool
	for scanner.Scan() {
		line := scanner.Text()

		// insert the new section before the last entry
		if strings.HasPrefix(line, "## ") && !written {
			fmt.Fprintf(fh, "## %s / %s\n\n", next.String(), time.Now().UTC().Format("2006-01-02"))
			for _, e := range cl {
				fmt.Fprint(fh, "* ")
				fmt.Fprintln(fh, e)
			}
			fmt.Fprintln(fh)

			written = true
		}

		// all existing lines are just copied over
		fmt.Fprintln(fh, line)
	}

	// renaming the new file to the old file
	return os.Rename("CHANGELOG.new", "CHANGELOG.md")
}

func updateCompletion() error {
	cmd := exec.Command("make", "completion")
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func updateManpage() error {
	cmd := exec.Command("make", "man")
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func writeVersion(v semver.Version) error {
	return os.WriteFile("VERSION", []byte(v.String()+"\n"), 0o644)
}

type tplPayload struct {
	Major uint64
	Minor uint64
	Patch uint64
	Build string
}

func writeVersionGo(v semver.Version) error {
	tmpl, err := template.New("version").Parse(verTmpl)
	if err != nil {
		return err
	}
	fh, err := os.Create("version.go")
	if err != nil {
		return err
	}
	defer fh.Close()

	build := "HEAD"
	if sv, err := gitCommitHash(); err == nil {
		build = sv
	}

	return tmpl.Execute(fh, tplPayload{
		Major: v.Major,
		Minor: v.Minor,
		Patch: v.Patch,
		Build: build,
	})
}

func isGitClean() bool {
	return gitutils.IsGitClean(".")
}

func versionFile() (semver.Version, error) {
	buf, err := os.ReadFile("VERSION")
	if err != nil {
		return semver.Version{}, err
	}

	return semver.Parse(strings.TrimSpace(string(buf)))
}

func gitCommitHash() (string, error) {
	buf, err := exec.Command("git", "rev-parse", "--short", "HEAD").CombinedOutput()
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(buf)), nil
}

func gitVersion() (semver.Version, error) {
	buf, err := exec.Command("git", "tag", "--sort=version:refname").CombinedOutput()
	if err != nil {
		return semver.Version{}, err
	}

	lines := strings.Split(strings.TrimSpace(string(buf)), "\n")
	if len(lines) < 1 {
		return semver.Version{}, fmt.Errorf("no output")
	}

	for i := len(lines); i > 0; i-- {
		sv := strings.TrimPrefix(lines[i-1], "v")
		v, err := semver.Parse(sv)
		if err != nil {
			continue
		}
		if len(v.Pre) > 0 {
			continue
		}

		return v, nil
	}

	return semver.Version{}, fmt.Errorf("no valid version found")
}

func changelogEntries(since semver.Version) ([]string, error) {
	// set up a custom output format for the git log command to make it easier to parse here.
	gitSep := "@@@GIT-SEP@@@"
	gitDelim := "@@@GIT-DELIM@@@"
	// full hash - subject - body
	// note: we don't use the hash at the moment
	prettyFormat := gitSep + "%H" + gitDelim + "%s" + gitDelim + "%b" + gitSep
	args := []string{
		"log",
		"v" + since.String() + "..HEAD",
		"--pretty=" + prettyFormat,
	}
	buf, err := exec.Command("git", args...).CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to run git %+v with error %w: %s", args, err, string(buf))
	}

	// gitSep separates each commit from the next
	notes := make([]string, 0, 10)
	commits := strings.Split(string(buf), gitSep)
	for _, commit := range commits {
		commit := strings.TrimSpace(commit)
		if commit == "" {
			continue
		}

		// inside each commit gitDelim seaparates each field from each other
		// p[0] - full hash
		// p[1] - subject
		// p[2] - body (might be empty)
		p := strings.Split(commit, gitDelim)
		if len(p) < 3 {
			// invalid entry, shouldn't happen
			continue
		}

		subject := strings.TrimSpace(p[1])

		// extract github issue numbers from the subject
		issues := []string{}
		if m := issueRE.FindStringSubmatch(strings.TrimSpace(subject)); len(m) > 1 {
			issues = append(issues, m[1])
		}

		// try to extract the release note from the subject
		if m := subjectRE.FindStringSubmatch(subject); len(m) > 1 {
			notes = append(notes, m[1])

			continue
		}

		// if no suitable subject was parsed, try to parse the body as well
		for _, line := range strings.Split(p[2], "\n") {
			line := strings.TrimSpace(line)

			if m := issueRE.FindStringSubmatch(line); len(m) > 1 {
				issues = append(issues, m[1])
			}

			if !strings.HasPrefix(line, "RELEASE_NOTES=") {
				continue
			}
			p := strings.Split(line, "=")
			if len(p) < 2 {
				continue
			}
			val := p[1]
			if strings.ToLower(val) == "n/a" {
				continue
			}
			if len(issues) > 0 {
				val += " (#" + strings.Join(issues, ", #") + ")"
			}
			notes = append(notes, val)
		}
	}

	sort.Strings(notes)

	return notes, nil
}

func usage() {
	fmt.Printf("Usage: %s [next version] [prev version]\n", "go run helpers/release/main.go")
}
