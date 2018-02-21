package cairo

// #cgo pkg-config: cairo cairo-gobject
// #include <stdlib.h>
// #include <cairo.h>
// #include <cairo-gobject.h>
import "C"

// Translate is a wrapper around cairo_translate.
func (v *Context) Translate(tx, ty float64) {
	C.cairo_translate(v.native(), C.double(tx), C.double(ty))
}

// Scale is a wrapper around cairo_scale.
func (v *Context) Scale(sx, sy float64) {
	C.cairo_scale(v.native(), C.double(sx), C.double(sy))
}

// Rotate is a wrapper around cairo_rotate.
func (v *Context) Rotate(angle float64) {
	C.cairo_rotate(v.native(), C.double(angle))
}

// TODO: The following depend on cairo_matrix_t:
//void 	cairo_transform ()
//void 	cairo_set_matrix ()
//void 	cairo_get_matrix ()
//void 	cairo_identity_matrix ()
//void 	cairo_user_to_device ()
//void 	cairo_user_to_device_distance ()
//void 	cairo_device_to_user ()
//void 	cairo_device_to_user_distance ()
