package fossilfs

import (
	"context"
	"strings"

	"bitbucket.org/creachadair/stringset"
)

type fossilStatus struct {
	Extra     stringset.Set
	Added     stringset.Set
	Edited    stringset.Set
	Unchanged stringset.Set
}

func (f *Fossil) getStatus(ctx context.Context) (fossilStatus, error) {
	stdout, _, err := f.captureCmd(ctx, "fossilStatus", "status", "--extra", "--all")
	if err != nil {
		return fossilStatus{}, err
	}

	s := fossilStatus{
		Extra:     stringset.New(),
		Added:     stringset.New(),
		Edited:    stringset.New(),
		Unchanged: stringset.New(),
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

func (fs *fossilStatus) Untracked() stringset.Set {
	return fs.Extra.Union(fs.Added).Union(fs.Edited)
}

func (fs *fossilStatus) Staged() stringset.Set {
	return fs.Edited.Union(fs.Added)
}
