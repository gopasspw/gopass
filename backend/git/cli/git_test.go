package cli

import (
	"bytes"
	"context"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	gpgmock "github.com/justwatchcom/gopass/backend/gpg/mock"
	"github.com/justwatchcom/gopass/utils/ctxutil"
	"github.com/justwatchcom/gopass/utils/out"
	"github.com/stretchr/testify/assert"
)

func TestGit(t *testing.T) {
	td, err := ioutil.TempDir("", "gopass-")
	if err != nil {
		t.Fatalf("Error: %s", err)
	}
	defer func() {
		_ = os.RemoveAll(td)
	}()

	ctx := context.Background()
	ctx = ctxutil.WithAlwaysYes(ctx, true)

	buf := &bytes.Buffer{}
	out.Stdout = buf
	defer func() {
		out.Stdout = os.Stdout
	}()

	gpg := gpgmock.New()
	git := New(td, gpg.Binary())

	assert.NoError(t, git.Init(ctx, "0xDEADBEEF", "Dead Beef", "dead.beef@example.org"))
	sv := git.Version(ctx)
	t.Logf("Version: %s", sv.String())

	assert.Equal(t, true, git.IsInitialized())
	tf := filepath.Join(td, "some-file")
	assert.NoError(t, ioutil.WriteFile(tf, []byte("foobar"), 0644))
	assert.NoError(t, git.Add(ctx, "some-file"))
	assert.Equal(t, true, git.HasStagedChanges(ctx))
	assert.Error(t, git.Push(ctx, "origin", "master"))
	assert.Error(t, git.Pull(ctx, "origin", "master"))

	// flaky
	//assert.NoError(t, git.Commit(ctx, "added some-file"))
}
