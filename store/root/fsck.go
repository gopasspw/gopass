package root

import (
	"context"

	"github.com/justwatchcom/gopass/utils/out"
)

// Fsck checks the stores integrity
func (r *Store) Fsck(ctx context.Context, prefix string) (map[string]uint64, error) {
	rc := make(map[string]uint64, 10)
	sh := make(map[string]string, 100)
	for _, alias := range r.MountPoints() {
		// check sub-store integrity
		counts, err := r.mounts[alias].Fsck(ctx, alias)
		if err != nil {
			return rc, err
		}
		for k, v := range counts {
			rc[k] += v
		}

		out.Green(ctx, "[%s] Store (%s) checked (%d OK, %d warnings, %d errors)", alias, r.mounts[alias].Path(), counts["ok"], counts["warn"], counts["err"])

		// check shadowing
		lst, err := r.mounts[alias].List(alias)
		if err != nil {
			return rc, err
		}
		for _, e := range lst {
			if a, found := sh[e]; found {
				out.Yellow(ctx, "Entry %s is being shadowed by %s", e, a)
			}
			sh[e] = alias
		}
	}

	counts, err := r.store.Fsck(ctx, "root")
	if err != nil {
		return rc, err
	}
	for k, v := range counts {
		rc[k] += v
	}
	out.Green(ctx, "[%s] Store checked (%d OK, %d warnings, %d errors)", r.store.Path(), counts["ok"], counts["warn"], counts["err"])
	// check shadowing
	lst, err := r.store.List("")
	if err != nil {
		return rc, err
	}
	for _, e := range lst {
		if a, found := sh[e]; found {
			out.Yellow(ctx, "Entry %s is being shadowed by %s", e, a)
		}
		sh[e] = ""
	}
	return rc, nil
}
