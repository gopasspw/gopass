package ondisk

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/cenkalti/backoff"
	"github.com/gopasspw/gopass/internal/backend/storage/kv/ondisk/gjs"
	"github.com/gopasspw/gopass/internal/debug"
	"github.com/gopasspw/gopass/pkg/fsutil"
	"github.com/minio/minio-go/v6"
)

// downloadFiles fetchs all blobs from the remote
func (o *OnDisk) downloadFiles(ctx context.Context) error {
	if o.mio == nil || o.mbu == "" {
		debug.Log("remote not initialized")
		return nil
	}
	for _, blob := range o.idx.ListBlobs() {
		debug.Log("downloading %s ...", blob)
		if err := o.downloadFile(ctx, blob, false); err != nil {
			return err
		}
	}
	return nil
}

// downloadFile fetches a single file from the remote
func (o *OnDisk) downloadFile(ctx context.Context, name string, force bool) error {
	fp := filepath.Join(o.dir, name)
	if fsutil.IsFile(fp) && !force {
		debug.Log("file %s already exists", fp)
		return nil
	}
	obj, err := o.mio.GetObjectWithContext(ctx, o.mbu, o.remoteFn(name), minio.GetObjectOptions{})
	if err != nil {
		debug.Log("failed to stat %s: %s", name, err)
		return nil
	}
	fh, err := os.OpenFile(fp, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer fh.Close()

	stat, err := obj.Stat()
	if err != nil {
		return err
	}
	if _, err := io.CopyN(fh, obj, stat.Size); err != nil {
		return err
	}
	return nil
}

// downloadBlob fetches a single blob
func (o *OnDisk) downloadBlob(ctx context.Context, name string) ([]byte, error) {
	obj, err := o.mio.GetObjectWithContext(ctx, o.mbu, o.remoteFn(name), minio.GetObjectOptions{})
	if err != nil {
		debug.Log("failed to stat %s: %s", name, err)
		return nil, err
	}
	buf := &bytes.Buffer{}
	stat, err := obj.Stat()
	if err != nil {
		return nil, err
	}
	if _, err := io.CopyN(buf, obj, stat.Size); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// uploadFiles uploads all blobs
func (o *OnDisk) uploadFiles(ctx context.Context) error {
	if o.mio == nil || o.mbu == "" {
		debug.Log("remote not initialized")
		return nil
	}
	for _, blob := range o.idx.ListBlobs() {
		debug.Log("uploading %s ...", blob)
		if err := o.uploadFile(ctx, blob, false); err != nil {
			return err
		}
	}
	return nil
}

// uploadFile uploads a single file
func (o *OnDisk) uploadFile(ctx context.Context, name string, force bool) error {
	fp := filepath.Join(o.dir, name)
	if !fsutil.IsFile(fp) {
		debug.Log("file %s doesn't exist (this should not happen, please report a bug)", fp)
		return nil
	}
	if !force {
		_, err := o.mio.StatObjectWithContext(ctx, o.mbu, o.remoteFn(name), minio.StatObjectOptions{})
		if err == nil {
			// TODO also compare sizes
			debug.Log("file %s already exists on the remote", fp)
			return nil
		}
	}
	fh, err := os.Open(fp)
	if err != nil {
		return err
	}
	defer fh.Close()
	fi, err := fh.Stat()
	if err != nil {
		return err
	}
	n, err := o.mio.PutObjectWithContext(ctx, o.mbu, o.remoteFn(name), fh, fi.Size(), minio.PutObjectOptions{ContentType: "application/octet-stream"})
	if err != nil {
		return err
	}
	debug.Log("Uploaded %d bytes to %s/%s", n, o.mbu, name)
	return nil
}

// uploadBlob uploads a single blob
func (o *OnDisk) uploadBlob(ctx context.Context, name string, buf []byte) error {
	n, err := o.mio.PutObjectWithContext(ctx, o.mbu, o.remoteFn(name), bytes.NewReader(buf), int64(len(buf)), minio.PutObjectOptions{ContentType: "text/plain"})
	if err != nil {
		return err
	}
	debug.Log("Uploaded %d bytes to %s/%s", n, o.mbu, name)
	return nil
}

func (o *OnDisk) remoteFn(name string) string {
	if o.mpf == "" {
		return name
	}
	return path.Join(o.mpf, name)
}

// isNotFound unwraps the error response and checks for a not found status
func isNotFound(err error) bool {
	e, ok := err.(minio.ErrorResponse)
	if !ok {
		return false
	}
	return e.StatusCode == 404
}

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
	buf, err := o.downloadBlob(ctx, idxFile)
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
	// upload local index
	err = o.saveIndexToDisk(ctx)
	if err != nil {
		return err
	}
	debug.Log("saved index after %s", time.Since(t0))
	if err := o.uploadFile(ctx, idxFile, true); err != nil {
		return err
	}
	debug.Log("uploaded index after %s", time.Since(t0))

	return nil
}
