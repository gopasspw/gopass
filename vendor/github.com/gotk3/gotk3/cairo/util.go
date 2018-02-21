package cairo

// #cgo pkg-config: cairo cairo-gobject
// #include <stdlib.h>
// #include <cairo.h>
// #include <cairo-gobject.h>
import "C"

func cairobool(b bool) C.cairo_bool_t {
	if b {
		return C.cairo_bool_t(1)
	}
	return C.cairo_bool_t(0)
}

func gobool(b C.cairo_bool_t) bool {
	if b != 0 {
		return true
	}
	return false
}
