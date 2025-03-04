package mockstore

import (
	"context"
	"testing"

	"github.com/gopasspw/gopass/pkg/gopass/secrets"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMockStore(t *testing.T) {
	ctx := context.Background()
	store := New("test")

	t.Run("String", func(t *testing.T) {
		assert.Equal(t, "mockstore", store.String())
	})

	t.Run("GetTemplate", func(t *testing.T) {
		data, err := store.GetTemplate(ctx, "test")
		require.NoError(t, err)
		assert.Empty(t, data)
	})

	t.Run("HasTemplate", func(t *testing.T) {
		assert.False(t, store.HasTemplate(ctx, "test"))
	})

	t.Run("ListTemplates", func(t *testing.T) {
		assert.Nil(t, store.ListTemplates(ctx, "test"))
	})

	t.Run("LookupTemplate", func(t *testing.T) {
		data, found := store.LookupTemplate(ctx, "test")
		assert.False(t, found)
		assert.Empty(t, data)
	})

	t.Run("RemoveTemplate", func(t *testing.T) {
		assert.NoError(t, store.RemoveTemplate(ctx, "test"))
	})

	t.Run("SetTemplate", func(t *testing.T) {
		assert.NoError(t, store.SetTemplate(ctx, "test", []byte("data")))
	})

	t.Run("TemplateTree", func(t *testing.T) {
		tree, err := store.TemplateTree(ctx)
		assert.Error(t, err)
		assert.Nil(t, tree)
	})

	t.Run("AddRecipient", func(t *testing.T) {
		assert.NoError(t, store.AddRecipient(ctx, "test"))
	})

	t.Run("GetRecipients", func(t *testing.T) {
		recipients, err := store.GetRecipients(ctx, "test")
		assert.Error(t, err)
		assert.Nil(t, recipients)
	})

	t.Run("RemoveRecipient", func(t *testing.T) {
		assert.NoError(t, store.RemoveRecipient(ctx, "test"))
	})

	t.Run("SaveRecipients", func(t *testing.T) {
		assert.NoError(t, store.SaveRecipients(ctx))
	})

	t.Run("Recipients", func(t *testing.T) {
		assert.Nil(t, store.Recipients(ctx))
	})

	t.Run("ImportMissingPublicKeys", func(t *testing.T) {
		assert.NoError(t, store.ImportMissingPublicKeys(ctx))
	})

	t.Run("ExportMissingPublicKeys", func(t *testing.T) {
		ok, err := store.ExportMissingPublicKeys(ctx, []string{"test"})
		require.NoError(t, err)
		assert.False(t, ok)
	})

	t.Run("Fsck", func(t *testing.T) {
		assert.NoError(t, store.Fsck(ctx, "test"))
	})

	t.Run("Path", func(t *testing.T) {
		assert.Empty(t, store.Path())
	})

	t.Run("URL", func(t *testing.T) {
		assert.Equal(t, "mockstore://", store.URL())
	})

	t.Run("Crypto", func(t *testing.T) {
		assert.NotNil(t, store.Crypto())
	})

	t.Run("Storage", func(t *testing.T) {
		assert.NotNil(t, store.Storage())
	})

	t.Run("GitInit", func(t *testing.T) {
		assert.NoError(t, store.GitInit(ctx, "test", "test"))
	})

	t.Run("Alias", func(t *testing.T) {
		assert.Equal(t, "test", store.Alias())
	})

	t.Run("Copy", func(t *testing.T) {
		sec := secrets.New()
		sec.SetPassword("password")
		err := store.Set(ctx, "from", sec)
		require.NoError(t, err)
		assert.NoError(t, store.Copy(ctx, "from", "to"))
		sec, err = store.Get(ctx, "to")
		require.NoError(t, err)
		assert.Equal(t, "password", sec.Password())
	})

	t.Run("Delete", func(t *testing.T) {
		sec := secrets.New()
		sec.Set("password", "password")
		assert.NoError(t, store.Set(ctx, "test", sec))
		assert.NoError(t, store.Delete(ctx, "test"))
		_, err := store.Get(ctx, "test")
		assert.Error(t, err)
	})

	t.Run("Equals", func(t *testing.T) {
		other := New("other")
		assert.False(t, store.Equals(other))
	})

	t.Run("Exists", func(t *testing.T) {
		sec := secrets.New()
		sec.Set("password", "password")
		assert.NoError(t, store.Set(ctx, "test", sec))
		assert.True(t, store.Exists(ctx, "test"))
	})

	t.Run("Get", func(t *testing.T) {
		sec := secrets.New()
		sec.SetPassword("password")
		assert.NoError(t, store.Set(ctx, "test", sec))
		sec, err := store.Get(ctx, "test")
		require.NoError(t, err)
		assert.Equal(t, "password", sec.Password())
	})

	t.Run("GetRevision", func(t *testing.T) {
		_, err := store.GetRevision(ctx, "test", "revision")
		assert.Error(t, err)
	})

	t.Run("Init", func(t *testing.T) {
		assert.NoError(t, store.Init(ctx, "test"))
	})

	t.Run("Initialized", func(t *testing.T) {
		assert.True(t, store.Initialized(ctx))
	})

	t.Run("IsDir", func(t *testing.T) {
		sec := secrets.New()
		sec.Set("password", "password")
		assert.NoError(t, store.Set(ctx, "test/dir", sec))
		assert.True(t, store.IsDir(ctx, "test"))
	})

	t.Run("List", func(t *testing.T) {
		sec := secrets.New()
		sec.Set("password", "password")
		assert.NoError(t, store.Set(ctx, "test", sec))
		list, err := store.List(ctx, "test")
		require.NoError(t, err)
		assert.Contains(t, list, "test")
	})

	t.Run("ListRevisions", func(t *testing.T) {
		revisions, err := store.ListRevisions(ctx, "test")
		require.NoError(t, err)
		assert.Nil(t, revisions)
	})

	t.Run("Move", func(t *testing.T) {
		sec := secrets.New()
		sec.SetPassword("password")
		assert.NoError(t, store.Set(ctx, "from", sec))
		assert.NoError(t, store.Move(ctx, "from", "to"))
		_, err := store.Get(ctx, "from")
		assert.Error(t, err)
		sec, err = store.Get(ctx, "to")
		require.NoError(t, err)
		assert.Equal(t, "password", sec.Password())
	})

	t.Run("Set", func(t *testing.T) {
		sec := secrets.New()
		sec.SetPassword("password")
		assert.NoError(t, store.Set(ctx, "test", sec))
		sec, err := store.Get(ctx, "test")
		require.NoError(t, err)
		assert.Equal(t, "password", sec.Password())
	})

	t.Run("Prune", func(t *testing.T) {
		assert.Error(t, store.Prune(ctx, "test"))
	})

	t.Run("Valid", func(t *testing.T) {
		assert.True(t, store.Valid())
	})

	t.Run("MountPoints", func(t *testing.T) {
		assert.Nil(t, store.MountPoints())
	})

	t.Run("Link", func(t *testing.T) {
		assert.NoError(t, store.Link(ctx, "from", "to"))
	})
}
