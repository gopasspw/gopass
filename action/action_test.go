package action

import (
	"bytes"
	"context"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/blang/semver"
	"github.com/google/go-cmp/cmp"
	gpgmock "github.com/justwatchcom/gopass/backend/gpg/mock"
	"github.com/justwatchcom/gopass/config"
)

func newMock(ctx context.Context, dir string) (*Action, error) {
	cfg := config.New()
	cfg.Root.Path = filepath.Join(dir, "store")
	sv := semver.Version{}
	gpg := gpgmock.New()

	if err := os.MkdirAll(cfg.Root.Path, 0700); err != nil {
		return nil, err
	}
	if err := os.Setenv("GOPASS_CONFIG", filepath.Join(dir, ".gopass.yml")); err != nil {
		return nil, err
	}
	if err := os.Setenv("GOPASS_HOMEDIR", dir); err != nil {
		return nil, err
	}
	if err := os.Unsetenv("PAGER"); err != nil {
		return nil, err
	}
	if err := os.Setenv("CHECKPOINT_DISABLE", "true"); err != nil {
		return nil, err
	}
	if err := os.Setenv("GOPASS_NO_NOTIFY", "true"); err != nil {
		return nil, err
	}
	if err := ioutil.WriteFile(filepath.Join(dir, "store", ".gpg-id"), []byte("0xDEADBEEF"), 0600); err != nil {
		return nil, err
	}
	if err := ioutil.WriteFile(filepath.Join(dir, "store", "foo.gpg"), []byte("0xDEADBEEF"), 0600); err != nil {
		return nil, err
	}

	return newAction(ctx, cfg, sv, gpg)
}

func capture(t *testing.T, fn func() error) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	done := make(chan string)
	go func() {
		buf := &bytes.Buffer{}
		_, _ = io.Copy(buf, r)
		done <- buf.String()
	}()

	err := fn()
	// back to normal
	_ = w.Close()
	os.Stdout = old
	if err != nil {
		t.Errorf("Error: %s", err)
	}
	out := <-done
	return strings.TrimSpace(out)
}

func TestAction(t *testing.T) {
	td, err := ioutil.TempDir("", "gopass-")
	if err != nil {
		t.Fatalf("Error: %s", err)
	}
	defer func() {
		_ = os.RemoveAll(td)
	}()

	ctx := context.Background()
	act, err := newMock(ctx, td)
	if err != nil {
		t.Fatalf("Error: %s", err)
	}

	if an := act.Name; an != "action.test" {
		t.Errorf("Wrong binary name: '%s' != '%s'", an, "action.test")
	}

	want := filepath.Join(td, "store")
	if as := act.String(); !strings.Contains(as, want) {
		t.Errorf("act.String(): '%s' != '%s'", want, as)
	}
	if !act.HasGPG() {
		t.Errorf("no gpg")
	}
	if lm := len(act.Store.Mounts()); lm != 0 {
		t.Errorf("Too many mounts: %d", lm)
	}
}

func TestUmask(t *testing.T) {
	for _, vn := range []string{"GOPASS_UMASK", "PASSWORD_STORE_UMASK"} {
		for in, out := range map[string]int{
			"002":      02,
			"0777":     0777,
			"000":      0,
			"07557575": 077,
		} {
			_ = os.Setenv(vn, in)
			if um := umask(); um != out {
				t.Errorf("[%s=%s] %o != %o", vn, in, um, out)
			}
			_ = os.Unsetenv(vn)
		}
	}
}

func TestGpgOpts(t *testing.T) {
	for _, vn := range []string{"GOPASS_GPG_OPTS", "PASSWORD_STORE_GPG_OPTS"} {
		for in, out := range map[string][]string{
			"": nil,
			"--decrypt --armor --recipient 0xDEADBEEF": {"--decrypt", "--armor", "--recipient", "0xDEADBEEF"},
		} {
			_ = os.Setenv(vn, in)
			if gp := gpgOpts(); !cmp.Equal(gp, out) {
				t.Errorf("[%s=%s] %+v != %+v", vn, in, gp, out)
			}
			_ = os.Unsetenv(vn)
		}
	}
}
