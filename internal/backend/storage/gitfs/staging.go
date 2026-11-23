package gitfs

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/gopasspw/gopass/internal/backend"
)

type fileVersion struct {
	mode string
	hash string
}

func (g *Git) headRef(ctx context.Context) (string, error) {
	out, stderr, err := g.captureCmd(ctx, "gitHead", "rev-parse", "--verify", "--quiet", "HEAD")
	if err != nil {
		if ee, ok := errors.AsType[*exec.ExitError](err); ok && ee.ExitCode() == 1 {
			return "", nil // an unborn branch has no files in HEAD
		}

		return "", fmt.Errorf("git rev-parse HEAD: %w: %s", err, strings.TrimSpace(string(stderr)))
	}

	return strings.TrimSpace(string(out)), nil
}

func (g *Git) fileAtHead(ctx context.Context, revision, name string) (fileVersion, error) {
	if revision == "" {
		return fileVersion{}, nil
	}
	out, stderr, err := g.captureCmd(ctx, "gitListHead", "--literal-pathspecs", "ls-tree", "-z", revision, "--", name)
	if err != nil {
		return fileVersion{}, fmt.Errorf("git ls-tree: %w: %s", err, strings.TrimSpace(string(stderr)))
	}
	if len(out) == 0 {
		return fileVersion{}, nil
	}
	meta, path, _ := strings.Cut(string(out), "\t")
	fields := strings.Fields(meta)
	if path != name+"\x00" || len(fields) != 3 || fields[1] != "blob" {
		return fileVersion{}, backend.ConflictingIndexError(name)
	}

	return fileVersion{mode: fields[0], hash: fields[2]}, nil
}

func (g *Git) fileInIndex(ctx context.Context, name string) (fileVersion, error) {
	out, stderr, err := g.captureCmd(ctx, "gitListIndex", "--literal-pathspecs", "ls-files", "-v", "--stage", "-z", "--", name)
	if err != nil {
		return fileVersion{}, fmt.Errorf("git ls-files: %w: %s", err, strings.TrimSpace(string(stderr)))
	}
	if len(out) == 0 {
		return fileVersion{}, nil
	}
	meta, path, _ := strings.Cut(string(out), "\t")
	fields := strings.Fields(meta)
	// Only a normal stage-zero entry is safe. In particular, reject
	// skip-worktree and assume-unchanged, even if status looks clean.
	if path != name+"\x00" || len(fields) != 4 || fields[0] != "H" || fields[3] != "0" {
		return fileVersion{}, backend.ConflictingIndexError(name)
	}

	return fileVersion{mode: fields[1], hash: fields[2]}, nil
}

// CheckUnstaged protects pre-existing index entries before an import writes.
func (g *Git) CheckUnstaged(ctx context.Context, names ...string) error {
	if !g.IsInitialized() {
		return nil
	}
	revision, err := g.headRef(ctx)
	if err != nil {
		return err
	}
	for _, name := range names {
		head, err := g.fileAtHead(ctx, revision, name)
		if err != nil {
			return err
		}
		staged, err := g.fileInIndex(ctx, name)
		if err != nil {
			return err
		}
		if staged != head {
			return backend.ConflictingIndexError(name)
		}
	}

	return nil
}

func (g *Git) hashFile(ctx context.Context, data []byte, write bool) (string, error) {
	args := []string{"hash-object", "--no-filters", "--stdin"}
	if write {
		args = append(args, "-w")
	}
	out, stderr, err := g.captureCmdInput(ctx, "gitHashFile", data, args...)
	if err != nil {
		return "", fmt.Errorf("git hash-object: %w: %s", err, strings.TrimSpace(string(stderr)))
	}

	return strings.TrimSpace(string(out)), nil
}

func (g *Git) worktreeFile(ctx context.Context, file backend.StagingFile, staged fileVersion, fileMode bool) (fileVersion, []byte, error) {
	info, err := os.Lstat(filepath.Join(g.fs.Path(), file.Name))
	if err != nil {
		return fileVersion{}, nil, err
	}
	if !info.Mode().IsRegular() {
		return fileVersion{}, nil, backend.ConflictingIndexError(file.Name)
	}
	data, err := g.fs.Get(ctx, file.Name)
	if err != nil {
		return fileVersion{}, nil, err
	}
	if file.Expected != nil && !bytes.Equal(file.Expected, data) {
		return fileVersion{}, nil, fmt.Errorf("worktree file %q changed during import recovery", file.Name)
	}
	hash, err := g.hashFile(ctx, data, false)
	if err != nil {
		return fileVersion{}, nil, err
	}
	mode := "100644"
	if fileMode && info.Mode()&0o100 != 0 || !fileMode && staged.mode == "100755" {
		mode = "100755"
	}

	return fileVersion{mode: mode, hash: hash}, data, nil
}

// PlanStaging compares the actual Git blobs and takes a worktree snapshot.
// A successful plan stages only the checked bytes, without Git clean filters.
func (g *Git) PlanStaging(ctx context.Context, files ...backend.StagingFile) (backend.StagingPlan, error) {
	plan := backend.StagingPlan{}
	if !g.IsInitialized() {
		return plan, nil
	}
	revision, err := g.headRef(ctx)
	if err != nil {
		return plan, err
	}
	cfg, err := g.ConfigGet(ctx, "core.filemode")
	if err != nil {
		return plan, err
	}
	var add []stagingSnapshot
	for _, file := range files {
		head, err := g.fileAtHead(ctx, revision, file.Name)
		if err != nil {
			return plan, err
		}
		staged, err := g.fileInIndex(ctx, file.Name)
		if err != nil {
			return plan, err
		}
		working, data, err := g.worktreeFile(ctx, file, staged, cfg != "false")
		if err != nil {
			return plan, err
		}
		if staged != head && staged != working {
			return plan, backend.ConflictingIndexError(file.Name)
		}
		if working != head {
			plan.Pending = true
		}
		if working != staged {
			add = append(add, stagingSnapshot{name: file.Name, version: working, index: staged, data: data})
		}
	}
	if len(add) > 0 {
		plan.Apply = func(ctx context.Context) error { return g.stageSnapshots(ctx, add) }
	}

	return plan, nil
}

type stagingSnapshot struct {
	name    string
	version fileVersion
	index   fileVersion
	data    []byte
}

func (g *Git) stageSnapshots(ctx context.Context, add []stagingSnapshot) error {
	var input bytes.Buffer
	for _, f := range add {
		version, err := g.fileInIndex(ctx, f.name)
		if err != nil {
			return err
		}
		if version != f.index {
			return backend.ConflictingIndexError(f.name)
		}
		hash, err := g.hashFile(ctx, f.data, true)
		if err != nil {
			return err
		}
		if hash != f.version.hash {
			return fmt.Errorf("git object hash changed for %q", f.name)
		}
		fmt.Fprintf(&input, "%s %s\t%s\x00", f.version.mode, hash, f.name)
	}
	_, stderr, err := g.captureCmdInput(ctx, "gitStageSnapshots", input.Bytes(), "update-index", "-z", "--index-info")
	if err != nil {
		return fmt.Errorf("git update-index: %w: %s", err, strings.TrimSpace(string(stderr)))
	}

	return nil
}

// StageUncommitted stages only verified pending paths.
func (g *Git) StageUncommitted(ctx context.Context, names ...string) (bool, error) {
	plan, err := g.PlanStaging(ctx, backend.StagingFiles(names...)...)
	if err != nil {
		return false, err
	}

	return plan.Stage(ctx)
}
