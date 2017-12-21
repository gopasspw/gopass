package action

import (
	"bytes"
	"context"
	"io"
	"io/ioutil"
	"os"
	"testing"

	"github.com/blang/semver"
	"github.com/google/go-cmp/cmp"
	gpgmock "github.com/justwatchcom/gopass/backend/gpg/mock"
	"github.com/justwatchcom/gopass/config"
)

func newMock(ctx context.Context, dir string) (*Action, error) {
	cfg := config.New()
	cfg.Root.Path = dir
	sv := semver.Version{}
	gpg := gpgmock.New()

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
	return <-done
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

	if as := act.String(); as != "Store(Path: "+td+", Mounts: )" {
		t.Errorf("act.String(): %s", as)
	}
	if !act.HasGPG() {
		t.Errorf("no gpg")
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
