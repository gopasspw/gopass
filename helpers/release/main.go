package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"sort"
	"strings"
	"text/template"
	"time"

	"github.com/blang/semver/v4"
)

var sleep = time.Second
var issueRE = regexp.MustCompile(`#(\d+)\b`)
var verTmpl = `package main

import (
	"strings"

	"github.com/blang/semver/v4"
)

func getVersion() semver.Version {
	sv, err := semver.Parse(strings.TrimPrefix(version, "v"))
	if err == nil {
		if commit != "" {
			sv.Build = []string{commit}
		}
		return sv
	}
	return semver.Version{
		Major: {{ .Major }},
		Minor: {{ .Minor }},
		Patch: {{ .Patch }},
		Pre: []semver.PRVersion{
			{VersionStr: "git"},
		},
		Build: []string{"HEAD"},
	}
}
`

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
	// - check that workdir is clean
	if !isGitClean() {
		panic("❌ git is dirty")
	}
	fmt.Println("✅ git is clean")
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
	// - check that workdir is clean
	if !isGitClean() {
		panic("git is dirty")
	}
	fmt.Println("✅ git is still clean")

	prevVer, nextVer := getVersions()

	fmt.Println()
	fmt.Printf("✅ New version will be: %s\n", nextVer.String())
	fmt.Println()
	fmt.Println("❓ Do you want to continue? (press any key to continue or Ctrl+C to abort)")
	fmt.Scanln()

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

	fmt.Printf("⚠ Get the PR merged and run 'git tag v%s && git push origin v%s' to kick off the release process.\n", nextVer.String(), nextVer.String())
	time.Sleep(sleep)
	fmt.Println()

	fmt.Println("💎🙌 Done 🚀🚀🚀🚀🚀🚀")
}

func getVersions() (semver.Version, semver.Version) {
	nextVerFlag := ""
	if len(os.Args) > 1 {
		nextVerFlag = strings.TrimSpace(strings.TrimPrefix(os.Args[1], "v"))
	}
	prevVerFlag := ""
	if len(os.Args) > 2 {
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

func gitCoMaster() error {
	cmd := exec.Command("git", "checkout", "master")
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func gitPom() error {
	cmd := exec.Command("git", "pull", "origin", "master")
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func gitCoRel(v semver.Version) error {
	cmd := exec.Command("git", "checkout", "-b", "release/v"+v.String())
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func gitCommit(v semver.Version) error {
	cmd := exec.Command("git", "add", "CHANGELOG.md", "VERSION", "version.go")
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}
	cmd = exec.Command("git", "commit", "-s", "-m", "Tag v"+v.String(), "-m", "RELEASE_NOTES=n/a")
	cmd.Stderr = os.Stderr
	return cmd.Run()
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

	fmt.Fprintf(fh, "## %s / %s\n\n", next.String(), time.Now().UTC().Format("2006-01-02"))
	for _, e := range cl {
		fmt.Fprint(fh, "* ")
		fmt.Fprintln(fh, e)
	}
	fmt.Fprintln(fh)

	ofh, err := os.Open("CHANGELOG.md")
	if err != nil {
		return err
	}
	defer ofh.Close()

	// then appending any existing content from the old file and ...
	if _, err := io.Copy(fh, ofh); err != nil {
		return err
	}

	// renaming the new file to the old file
	return os.Rename("CHANGELOG.new", "CHANGELOG.md")
}

func writeVersion(v semver.Version) error {
	return os.WriteFile("VERSION", []byte(v.String()+"\n"), 0644)
}

type tplPayload struct {
	Major uint64
	Minor uint64
	Patch uint64
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
	return tmpl.Execute(fh, tplPayload{
		Major: v.Major,
		Minor: v.Minor,
		Patch: v.Patch,
	})
}

func isGitClean() bool {
	buf, err := exec.Command("git", "diff", "--stat").CombinedOutput()
	if err != nil {
		panic(err)
	}
	return strings.TrimSpace(string(buf)) == ""
}

func versionFile() (semver.Version, error) {
	buf, err := os.ReadFile("VERSION")
	if err != nil {
		return semver.Version{}, err
	}
	return semver.Parse(strings.TrimSpace(string(buf)))
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
	return semver.Parse(strings.TrimPrefix(lines[len(lines)-1], "v"))
}

func changelogEntries(since semver.Version) ([]string, error) {
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

	notes := make([]string, 0, 10)
	commits := strings.Split(string(buf), gitSep)
	for _, commit := range commits {
		commit := strings.TrimSpace(commit)
		if commit == "" {
			continue
		}
		p := strings.Split(commit, gitDelim)
		if len(p) < 3 {
			// invalid entry, shouldn't happen
			continue
		}

		issues := []string{}
		if m := issueRE.FindStringSubmatch(strings.TrimSpace(p[1])); len(m) > 1 {
			issues = append(issues, m[1])
		}

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
