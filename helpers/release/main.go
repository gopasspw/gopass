package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"sort"
	"strings"
	"text/template"
	"time"

	"github.com/blang/semver/v4"
)

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
	// - calculate next version
	gitVer, err := gitVersion()
	if err != nil {
		panic(err)
	}
	vfVer, err := versionFile()
	if err != nil {
		panic(err)
	}

	if gitVer.NE(vfVer) {
		fmt.Printf("git version: %q != VERSION: %q", gitVer.String(), vfVer.String())
		panic("version mismatch")
	}
	fmt.Printf("✅ previous version is consistent (%s)\n", gitVer.String())

	nextVer := gitVer
	if len(os.Args) > 1 {
		nextVer = semver.MustParse(strings.TrimPrefix(os.Args[1], "v"))
		if nextVer.LTE(gitVer) {
			panic("next version must be greather than the previous version")
		}
	} else {
		nextVer.IncrementPatch()
	}

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
	time.Sleep(100 * time.Millisecond)
	// - update version.go
	if err := writeVersionGo(nextVer); err != nil {
		panic(err)
	}
	fmt.Println("✅ Wrote version.go")
	time.Sleep(100 * time.Millisecond)
	// - update CHANGELOG.md
	if err := writeChangelog(gitVer, nextVer); err != nil {
		panic(err)
	}
	fmt.Println("✅ Updated CHANGELOG.md")
	time.Sleep(100 * time.Millisecond)

	// - create PR
	//   git checkout -b release/vX.Y.Z
	if err := gitCoRel(nextVer); err != nil {
		panic(err)
	}
	fmt.Printf("✅ Created branch release/v%s\n", nextVer.String())
	time.Sleep(100 * time.Millisecond)

	// commit changes
	if err := gitCommit(nextVer); err != nil {
		panic(err)
	}
	fmt.Printf("✅ Committed changes to release/v%s\n", nextVer.String())
	time.Sleep(100 * time.Millisecond)

	fmt.Println("🏁 Done")
	fmt.Printf("⚠ Prepared release of gopass %s. Now push this branch to your fork and open a pull request against gopasspw/gopass master.\n", nextVer.String())
	fmt.Println()
	fmt.Printf("⚠ After that PR has been merged tag it as 'v%s' and push that tag to kick of the release proecess.\n", nextVer.String())
	fmt.Println()
}

func gitCoMaster() error {
	cmd := exec.Command("git", "checkout", "master")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func gitPom() error {
	cmd := exec.Command("git", "pull", "origin", "master")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func gitCoRel(v semver.Version) error {
	cmd := exec.Command("git", "checkout", "-b", "release/v"+v.String())
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func gitCommit(v semver.Version) error {
	cmd := exec.Command("git", "add", "CHANGELOG.md", "VERSION", "version.go")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}
	cmd = exec.Command("git", "commit", "-s", "-m", "Tag v"+v.String())
	cmd.Stdout = os.Stdout
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
	return ioutil.WriteFile("VERSION", []byte(v.String()), 0644)
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
	buf, err := ioutil.ReadFile("VERSION")
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
	buf, err := exec.Command("git", "log", "v"+since.String()+"..HEAD", "--pretty=full", "--grep=RELEASE_NOTES").CombinedOutput()
	if err != nil {
		return nil, err
	}
	notes := make([]string, 0, 10)
	lines := strings.Split(string(buf), "\n")
	for _, line := range lines {
		line := strings.TrimSpace(line)
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
		notes = append(notes, val)
	}

	sort.Strings(notes)
	return notes, nil
}
