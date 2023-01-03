package set

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestString(t *testing.T) {
	t.Parallel()

	s1 := New[string]()
	assert.Equal(t, "ø", s1.String())

	s1.Add("a", "b", "c")
	assert.Equal(t, "{a, b, c}", s1.String())

	s2 := New(1, 2, 3, 4)
	assert.Equal(t, "{1, 2, 3, 4}", s2.String())
}

func TestElements(t *testing.T) {
	t.Parallel()

	s1 := New("c", "b", "a")
	assert.Equal(t, []string{"a", "b", "c"}, s1.Elements())
}

func TestClone(t *testing.T) {
	t.Parallel()

	s1 := New("a", "b", "c")
	s2 := s1.Clone()

	assert.Equal(t, s1, s2)

	s1.Add("d")
	assert.NotEqual(t, s1, s2)
}

func TestUpdate(t *testing.T) {
	t.Parallel()

	s1 := New("a", "b", "c")
	s1.Update(New("c", "d", "e"))

	assert.Equal(t, New("a", "b", "c", "d", "e"), s1)

	var s2 Set[string]
	s2.Update(s1)
	assert.Equal(t, New("a", "b", "c", "d", "e"), s2)
}

func TestEquals(t *testing.T) {
	t.Parallel()

	s1 := New("a", "b", "c")
	s2 := s1.Clone()

	assert.True(t, s1.Equals(s2))

	s1.Add("d")
	assert.False(t, s1.Equals(s2))
}

func TestIsSubset(t *testing.T) {
	t.Parallel()

	s1 := New("a", "b", "b", "c", "d")
	s2 := New("c", "d")

	assert.True(t, s2.IsSubset(s1))

	s3 := New[string]()
	assert.True(t, s3.IsSubset(s2))
	assert.False(t, s2.IsSubset(s3))

	s4 := New("foo")
	assert.False(t, s4.IsSubset(s1))
}

func TestUnion(t *testing.T) {
	t.Parallel()

	s1 := New("a", "b", "b", "c", "d")
	s2 := New("c", "d")
	s3 := New("foo", "bar")

	us := s1.Union(s2).Union(s3)

	assert.Equal(t, New("a", "b", "c", "d", "foo", "bar"), us)

	s4 := New[string]()
	assert.Equal(t, s3, s4.Union(s3))

	assert.Equal(t, s3, s3.Union(s4))
}

func TestDifference(t *testing.T) {
	t.Parallel()

	s1 := New("a", "b", "c", "d")
	s2 := New("c", "d")

	ds := s1.Difference(s2)

	assert.Equal(t, New("a", "b"), ds)
}

func TestSymmetricDifference(t *testing.T) {
	t.Parallel()

	s1 := New("a", "b", "c", "d")
	s2 := New("c", "d", "e")

	ds := s1.SymmetricDifference(s2)

	assert.Equal(t, New("a", "b", "e"), ds)
}

func TestAdd(t *testing.T) {
	t.Parallel()

	var s1 Set[string]
	s1.Add("a")

	assert.Equal(t, New("a"), s1)
}

func TestRemove(t *testing.T) {
	t.Parallel()

	s1 := New("a", "b", "b", "c", "d")
	s2 := New("c", "d")
	s1.Remove(s2)

	assert.Equal(t, New("a", "b"), s1)

	s3 := New[string]()
	assert.False(t, s3.Remove(New("a")))
}

func TestDiscard(t *testing.T) {
	t.Parallel()

	s1 := New("a", "b", "b", "c", "d")
	s1.Discard("c", "d")

	assert.Equal(t, New("a", "b"), s1)

	s3 := New[string]()
	assert.False(t, s3.Discard("a"))
}

func TestMap(t *testing.T) {
	t.Parallel()

	s1 := New(1, 2, 3)
	s2 := s1.Map(func(i int) int {
		return i + 1
	})

	assert.Equal(t, New(2, 3, 4), s2)
}

func TestEach(t *testing.T) {
	t.Parallel()

	seen := 0

	s1 := New(1, 2, 3)
	s1.Each(func(i int) {
		seen++
	})
	assert.Equal(t, s1.Len(), seen)
}

func TestSelect(t *testing.T) {
	t.Parallel()

	s1 := New(1, 2, 3, 4, 5, 6, 7, 8, 9, 10)
	s2 := s1.Select(func(i int) bool {
		return i%2 == 0
	})

	assert.Equal(t, New(2, 4, 6, 8, 10), s2)
}

func TestPartition(t *testing.T) {
	t.Parallel()

	s1 := New(1, 2, 3, 4, 5, 6, 7, 8, 9, 10)
	s2, s3 := s1.Partition(func(i int) bool {
		return i%2 == 0
	})

	assert.Equal(t, New(2, 4, 6, 8, 10), s2)
	assert.Equal(t, New(1, 3, 5, 7, 9), s3)
}

func TestChoose(t *testing.T) {
	t.Parallel()

	s1 := New(1, 2, 3, 4, 5, 6, 7, 8, 9, 10)
	v, ok := s1.Choose(func(i int) bool {
		return i%3 == 0 && i/3 == 3
	})
	assert.Equal(t, 9, v)
	assert.True(t, ok)

	v, ok = s1.Choose(func(i int) bool {
		return i > 1024
	})
	assert.Equal(t, 0, v)
	assert.False(t, ok)

	s2 := New(1)
	v, ok = s2.Choose(nil)
	assert.Equal(t, v, 1)
	assert.True(t, ok)
}

func TestCount(t *testing.T) {
	t.Parallel()

	s1 := New(1, 2, 3, 4, 5, 6, 7, 8, 9, 10)
	c := s1.Count(func(i int) bool {
		return i%3 == 0
	})
	assert.Equal(t, 3, c)
}
