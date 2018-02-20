package glib

// #cgo pkg-config: glib-2.0 gobject-2.0
// #include <glib.h>
// #include <glib-object.h>
// #include "glib.go.h"
import "C"
import "unsafe"

// SList is a representation of Glib's GSList. A SList must be manually freed
// by either calling Free() or FreeFull()
type SList struct {
	list *C.struct__GSList
}

func WrapSList(obj uintptr) *SList {
	return wrapSList((*C.struct__GSList)(unsafe.Pointer(obj)))
}

func wrapSList(obj *C.struct__GSList) *SList {
	if obj == nil {
		return nil
	}

	//NOTE a list should be freed by calling either
	//g_slist_free() or g_slist_free_full(). However, it's not possible to use a
	//finalizer for this.
	return &SList{obj}
}

func (v *SList) Native() uintptr {
	return uintptr(unsafe.Pointer(v.list))
}

func (v *SList) native() *C.struct__GSList {
	if v == nil || v.list == nil {
		return nil
	}
	return v.list
}

func (v *SList) Append(data uintptr) *SList {
	ret := C.g_slist_append(v.native(), C.gpointer(data))
	if ret == v.native() {
		return v
	}

	return wrapSList(ret)
}

// Length is a wrapper around g_slist_length().
func (v *SList) Length() uint {
	return uint(C.g_slist_length(v.native()))
}

// Next is a wrapper around the next struct field
func (v *SList) Next() *SList {
	n := v.native()
	if n == nil {
		return nil
	}

	return wrapSList(n.next)
}

// Foreach acts the same as g_slist_foreach().
// No user_data arguement is implemented because of Go clojure capabilities.
func (v *SList) Foreach(fn func(ptr unsafe.Pointer)) {
	for l := v; l != nil; l = l.Next() {
		fn(unsafe.Pointer(l.native().data))
	}
}

// Free is a wrapper around g_slist_free().
func (v *SList) Free() {
	C.g_slist_free(v.native())
	v.list = nil
}

// FreeFull is a wrapper around g_slist_free_full().
func (v *SList) FreeFull() {
	//TODO implement GDestroyNotify callback
	C.g_slist_free_full(v.native(), nil)
	v.list = nil
}

// GSList * 	g_slist_alloc ()
// GSList * 	g_slist_prepend ()
// GSList * 	g_slist_insert ()
// GSList * 	g_slist_insert_before ()
// GSList * 	g_slist_insert_sorted ()
// GSList * 	g_slist_remove ()
// GSList * 	g_slist_remove_link ()
// GSList * 	g_slist_delete_link ()
// GSList * 	g_slist_remove_all ()
// void 	g_slist_free_1 ()
// GSList * 	g_slist_copy ()
// GSList * 	g_slist_copy_deep ()
// GSList * 	g_slist_reverse ()
// GSList * 	g_slist_insert_sorted_with_data ()
// GSList * 	g_slist_sort ()
// GSList * 	g_slist_sort_with_data ()
// GSList * 	g_slist_concat ()
// GSList * 	g_slist_last ()
// GSList * 	g_slist_nth ()
// gpointer 	g_slist_nth_data ()
// GSList * 	g_slist_find ()
// GSList * 	g_slist_find_custom ()
// gint 	g_slist_position ()
// gint 	g_slist_index ()
