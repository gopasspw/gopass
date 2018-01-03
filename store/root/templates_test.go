package root

import (
	"context"
	"io/ioutil"
	"os"
	"testing"

	"github.com/fatih/color"
	"github.com/justwatchcom/gopass/utils/ctxutil"
	"github.com/justwatchcom/gopass/utils/out"
	"github.com/stretchr/testify/assert"
)

func TestTemplate(t *testing.T) {
	tempdir, err := ioutil.TempDir("", "gopass-")
	if err != nil {
		t.Fatalf("Failed to create tempdir: %s", err)
	}
	defer func() {
		_ = os.RemoveAll(tempdir)
	}()

	ctx := context.Background()
	ctx = ctxutil.WithAlwaysYes(ctx, true)
	ctx = out.WithHidden(ctx, true)
	color.NoColor = true

	rs, err := createRootStore(ctx, tempdir)
	assert.NoError(t, err)

	tt, err := rs.TemplateTree()
	assert.NoError(t, err)
	assert.Equal(t, "gopass\n", tt.Format(0))
}
