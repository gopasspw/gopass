package fossilfs

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockFs is a mock implementation of the underlying filesystem.
type MockFs struct {
	mock.Mock
}

func (m *MockFs) Get(ctx context.Context, name string) ([]byte, error) {
	args := m.Called(ctx, name)
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockFs) Set(ctx context.Context, name string, value []byte) error {
	args := m.Called(ctx, name, value)
	return args.Error(0)
}

func (m *MockFs) Delete(ctx context.Context, name string) error {
	args := m.Called(ctx, name)
	return args.Error(0)
}

func (m *MockFs) Exists(ctx context.Context, name string) bool {
	args := m.Called(ctx, name)
	return args.Bool(0)
}

func (m *MockFs) List(ctx context.Context, prefix string) ([]string, error) {
	args := m.Called(ctx, prefix)
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockFs) IsDir(ctx context.Context, name string) bool {
	args := m.Called(ctx, name)
	return args.Bool(0)
}

func (m *MockFs) Prune(ctx context.Context, prefix string) error {
	args := m.Called(ctx, prefix)
	return args.Error(0)
}

func (m *MockFs) Path() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockFs) Fsck(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockFs) Link(ctx context.Context, from, to string) error {
	args := m.Called(ctx, from, to)
	return args.Error(0)
}

func (m *MockFs) Move(ctx context.Context, src, dst string, del bool) error {
	args := m.Called(ctx, src, dst, del)
	return args.Error(0)
}

func TestFossil_Get(t *testing.T) {
	mockFs := new(MockFs)
	fossil := &Fossil{fs: mockFs}
	ctx := context.TODO()
	name := "test"

	mockFs.On("Get", ctx, name).Return([]byte("content"), nil)

	content, err := fossil.Get(ctx, name)
	assert.NoError(t, err)
	assert.Equal(t, []byte("content"), content)

	mockFs.AssertExpectations(t)
}

func TestFossil_Set(t *testing.T) {
	mockFs := new(MockFs)
	fossil := &Fossil{fs: mockFs}
	ctx := context.TODO()
	name := "test"
	value := []byte("content")

	mockFs.On("Set", ctx, name, value).Return(nil)

	err := fossil.Set(ctx, name, value)
	assert.NoError(t, err)

	mockFs.AssertExpectations(t)
}

func TestFossil_Delete(t *testing.T) {
	mockFs := new(MockFs)
	fossil := &Fossil{fs: mockFs}
	ctx := context.TODO()
	name := "test"

	mockFs.On("Delete", ctx, name).Return(nil)

	err := fossil.Delete(ctx, name)
	assert.NoError(t, err)

	mockFs.AssertExpectations(t)
}

func TestFossil_Exists(t *testing.T) {
	mockFs := new(MockFs)
	fossil := &Fossil{fs: mockFs}
	ctx := context.TODO()
	name := "test"

	mockFs.On("Exists", ctx, name).Return(true)

	exists := fossil.Exists(ctx, name)
	assert.True(t, exists)

	mockFs.AssertExpectations(t)
}

func TestFossil_List(t *testing.T) {
	mockFs := new(MockFs)
	fossil := &Fossil{fs: mockFs}
	ctx := context.TODO()
	prefix := "test"

	mockFs.On("List", ctx, prefix).Return([]string{"foo", "bar"}, nil)

	list, err := fossil.List(ctx, prefix)
	assert.NoError(t, err)
	assert.Equal(t, []string{"foo", "bar"}, list)

	mockFs.AssertExpectations(t)
}

func TestFossil_IsDir(t *testing.T) {
	mockFs := new(MockFs)
	fossil := &Fossil{fs: mockFs}
	ctx := context.TODO()
	name := "test"

	mockFs.On("IsDir", ctx, name).Return(true)

	isDir := fossil.IsDir(ctx, name)
	assert.True(t, isDir)

	mockFs.AssertExpectations(t)
}

func TestFossil_Prune(t *testing.T) {
	mockFs := new(MockFs)
	fossil := &Fossil{fs: mockFs}
	ctx := context.TODO()
	prefix := "test"

	mockFs.On("Prune", ctx, prefix).Return(nil)

	err := fossil.Prune(ctx, prefix)
	assert.NoError(t, err)

	mockFs.AssertExpectations(t)
}

func TestFossil_String(t *testing.T) {
	mockFs := new(MockFs)
	fossil := &Fossil{fs: mockFs}

	mockFs.On("Path").Return("/path/to/storage")

	str := fossil.String()
	assert.Contains(t, str, "fossilfs(")
	assert.Contains(t, str, "path:/path/to/storage")

	mockFs.AssertExpectations(t)
}

func TestFossil_Path(t *testing.T) {
	mockFs := new(MockFs)
	fossil := &Fossil{fs: mockFs}

	mockFs.On("Path").Return("/path/to/storage")

	path := fossil.Path()
	assert.Equal(t, "/path/to/storage", path)

	mockFs.AssertExpectations(t)
}

func TestFossil_Fsck(t *testing.T) {
	mockFs := new(MockFs)
	fossil := &Fossil{fs: mockFs}
	ctx := context.TODO()

	mockFs.On("Fsck", ctx).Return(nil)

	err := fossil.Fsck(ctx)
	assert.NoError(t, err)

	mockFs.AssertExpectations(t)
}

func TestFossil_Link(t *testing.T) {
	mockFs := new(MockFs)
	fossil := &Fossil{fs: mockFs}
	ctx := context.TODO()
	from := "from"
	to := "to"

	mockFs.On("Link", ctx, from, to).Return(nil)

	err := fossil.Link(ctx, from, to)
	assert.NoError(t, err)

	mockFs.AssertExpectations(t)
}

func TestFossil_Move(t *testing.T) {
	mockFs := new(MockFs)
	fossil := &Fossil{fs: mockFs}
	ctx := context.TODO()
	src := "src"
	dst := "dst"
	del := true

	mockFs.On("Move", ctx, src, dst, del).Return(nil)

	err := fossil.Move(ctx, src, dst, del)
	assert.NoError(t, err)

	mockFs.AssertExpectations(t)
}
