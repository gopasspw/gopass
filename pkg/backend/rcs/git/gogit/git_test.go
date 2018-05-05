package gogit

import (
	"context"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/justwatchcom/gopass/pkg/ctxutil"
	"github.com/justwatchcom/gopass/pkg/store"

	"github.com/stretchr/testify/assert"
	"gopkg.in/src-d/go-git.v4/config"
)

func TestCloneLocal(t *testing.T) {
	td, err := ioutil.TempDir("", "gopass-")
	assert.NoError(t, err)
	defer func() {
		_ = os.RemoveAll(td)
	}()

	ctx := context.Background()
	ctx = ctxutil.WithAlwaysYes(ctx, true)

	// init git repo
	repo := filepath.Join(td, "git")
	assert.NoError(t, run(ctx, "", "git", "init", repo))
	assert.NoError(t, ioutil.WriteFile(filepath.Join(repo, "foo.txt"), []byte("hello world"), 0644))
	assert.NoError(t, run(ctx, repo, "git", "add", "foo.txt"))
	assert.NoError(t, run(ctx, repo, "git", "commit", "-am'foo'"))

	// clone
	path := filepath.Join(td, "gogit")
	g, err := Clone(ctx, repo, path)
	assert.NoError(t, err)
	assert.NotNil(t, g)

	// list remotes
	list, err := g.repo.Remotes()
	assert.NoError(t, err)
	t.Logf("Remotes: %+v", list)

	// add file
	assert.NoError(t, ioutil.WriteFile(filepath.Join(path, "bar.txt"), []byte("zab"), 0644))
	assert.NoError(t, g.Add(ctx, "bar.txt"))
	assert.NoError(t, g.Commit(ctx, "Added bar.txt"))
	assert.EqualError(t, g.Commit(ctx, "boooo"), store.ErrGitNothingToCommit.Error())

	// push to remote
	assert.Error(t, g.PushPull(ctx, "push", "", ""))

	// test revisions
	_, err = g.Revisions(ctx, "foo")
	assert.Error(t, err)
	_, err = g.GetRevision(ctx, "foo", "bar")
	assert.Error(t, err)
}

func TestCloneSSH(t *testing.T) {
	t.Skip("need remote")

	td, err := ioutil.TempDir("", "gopass-")
	assert.NoError(t, err)
	defer func() {
		_ = os.RemoveAll(td)
	}()

	ctx := context.Background()
	ctx = ctxutil.WithAlwaysYes(ctx, true)

	// clone
	path := filepath.Join(td, "gogit")
	g, err := Clone(ctx, "", path)
	assert.NoError(t, err)
	assert.NotNil(t, g)

	// list remotes
	list, err := g.repo.Remotes()
	assert.NoError(t, err)
	t.Logf("Remotes: %+v", list)
}

func TestInit(t *testing.T) {
	td, err := ioutil.TempDir("", "gopass-")
	assert.NoError(t, err)
	defer func() {
		//_ = os.RemoveAll(td)
	}()

	ctx := context.Background()
	ctx = ctxutil.WithAlwaysYes(ctx, true)

	// init git repo
	repo := filepath.Join(td, "git")
	assert.NoError(t, run(ctx, "", "git", "init", "--bare", repo))

	// init
	path := filepath.Join(td, "gogit")
	g, _ := Init(ctx, path)
	//assert.NoError(t, err)
	assert.NotNil(t, g)

	// add remote
	_, err = g.repo.CreateRemote(&config.RemoteConfig{
		Name: "origin",
		URLs: []string{"file://" + repo},
	})
	assert.NoError(t, err)

	// list remotes
	list, err := g.repo.Remotes()
	assert.NoError(t, err)
	t.Logf("Remotes: %+v", list)

	// add file
	assert.NoError(t, ioutil.WriteFile(filepath.Join(path, "bar.txt"), []byte("zab"), 0644))
	assert.NoError(t, g.Add(ctx, "bar.txt"))
	assert.NoError(t, g.Commit(ctx, "Added bar.txt"))
	assert.EqualError(t, g.Commit(ctx, "boooo"), store.ErrGitNothingToCommit.Error())

	// push to remote
	assert.NoError(t, g.PushPull(ctx, "push", "", ""))

	g, err = Open(path)
	assert.NoError(t, err)
	assert.Error(t, g.Cmd(ctx, "foo", "bar"))
	assert.Error(t, g.InitConfig(ctx, "foo", "bar"))
	assert.Equal(t, "go-git", g.Name())
	assert.NoError(t, g.AddRemote(ctx, "foo", "file:///tmp/foo"))

	// list remotes
	list, err = g.repo.Remotes()
	assert.NoError(t, err)
	t.Logf("Remotes: %+v", list)
}

func run(ctx context.Context, wd, command string, args ...string) error {
	bin, err := exec.LookPath(command)
	if err != nil {
		return err
	}
	cmd := exec.CommandContext(ctx, bin, args...)
	cmd.Stdout = os.Stdout
	if wd != "" {
		cmd.Dir = wd
	}
	return cmd.Run()
}
