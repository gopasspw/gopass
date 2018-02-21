package cairo

// #cgo pkg-config: cairo cairo-gobject
// #include <stdlib.h>
// #include <cairo.h>
// #include <cairo-gobject.h>
import "C"
import (
	"unsafe"
)

// Format is a representation of Cairo's cairo_format_t.
type Format int

const (
	FORMAT_INVALID   Format = C.CAIRO_FORMAT_INVALID
	FORMAT_ARGB32    Format = C.CAIRO_FORMAT_ARGB32
	FORMAT_RGB24     Format = C.CAIRO_FORMAT_RGB24
	FORMAT_A8        Format = C.CAIRO_FORMAT_A8
	FORMAT_A1        Format = C.CAIRO_FORMAT_A1
	FORMAT_RGB16_565 Format = C.CAIRO_FORMAT_RGB16_565
	FORMAT_RGB30     Format = C.CAIRO_FORMAT_RGB30
)

func marshalFormat(p uintptr) (interface{}, error) {
	c := C.g_value_get_enum((*C.GValue)(unsafe.Pointer(p)))
	return Format(c), nil
}
