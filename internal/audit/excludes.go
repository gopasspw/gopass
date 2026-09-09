package audit

import (
	"context"
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/gopasspw/gopass/internal/store"
	"github.com/gopasspw/gopass/internal/store/root"
	"github.com/gopasspw/gopass/internal/tree"
	"github.com/gopasspw/gopass/pkg/debug"
)

type res []*regexp.Regexp

func (r res) Matches(s string) bool {
	for _, re := range r {
		if re.MatchString(s) {
			debug.Log("Matched %s against %s", s, re.String())

			return true
		}
	}

	return false
}

// FilteredList returns a list of all secrets in the given store, filtered against the .gopass-audit-ignore file in each mount point.
func FilteredList(ctx context.Context, rs *root.Store) ([]string, error) {
	t, err := rs.Tree(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get store tree: %w", err)
	}

	list := t.List(tree.INF)
	if len(list) < 1 {
		return list, nil
	}

	// Collect the exclude patterns for every mount point (including the root store).
	// The longer a mount point the more specific it is, so we sort them by descending
	// length to make sure the most specific mount point wins for every secret.
	mps := rs.MountPoints()
	sort.Sort(sort.Reverse(store.ByPathLen(mps)))

	excludes := make(map[string]string, len(mps)+1)
	for _, mp := range mps {
		excludes[mp] = loadExcludes(ctx, rs, mp)
	}
	// the root store is the fallback for secrets not in any mount
	excludes[""] = loadExcludes(ctx, rs, "")

	// Group the secrets by their mount point, so that each excludes file
	// only needs to be parsed once.
	byMount := make(map[string][]string, len(mps)+1)
	for _, name := range list {
		mp := rs.MountPoint(name)
		byMount[mp] = append(byMount[mp], name)
	}

	out := make([]string, 0, len(list))
	for mp, secrets := range byMount {
		out = append(out, FilterExcludes(excludes[mp], secrets)...)
	}
	sort.Strings(out)

	return out, nil
}

// loadExcludes returns the content of the .gopass-audit-ignore file at the
// root of the given mount point, if any.
func loadExcludes(ctx context.Context, rs *root.Store, mp string) string {
	st := rs.Storage(ctx, mp)
	if st == nil {
		return ""
	}
	buf, err := st.Get(ctx, ".gopass-audit-ignore")
	if err != nil || buf == nil {
		return ""
	}

	return string(buf)
}

// FilterExcludes filters the given list of secrets against the given exclude patterns (RE2 syntax).
func FilterExcludes(excludes string, in []string) []string {
	debug.Log("Filtering %d secrets against %d exclude patterns", len(in), strings.Count(excludes, "\n"))

	res := make(res, 0, 10)
	for line := range strings.SplitSeq(excludes, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "#") {
			continue
		}
		re, err := regexp.Compile(line)
		if err != nil {
			debug.Log("failed to compile exclude pattern %q: %s", line, err)

			continue
		}
		debug.Log("Adding exclude pattern %q", re.String())
		res = append(res, re)
	}

	// shortcut if we have no excludes
	if len(res) < 1 {
		return in
	}

	// check all secrets against all excludes
	out := make([]string, 0, len(in))
	for _, s := range in {
		if res.Matches(s) {
			continue
		}
		out = append(out, s)
	}

	return out
}
