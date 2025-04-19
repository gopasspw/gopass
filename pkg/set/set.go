// Package set provides a generic set implementation.
package set

import (
	"fmt"
	"strings"

	"golang.org/x/exp/constraints"
)

// Set is a generic set type.
type Set[K constraints.Ordered] map[K]bool

// New initializes a new Set with the given elements.
func New[K constraints.Ordered](elems ...K) Set[K] {
	s := make(map[K]bool, len(elems))

	for _, e := range elems {
		s[e] = true
	}

	return s
}

// String returns a string representation of the set.
func (s Set[K]) String() string {
	if s.Empty() {
		return "ø"
	}
	elems := make([]string, len(s))
	for i, e := range s.Elements() {
		elems[i] = fmt.Sprintf("%v", e)
	}

	return fmt.Sprintf("{%s}", strings.Join(elems, ", "))
}

// Elements returns the elements of the set in
// sorted order.
func (s Set[K]) Elements() []K {
	return SortedKeys(s)
}

// Empty returns true if the set is empty.
func (s Set[K]) Empty() bool {
	return len(s) == 0
}

// Len returns the length of the set.
func (s Set[K]) Len() int {
	return len(s)
}

// Clone creates a copy of the set.
func (s Set[K]) Clone() Set[K] {
	c := Set[K]{}
	c.Update(s)

	return c
}

// Update adds all elements from s2 to the set.
func (s *Set[K]) Update(s2 Set[K]) bool {
	il := len(*s)
	if *s == nil && len(s2) > 0 {
		*s = make(Set[K], len(s2))
	}
	for k := range s2 {
		(*s)[k] = true
	}

	return len(*s) != il
}

// Equals returns true if s and s2 contain
// exactly the same elements.
func (s Set[K]) Equals(s2 Set[K]) bool {
	return len(s) == len(s2) && s.IsSubset(s2)
}

// Contains returs true if the set contains the presented
// element.
func (s Set[K]) Contains(e K) bool {
	for k := range s {
		if k == e {
			return true
		}
	}

	return false
}

// IsSubset returns true if all elements of s
// are contained in s2.
func (s Set[K]) IsSubset(s2 Set[K]) bool {
	if s.Empty() {
		return true
	}
	if len(s) > len(s2) {
		return false
	}

	for k := range s {
		if !s2[k] {
			return false
		}
	}

	return true
}

// Union returns a new set containing all elements from
// s and s2. A ∪ B.
func (s Set[K]) Union(s2 Set[K]) Set[K] {
	if s.Empty() {
		return s2
	}
	if s2.Empty() {
		return s
	}

	set := make(Set[K])
	for k := range s {
		set[k] = true
	}
	for k := range s2 {
		set[k] = true
	}

	return set
}

// Difference returns the set difference. That is all the things that are in s
// but not in s2. A \ B.
func (s Set[K]) Difference(s2 Set[K]) Set[K] {
	if s2.Empty() {
		return s
	}
	if s.Empty() {
		return New[K]()
	}

	set := make(Set[K])
	for k := range s {
		if s2[k] {
			continue
		}

		set[k] = true
	}

	return set
}

// SymmetricDifference returns the symmetric difference. That is all the things that are
// in s or s2 but not in both. A Δ B.
func (s Set[K]) SymmetricDifference(s2 Set[K]) Set[K] {
	if s2.Empty() {
		return s
	}
	if s.Empty() {
		return s2
	}

	set := make(Set[K])
	for k := range s {
		if s2[k] {
			continue
		}

		set[k] = true
	}
	for k := range s2 {
		if s[k] {
			continue
		}

		set[k] = true
	}

	return set
}

// Add adds the given elements to the set.
func (s *Set[K]) Add(elems ...K) bool {
	il := len(*s)
	if *s == nil {
		*s = make(Set[K])
	}
	for _, k := range elems {
		(*s)[k] = true
	}

	return len(*s) != il
}

// Remove deletes the given element from the set.
func (s Set[K]) Remove(s2 Set[K]) bool {
	if s.Empty() {
		return false
	}
	il := len(s)
	for k := range s2 {
		delete(s, k)
	}

	return len(s) != il
}

// Discard deletes the given elements from the set.
func (s Set[K]) Discard(elems ...K) bool {
	if s.Empty() {
		return false
	}
	il := len(s)
	for _, e := range elems {
		delete(s, e)
	}

	return len(s) != il
}

// Map returns a new set by applied the function f
// to all it's elements.
func (s Set[K]) Map(f func(K) K) Set[K] {
	out := make(Set[K], len(s))
	for k := range s {
		out.Add(f(k))
	}

	return out
}

// Each applies the function f to all it's elements.
func (s Set[K]) Each(f func(K)) {
	for k := range s {
		f(k)
	}
}

// Select returns a new set with all the elements for
// that f returns true.
func (s Set[K]) Select(f func(K) bool) Set[K] {
	out := make(Set[K], len(s))
	for k := range s {
		if f(k) {
			out.Add(k)
		}
	}

	return out
}

// Partition returns two new sets: the first contains all
// the elements for which f returns true. The seconds the others.
func (s Set[K]) Partition(f func(K) bool) (Set[K], Set[K]) {
	yes := make(Set[K], len(s))
	no := make(Set[K], len(s))
	for k := range s {
		if f(k) {
			yes.Add(k)

			continue
		}

		no.Add(k)
	}

	return yes, no
}

// Choose returns the first element for which f returns true.
func (s Set[K]) Choose(f func(K) bool) (K, bool) {
	if f == nil {
		for k := range s {
			return k, true
		}
	}
	for k := range s {
		if f(k) {
			return k, true
		}
	}

	var zero K

	return zero, false
}

// Count returns the number of elements for which f returns true.
func (s Set[K]) Count(f func(K) bool) int {
	n := 0

	for k := range s {
		if f(k) {
			n++
		}
	}

	return n
}
