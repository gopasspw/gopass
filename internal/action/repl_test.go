package action

import (
	"bytes"
	"context"
	"testing"

	"github.com/ergochat/readline"
	"github.com/gopasspw/gopass/internal/out"
	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v2"
)

func TestREPL(t *testing.T) {
	ctx := context.Background()
	app := cli.NewApp()
	app.Commands = []*cli.Command{
		{
			Name: "test",
			Action: func(c *cli.Context) error {
				out.Printf(c.Context, "test command executed")
				return nil
			},
		},
	}
	action := &Action{
		Store: &mockStore{},
	}

	buf := &bytes.Buffer{}
	out.Stdout = buf
	out.Stderr = buf

	rl, err := readline.NewEx(&readline.Config{
		Prompt: "gopass> ",
		Stdin:  bytes.NewBufferString("test\nquit\n"),
	})
	assert.NoError(t, err)

	defer func() {
		_ = rl.Close()
	}()

	err = action.REPL(cli.NewContext(app, nil, nil))
	assert.NoError(t, err)
	assert.Contains(t, buf.String(), "test command executed")
}

func TestEntriesForCompleter(t *testing.T) {
	ctx := context.Background()
	action := &Action{
		Store: &mockStore{
			entries: []string{"foo", "bar"},
		},
	}

	completers, err := action.entriesForCompleter(ctx)
	assert.NoError(t, err)
	assert.Len(t, completers, 2)
}

func TestReplCompleteRecipients(t *testing.T) {
	ctx := context.Background()
	action := &Action{
		Store: &mockStore{
			recipients: []string{"alice", "bob"},
		},
	}

	cmd := &cli.Command{
		Name: "remove",
	}

	completers := action.replCompleteRecipients(ctx, cmd)
	assert.Len(t, completers, 2)
}

func TestReplCompleteTemplates(t *testing.T) {
	ctx := context.Background()
	action := &Action{
		Store: &mockStore{
			templates: []string{"tmpl1", "tmpl2"},
		},
	}

	cmd := &cli.Command{
		Name: "templates",
	}

	completers := action.replCompleteTemplates(ctx, cmd)
	assert.Len(t, completers, 2)
}

type mockStore struct {
	entries    []string
	recipients []string
	templates  []string
}

func (m *mockStore) List(ctx context.Context, prefix string) ([]string, error) {
	return m.entries, nil
}

func (m *mockStore) Lock() error {
	return nil
}

func (m *mockStore) recipientsList(ctx context.Context) []string {
	return m.recipients
}

func (m *mockStore) templatesList(ctx context.Context) []string {
	return m.templates
}
