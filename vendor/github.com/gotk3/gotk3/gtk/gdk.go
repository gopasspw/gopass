package gtk

// #include <gtk/gtk.h>
// #include "gtk.go.h"
import "C"
import "github.com/gotk3/gotk3/gdk"

func nativeGdkRectangle(rect gdk.Rectangle) *C.GdkRectangle {
	// Note: Here we can't use rect.GdkRectangle because it would return
	// C type prefixed with gdk package. A ways how to resolve this Go
	// issue with same C structs in different Go packages is documented
	// here https://github.com/golang/go/issues/13467 .
	// This is the easiest way how to resolve the problem.
	return &C.GdkRectangle{
		x:      C.int(rect.GetX()),
		y:      C.int(rect.GetY()),
		width:  C.int(rect.GetWidth()),
		height: C.int(rect.GetHeight()),
	}
}
