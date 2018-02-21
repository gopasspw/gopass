// +build !gtk_3_6,!gtk_3_8,!gtk_3_10,!gtk_3_12,!gtk_3_14

// See: https://developer.gnome.org/gtk3/3.16/api-index-3-16.html

package gtk

// #cgo pkg-config: gtk+-3.0
// #include <gtk/gtk.h>
// #include "gtk.go.h"
import "C"
import (
	"runtime"
	"unsafe"
)

// PaperSizeNewFromIpp is a wrapper around gtk_paper_size_new_from_ipp().
func PaperSizeNewFromIPP(name string, width, height float64) (*PaperSize, error) {
	cstr := C.CString(name)
	defer C.free(unsafe.Pointer(cstr))

	c := C.gtk_paper_size_new_from_ipp((*C.gchar)(cstr), C.gdouble(width), C.gdouble(height))
	if c == nil {
		return nil, nilPtrErr
	}

	t := &PaperSize{c}
	runtime.SetFinalizer(t, (*PaperSize).free)
	return t, nil
}

// IsIPP() is a wrapper around gtk_paper_size_is_ipp().
func (ps *PaperSize) IsIPP() bool {
	c := C.gtk_paper_size_is_ipp(ps.native())
	return gobool(c)
}
