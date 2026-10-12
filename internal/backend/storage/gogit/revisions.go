package gogit

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/gopasspw/gopass/internal/backend"
)

// Revisions lists all available revisions of the named entity.
func (g *GoGit) Revisions(_ context.Context, name string) ([]backend.Revision, error) {
	head, err := g.repo.Head()
	if err != nil {
		// Unborn HEAD: no history yet.
		return nil, nil
	}

	rel := filepath.ToSlash(strings.TrimPrefix(name, g.fs.Path()+"/"))
	rel = strings.TrimPrefix(rel, "./")

	// go-git's LogOptions.PathFilter is unreliable for merge commits: it
	// compares each commit with the next one in iteration order rather than
	// with its parent, so it can list commits that never touched the path.
	// Walk the full log and compare each commit against its first parent.
	iter, err := g.repo.Log(&git.LogOptions{
		From:  head.Hash(),
		Order: git.LogOrderCommitterTime,
	})
	if err != nil {
		return nil, err
	}
	defer iter.Close()

	revs := make([]backend.Revision, 0, 16)
	err = iter.ForEach(func(c *object.Commit) error {
		touched, err := commitTouchesPath(c, rel)
		if err != nil {
			return err
		}
		if !touched {
			return nil
		}

		rev := backend.Revision{
			Hash:        c.Hash.String(),
			AuthorName:  c.Author.Name,
			AuthorEmail: c.Author.Email,
			Date:        c.Author.When,
		}
		subject, body := splitCommitMessage(c.Message)
		rev.Subject = subject
		rev.Body = body
		revs = append(revs, rev)

		return nil
	})
	if err != nil {
		return nil, err
	}

	return revs, nil
}

// commitTouchesPath reports whether the commit changed the given path relative
// to all of its parents. A commit that is TREESAME to any parent (the path has
// the same blob as in that parent) is dropped, matching git's history
// simplification: a merge commit whose remote side changed the path is not
// listed alongside the real remote commit. For a root commit the path counts as
// touched when it exists at all.
func commitTouchesPath(c *object.Commit, rel string) (bool, error) {
	tree, err := c.Tree()
	if err != nil {
		return false, err
	}
	cur, err := entryHash(tree, rel)
	if err != nil {
		return false, err
	}

	if c.NumParents() == 0 {
		return cur != plumbing.ZeroHash, nil
	}

	for i := range c.NumParents() {
		parent, err := c.Parent(i)
		if err != nil {
			return false, err
		}
		parentTree, err := parent.Tree()
		if err != nil {
			return false, err
		}
		prev, err := entryHash(parentTree, rel)
		if err != nil {
			return false, err
		}
		if cur == prev {
			// TREESAME to this parent -> not a change introduced here.
			return false, nil
		}
	}

	return true, nil
}

// entryHash returns the blob hash of rel in tree, or the zero hash when the
// path does not exist in that tree.
func entryHash(tree *object.Tree, rel string) (plumbing.Hash, error) {
	e, err := tree.FindEntry(rel)
	if err != nil {
		if errors.Is(err, object.ErrEntryNotFound) || errors.Is(err, object.ErrFileNotFound) || errors.Is(err, object.ErrDirectoryNotFound) {
			return plumbing.ZeroHash, nil
		}

		return plumbing.ZeroHash, err
	}

	return e.Hash, nil
}

// GetRevision returns the content of any revision of the named entity.
func (g *GoGit) GetRevision(_ context.Context, name, revision string) ([]byte, error) {
	name = strings.TrimSpace(name)
	revision = strings.TrimSpace(revision)

	rel := filepath.ToSlash(strings.TrimPrefix(name, g.fs.Path()+"/"))
	rel = strings.TrimPrefix(rel, "./")

	hash, err := g.resolveRevision(revision)
	if err != nil {
		return nil, err
	}

	commit, err := g.repo.CommitObject(hash)
	if err != nil {
		return nil, err
	}
	file, err := commit.File(rel)
	if err != nil {
		return nil, fmt.Errorf("failed to get %q at %q: %w", rel, revision, err)
	}
	content, err := file.Contents()
	if err != nil {
		return nil, err
	}

	return []byte(content), nil
}

// resolveRevision resolves a revision string the same way the fs backend treats
// it: "latest" and "" mean HEAD.
func (g *GoGit) resolveRevision(revision string) (plumbing.Hash, error) {
	switch revision {
	case "", "latest":
		head, err := g.repo.Head()
		if err != nil {
			return plumbing.ZeroHash, err
		}

		return head.Hash(), nil
	default:
		hash, err := g.repo.ResolveRevision(plumbing.Revision(revision))
		if err != nil {
			return plumbing.ZeroHash, fmt.Errorf("failed to resolve revision %q: %w", revision, err)
		}

		return *hash, nil
	}
}

// splitCommitMessage splits a commit message into its subject line and body.
func splitCommitMessage(msg string) (string, string) {
	msg = strings.TrimRight(msg, "\n")
	subject, body, found := strings.Cut(msg, "\n")
	if !found {
		return subject, ""
	}

	return subject, strings.TrimSpace(body)
}
