package fs

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/gopasspw/gopass/pkg/debug"
)

// addRel adds the required number of relative elements to go from dst back to
// src.
func addRel(src, dst string) string {
	for i := 0; i < strings.Count(dst, "/"); i++ {
		src = "../" + src
	}
	return src
}

// longestCommonPrefix finds the longest common prefix directory.
func longestCommonPrefix(l, r string) string {
	var prefix string
	for i := 0; i < len(l) && i < len(r); i++ {
		if l[i] != r[i] {
			prefix = l[:i]
			break
		}
	}
	if !strings.Contains(prefix, "/") {
		return prefix
	}
	return prefix[:strings.LastIndex(prefix, "/")]
}

// Link creates a symlink, i.e. an alias to reach the same secret
// through different names.
func (s *Store) Link(ctx context.Context, from, to string) error {
	if runtime.GOOS == "windows" {
		from = filepath.FromSlash(from)
		to = filepath.FromSlash(to)
	}
	fromPath := filepath.Join(s.path, from)
	toPath := filepath.Join(s.path, to)
	prefix := longestCommonPrefix(fromPath, toPath)

	fromRel := strings.TrimPrefix(fromPath, prefix+string(filepath.Separator))
	toRel := strings.TrimPrefix(toPath, prefix+string(filepath.Separator))

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("can not get current working directory: %w", err)
	}
	defer func() {
		os.Chdir(cwd)
	}()

	toDir := filepath.Dir(toPath)
	if err := os.MkdirAll(toDir, 0o700); err != nil {
		return fmt.Errorf("failed to create destination dir %q: %w", toDir, err)
	}

	if err := os.Chdir(toDir); err != nil {
		return fmt.Errorf("can no change to link dir %q: %w", toDir, err)
	}

	linkDst := addRel(fromRel, toRel)

	debug.Log("path: %q\n\tfromPath:\t%q\n\ttoPath:\t\t%q\n\tprefix:\t\t%q\n\tfromRel:\t%q\n\ttoRel:\t\t%q\n\ttoDir:\t\t%q\n\tlinkDst:\t%q",
		s.path, fromPath, toPath, prefix, fromRel, toRel, toDir, linkDst)
	return os.Symlink(linkDst, filepath.Base(toRel))
}
