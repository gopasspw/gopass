package fossilfs

import (
	"context"
	"strings"

	"github.com/gopasspw/gopass/pkg/set"
)

type fossilStatus struct {
	Extra     set.Set[string]
	Added     set.Set[string]
	Edited    set.Set[string]
	Unchanged set.Set[string]
}

func (f *Fossil) getStatus(ctx context.Context) (fossilStatus, error) {
	stdout, _, err := f.captureCmd(ctx, "fossilStatus", "status", "--extra", "--all")
	if err != nil {
		return fossilStatus{}, err
	}

	s := fossilStatus{
		Extra:     set.New[string](),
		Added:     set.New[string](),
		Edited:    set.New[string](),
		Unchanged: set.New[string](),
	}
	for _, line := range strings.Split(string(stdout), "\n") {
		op, file, found := strings.Cut(line, " ")
		if !found {
			continue
		}
		switch op {
		case "ADDED":
			s.Added.Add(strings.TrimSpace(file))
		case "UNCHANGED":
			s.Unchanged.Add(strings.TrimSpace(file))
		case "EXTRA":
			s.Added.Add(strings.TrimSpace(file))
		case "EDITED":
			s.Edited.Add(strings.TrimSpace(file))
		}
	}

	return s, nil
}

func (fs *fossilStatus) Untracked() set.Set[string] {
	return fs.Extra.Union(fs.Added).Union(fs.Edited)
}

func (fs *fossilStatus) Staged() set.Set[string] {
	return fs.Edited.Union(fs.Added)
}
