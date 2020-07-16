package ondisk

import (
	"context"
	"fmt"
	"time"

	"github.com/cenkalti/backoff"
	"github.com/gopasspw/gopass/internal/backend/crypto/age"
	"github.com/gopasspw/gopass/internal/backend/storage/kv/ondisk/gjs"
	"github.com/gopasspw/gopass/internal/debug"
	"github.com/gopasspw/gopass/internal/recipients"
	"github.com/minio/minio-go/v6"
)

// awaitLock waits up to 10s to acquire the lock on the remote
func (o *OnDisk) awaitLock(ctx context.Context) error {
	bo := backoff.NewExponentialBackOff()
	bo.MaxElapsedTime = 10 * time.Second
	boc := backoff.WithContext(bo, ctx)
	debug.Log("waiting for sync lock up to %s", bo.MaxElapsedTime)
	return backoff.RetryNotify(func() error {
		oi, err := o.mio.StatObjectWithContext(ctx, o.mbu, idxLockFile, minio.StatObjectOptions{})
		if err != nil {
			if isNotFound(err) {
				return nil
			}
			debug.Log("lock not found: %s", err)
			return err
		}
		if oi.Size > 0 {
			return fmt.Errorf("locked")
		}
		return nil
	}, boc, func(err error, ts time.Duration) {
		debug.Log("still waiting for lock after %s: %s", ts, err)
	})
}

// removeLock tries to remove the lock from the remote
func (o *OnDisk) removeLock() error {
	err := o.mio.RemoveObject(o.mbu, idxLockFile)
	if err != nil {
		debug.Log("Failed to remove lock %s: %s", idxLockFile, err)
	}
	return err
}

// syncIndex attempts to sync the index with the remote
func (o *OnDisk) syncIndex(ctx context.Context) error {
	if o.mio == nil || o.mbu == "" {
		debug.Log("remote not initialized")
		return nil
	}
	bo := backoff.WithMaxRetries(backoff.NewConstantBackOff(time.Second), 3)
	boc := backoff.WithContext(bo, ctx)
	debug.Log("trying to sync index ...")
	return backoff.RetryNotify(func() error {
		return o.trySyncIndex(ctx)
	}, boc, func(err error, ts time.Duration) {
		debug.Log("still trying to sync after %s: %s", ts, err)
	})
}

// trySyncIndex tries to sync the index once (locking retries)
func (o *OnDisk) trySyncIndex(ctx context.Context) error {
	t0 := time.Now().UTC()
	// check lock file, retry if found
	if err := o.awaitLock(ctx); err != nil {
		return err
	}
	debug.Log("acquired lock after %s", time.Since(t0))
	// create lock file
	if err := o.uploadBlob(ctx, idxLockFile, []byte("foobar")); err != nil {
		return err
	}
	debug.Log("created lock after %s", time.Since(t0))
	// remove lock file when we're done
	defer o.removeLock()
	// download remote index
	buf, err := o.downloadBlob(ctx, idxFileRemote)
	// not found is not an error, if the remote has no index
	// we'll just upload ours
	if err != nil && !isNotFound(err) {
		return err
	}
	idxRemote := &gjs.Store{}
	debug.Log("fetched remote index after %s", time.Since(t0))
	// decrypt
	if !isNotFound(err) {
		idxRemote, err = o.loadIndex(ctx, buf)
		if err != nil {
			return err
		}
		debug.Log("decrypted remote index after %s", time.Since(t0))
	}
	// merge
	debug.Log("merging local index (%d blobs) with remote index (%d blobs)", len(o.idx.ListBlobs()), len(idxRemote.ListBlobs()))
	o.idx.Merge(idxRemote)
	debug.Log("merged indices after %s (%d blobs)", time.Since(t0), len(o.idx.ListBlobs()))
	// updated local index
	if err := o.saveIndexToDisk(ctx); err != nil {
		return err
	}
	debug.Log("saved index after %s", time.Since(t0))
	// encrypt remote index for all current recipients of the store
	aIds, err := o.Get(ctx, age.IDFile)
	if err != nil {
		return err
	}
	rs := recipients.Unmarshal(aIds)
	buf, err = o.saveIndex(ctx, rs...)
	if err != nil {
		return err
	}
	debug.Log("remote index encrypted for %+v after %s", rs, time.Since(t0))
	// update remote index
	if err := o.uploadBlob(ctx, idxFileRemote, buf); err != nil {
		return err
	}
	debug.Log("uploaded index after %s", time.Since(t0))

	return nil
}
