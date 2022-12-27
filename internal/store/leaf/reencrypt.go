package leaf

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"

	"github.com/gopasspw/gopass/internal/config"
	"github.com/gopasspw/gopass/internal/out"
	"github.com/gopasspw/gopass/internal/store"
	"github.com/gopasspw/gopass/pkg/ctxutil"
	"github.com/gopasspw/gopass/pkg/debug"
	"github.com/gopasspw/gopass/pkg/termio"
)

// nolint:ifshort
// reencrypt will re-encrypt all entries for the current recipients.
func (s *Store) reencrypt(ctx context.Context) error {
	entries, err := s.List(ctx, "")
	if err != nil {
		return fmt.Errorf("failed to list store: %w", err)
	}

	// Most gnupg setups don't work well with concurrency > 1, but
	// for other backends - e.g. age - this could very well be > 1.
	conc := s.crypto.Concurrency()

	// save original value of auto push
	{
		// shadow ctx in this block only
		ctx := ctxutil.WithGitCommit(ctx, false)

		// progress bar
		bar := termio.NewProgressBar(int64(len(entries)))
		bar.Hidden = !ctxutil.IsTerminal(ctx) || ctxutil.IsHidden(ctx)

		var wg sync.WaitGroup
		jobs := make(chan string)
		// We use a logger to write without race condition on stdout
		logger := log.New(os.Stdout, "", 0)
		out.Printf(ctx, "Starting reencrypt")

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

				return fmt.Errorf("context canceled")
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
	if conc > 1 {
		for _, name := range entries {
			p := s.Passfile(name)
			if err := s.storage.Add(ctx, p); err != nil {
				if errors.Is(err, store.ErrGitNotInit) {
					debug.Log("skipping git add - git not initialized")

					continue
				}

				return fmt.Errorf("failed to add %q to git: %w", p, err)
			}

			debug.Log("added %s to git", p)
		}
	}

	if err := s.storage.Commit(ctx, ctxutil.GetCommitMessage(ctx)); err != nil {
		switch {
		case errors.Is(err, store.ErrGitNotInit):
			debug.Log("skipping git commit - git not initialized")
		case errors.Is(err, store.ErrGitNothingToCommit):
			debug.Log("skipping git commit - nothing to commit")
		default:
			return fmt.Errorf("failed to commit changes to git: %w", err)
		}
	}

	return s.reencryptGitPush(ctx)
}

func (s *Store) reencryptGitPush(ctx context.Context) error {
	if !config.Bool(ctx, "core.autosync") {
		debug.Log("not pushing to git remote, core.autosync is false")

		return nil
	}

	if err := s.storage.Push(ctx, "", ""); err != nil {
		if errors.Is(err, store.ErrGitNotInit) {
			msg := "Warning: git is not initialized for this.storage. Ignoring auto-push option\n" +
				"Run: gopass git init"
			debug.Log(msg)

			return nil
		}

		if errors.Is(err, store.ErrGitNoRemote) {
			msg := "Warning: git has no remote. Ignoring auto-push option\n" +
				"Run: gopass git remote add origin ..."
			debug.Log(msg)

			return nil
		}

		return fmt.Errorf("failed to push change to git remote: %w", err)
	}

	return nil
}
