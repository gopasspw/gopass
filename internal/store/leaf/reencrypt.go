package leaf

import (
	"context"
	"log"
	"os"
	"strings"
	"sync"

	"github.com/gopasspw/gopass/internal/debug"
	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/internal/store"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/pkg/errors"
)

// reencrypt will re-encrypt all entries for the current recipients
func (s *Store) reencrypt(ctx context.Context) error {
	entries, err := s.List(ctx, "")
	if err != nil {
		return errors.Wrapf(err, "failed to list store")
	}

	// save original value of auto push
	{
		// shadow ctx in this block only
		ctx := ctxutil.WithGitCommit(ctx, false)

		// progress bar
		bar := out.NewProgressBar(ctx, int64(len(entries)))
		if !ctxutil.IsTerminal(ctx) || out.IsHidden(ctx) {
			bar = nil
		}
		var wg sync.WaitGroup
		jobs := make(chan string)
		// We use a logger to write without race condition on stdout
		logger := log.New(os.Stdout, "", 0)
		out.Print(ctx, "Starting reencrypt")
		// We spawn as many workers as we have set in the concurrency setting
		// GetConcurrency will return 1 if the concurrency setting is not set
		// or if it set to a value below 1.
		conc := ctxutil.GetConcurrency(ctx)
		for i := 0; i < conc; i++ {
			wg.Add(1) // we start a new job
			go func(workerId int) {
				// the workers are fed through an unbuffered channel
				for e := range jobs {
					content, err := s.Get(ctx, e)
					if err != nil {
						logger.Printf("Worker %d: Failed to get current value for %s: %s\n", workerId, e, err)
						continue
					}
					if err := s.Set(WithNoGitOps(ctx, conc > 1), e, content); err != nil {
						logger.Printf("Worker %d: Failed to write %s: %s\n", workerId, e, err)
						continue
					}
				}
				wg.Done() // report the job as finished
			}(i)
		}
		for _, e := range entries {
			// check for context cancelation
			select {
			case <-ctx.Done():
				// We close the channel, so the worker will terminate
				close(jobs)
				// we wait for all workers to have finished
				wg.Wait()
				return errors.New("context canceled")
			default:
			}

			if bar != nil {
				bar.Inc()
			}

			e = strings.TrimPrefix(e, s.alias)
			jobs <- e
		}
		// We close the channel, so the workers will terminate
		close(jobs)
		// we wait for all workers to have finished
		wg.Wait()
		bar.Done()
	}

	// if we were working concurrently, we couldn't git add during the process
	// to avoid a race condition on git .index.lock file, so we do it now.
	if ctxutil.HasConcurrency(ctx) {
		for _, name := range entries {
			p := s.passfile(name)
			if err := s.storage.Add(ctx, p); err != nil {
				switch errors.Cause(err) {
				case store.ErrGitNotInit:
					debug.Log("skipping git add - git not initialized")
					continue
				default:
					return errors.Wrapf(err, "failed to add '%s' to git", p)
				}
			}
			debug.Log("added %s to git", p)
		}
	}

	if err := s.storage.Commit(ctx, ctxutil.GetCommitMessage(ctx)); err != nil {
		switch errors.Cause(err) {
		case store.ErrGitNotInit:
			debug.Log("skipping git commit - git not initialized")
		case store.ErrGitNothingToCommit:
			debug.Log("skipping git commit - nothing to commit")
		default:
			return errors.Wrapf(err, "failed to commit changes to git")
		}
	}

	return s.reencryptGitPush(ctx)
}

func (s *Store) reencryptGitPush(ctx context.Context) error {
	if err := s.storage.Push(ctx, "", ""); err != nil {
		if errors.Cause(err) == store.ErrGitNotInit {
			msg := "Warning: git is not initialized for this.storage. Ignoring auto-push option\n" +
				"Run: gopass git init"
			out.Error(ctx, msg)
			return nil
		}
		if errors.Cause(err) == store.ErrGitNoRemote {
			msg := "Warning: git has no remote. Ignoring auto-push option\n" +
				"Run: gopass git remote add origin ..."
			out.Yellow(ctx, msg)
			return nil
		}
		return errors.Wrapf(err, "failed to push change to git remote")
	}
	return nil
}
