package gpgid

import (
	"context"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	gpgmock "github.com/justwatchcom/gopass/pkg/backend/crypto/plain"
	gitmock "github.com/justwatchcom/gopass/pkg/backend/rcs/noop"
)

func TestInitLoadVerify(t *testing.T) {
	ctx := context.Background()
	gpgm := gpgmock.New()
	gitm := gitmock.New()

	tempdir, err := ioutil.TempDir("", "gopass-")
	if err != nil {
		t.Fatalf("Failed to create tempdir: %s", err)
	}
	t.Logf("Tempdir: %s", tempdir)
	defer func() {
		_ = os.RemoveAll(tempdir)
	}()
	idfile := filepath.Join(tempdir, ".gpg-id")
	if err := ioutil.WriteFile(idfile, []byte("0xDEADBEEF"), 0600); err != nil {
		t.Fatalf("failed to write idfile: %s", err)
	}

	a, err := Init(ctx, gpgm, gitm, idfile)
	if err != nil {
		t.Fatalf("failed to init store: %+v", err)
	}
	t.Logf("a.Recipients: %s", a.Recipients())

	if !reflect.DeepEqual(a.Recipients(), []string{"0xDEADBEEF"}) {
		t.Errorf("Slice mismatch for a")
	}

	b, err := Load(ctx, gpgm, gitm, idfile)
	if err != nil {
		t.Fatalf("failed to load store: %+v", err)
	}
	t.Logf("b.Recipients: %s", b.Recipients())

	if !reflect.DeepEqual(b.Recipients(), []string{"0xDEADBEEF"}) {
		t.Errorf("Slice mismatch for b")
	}

	if err := b.Save(ctx); err != nil {
		t.Fatalf("failed to save acl: %+v", err)
	}

	if err := b.Add(ctx, "0xFEEDBEEF"); err != nil {
		t.Fatalf("failed to add recipient: %s", err)
	}

	c, err := Load(ctx, gpgm, gitm, idfile)
	if err != nil {
		t.Fatalf("failed to load store: %+v", err)
	}
	t.Logf("c.Recipients: %s", c.Recipients())

	if !reflect.DeepEqual(c.Recipients(), []string{"0xDEADBEEF", "0xFEEDBEEF"}) {
		t.Errorf("Slice mismatch for c")
	}

	if err := b.Remove(ctx, "0xDEADBEEF"); err != nil {
		t.Fatalf("failed to remove recipient: %s", err)
	}

	d, err := Load(ctx, gpgm, gitm, idfile)
	if err != nil {
		t.Fatalf("failed to load store: %+v", err)
	}
	t.Logf("d.Recipients: %s", d.Recipients())

	if !reflect.DeepEqual(d.Recipients(), []string{"0xFEEDBEEF"}) {
		t.Errorf("Slice mismatch for d")
	}

	if err := d.verify(ctx); err != nil {
		t.Errorf("Should verify d")
	}
}

func TestInitVerify(t *testing.T) {
	ctx := context.Background()
	gpgm := gpgmock.New()
	gitm := gitmock.New()

	tempdir, err := ioutil.TempDir("", "gopass-")
	if err != nil {
		t.Fatalf("Failed to create tempdir: %s", err)
	}
	t.Logf("Tempdir: %s", tempdir)
	defer func() {
		_ = os.RemoveAll(tempdir)
	}()
	idfile := filepath.Join(tempdir, ".gpg-id")
	if err := ioutil.WriteFile(idfile, []byte("0xDEADBEEF"), 0600); err != nil {
		t.Fatalf("failed to write idfile: %s", err)
	}

	a, err := Init(ctx, gpgm, gitm, idfile)
	if err != nil {
		t.Fatalf("failed to init store: %+v", err)
	}
	t.Logf("a.Recipients: %s", a.Recipients())

	if !reflect.DeepEqual(a.Recipients(), []string{"0xDEADBEEF"}) {
		t.Errorf("Slice mismatch for a")
	}

	if err := a.verify(ctx); err != nil {
		t.Errorf("Should verify a")
	}

	t.Logf("Tokens: %s", a.tokens)
	a.tokens = nil
	if err := a.marshalTokenFile(ctx); err != nil {
		t.Errorf("failed to marshal token file: %s", err)
	}

	if err := a.unmarshalTokenFile(ctx); err != nil {
		t.Errorf("failed to unmarshal token file: %s", err)
	}
	t.Logf("Tokens: %s", a.tokens)
	if err := a.verify(ctx); err == nil {
		t.Errorf("Should NOT verify a")
	}
}
