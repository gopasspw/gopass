package ondisk

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/gopasspw/gopass/internal/debug"
	"github.com/gopasspw/gopass/pkg/fsutil"
	"github.com/minio/minio-go/v7"
)

// RemoteConfig is a remote config
type RemoteConfig struct {
	SSL    bool   `json:"ssl"`
	KeyID  string `json:"key"`
	Secret string `json:"secret"`
	Host   string `json:"host"`
	Bucket string `json:"bucket"`
	Prefix string `json:"prefix"`
}

// String implements fmt.Stringer
func (r *RemoteConfig) String() string {
	if r == nil {
		return "<empty>"
	}
	return fmt.Sprintf("Host: %s, SSL: %t, Key: %s, Bucket: %s, Prefix: %s, Secret: <omitted>", r.Host, r.SSL, r.KeyID, r.Bucket, r.Prefix)
}

// SetRemote updates the remote config for this store
func (o *OnDisk) SetRemote(ctx context.Context, urlStr string) error {
	if urlStr == "" {
		debug.Log("removing remote config")
		return o.saveRemoteConfig(ctx, &RemoteConfig{})
	}
	u, err := url.Parse(urlStr)
	if err != nil {
		return err
	}
	keyid := ""
	secret := ""
	if u.User != nil {
		keyid = u.User.Username()
		secret, _ = u.User.Password()
	}
	bucket := ""
	prefix := ""
	p := strings.Split(u.Path, "/")
	if len(p) < 1 {
		return fmt.Errorf("need bucket")
	}
	bucket = p[0]
	if len(p) > 1 {
		prefix = p[1]
	}
	cfg := &RemoteConfig{
		SSL:    u.Scheme == "https",
		KeyID:  keyid,
		Secret: secret,
		Host:   u.Host,
		Bucket: bucket,
		Prefix: prefix,
	}
	return o.saveRemoteConfig(ctx, cfg)
}

// GetRemote reads the remote config from disk
func (o *OnDisk) GetRemote(ctx context.Context) (*RemoteConfig, error) {
	return o.loadRemoteConfig(ctx)
}

func (o *OnDisk) loadRemoteConfig(ctx context.Context) (*RemoteConfig, error) {
	path := filepath.Join(o.dir, cfgRemote)
	buf, err := ioutil.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			debug.Log("no remote config found")
			return &RemoteConfig{}, nil
		}
		return nil, err
	}
	debug.Log("loading remote config from %s", path)
	plain, err := o.age.Decrypt(ctx, buf)
	if err != nil {
		return nil, err
	}
	debug.Log("JSON: %s", string(plain))
	cfg := &RemoteConfig{}
	err = json.Unmarshal(plain, cfg)
	return cfg, err
}

func (o *OnDisk) saveRemoteConfig(ctx context.Context, cfg *RemoteConfig) error {
	plain, err := json.Marshal(cfg)
	if err != nil {
		return err
	}
	buf, err := o.age.Encrypt(ctx, plain, []string{}) // only encrypt for ourself
	if err != nil {
		return err
	}
	fn := filepath.Join(o.dir, cfgRemote)
	debug.Log("saving remote config to %s (%d bytes)", fn, len(buf))
	return ioutil.WriteFile(fn, buf, 0600)
}

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
	obj, err := o.mio.GetObject(ctx, o.mbu, o.remoteFn(name), minio.GetObjectOptions{})
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
	obj, err := o.mio.GetObject(ctx, o.mbu, o.remoteFn(name), minio.GetObjectOptions{})
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
	fh, err := os.Open(fp)
	if err != nil {
		return err
	}
	defer fh.Close()
	fi, err := fh.Stat()
	if err != nil {
		return err
	}
	if !force {
		stat, err := o.mio.StatObject(ctx, o.mbu, o.remoteFn(name), minio.StatObjectOptions{})
		if err == nil && stat.Size == fi.Size() {
			debug.Log("file %s already exists on the remote with the same size", fp, stat.Size)
			return nil
		}
	}
	n, err := o.mio.PutObject(ctx, o.mbu, o.remoteFn(name), fh, fi.Size(), minio.PutObjectOptions{ContentType: "application/octet-stream"})
	if err != nil {
		return err
	}
	debug.Log("Uploaded %d bytes to %s/%s", n, o.mbu, name)
	return nil
}

// uploadBlob uploads a single blob
func (o *OnDisk) uploadBlob(ctx context.Context, name string, buf []byte) error {
	n, err := o.mio.PutObject(ctx, o.mbu, o.remoteFn(name), bytes.NewReader(buf), int64(len(buf)), minio.PutObjectOptions{ContentType: "application/octet-stream"})
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
