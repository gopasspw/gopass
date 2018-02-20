// +build !gtk_3_6,!gtk_3_8,!gtk_3_10

// See: https://developer.gnome.org/gtk3/3.12/api-index-3-12.html

package gtk

// #include <gtk/gtk.h>
// #include "gtk.go.h"
import "C"
import "unsafe"

// GetAccelsForAction is a wrapper around gtk_application_get_accels_for_action().
func (v *Application) GetAccelsForAction(act string) []string {
	cstr1 := (*C.gchar)(C.CString(act))
	defer C.free(unsafe.Pointer(cstr1))

	var descs []string
	c := C.gtk_application_get_accels_for_action(v.native(), cstr1)
	originalc := c
	defer C.g_strfreev(originalc)

	for *c != nil {
		descs = append(descs, C.GoString((*C.char)(*c)))
		c = C.next_gcharptr(c)
	}

	return descs
}

// SetAccelsForAction is a wrapper around gtk_application_set_accels_for_action().
func (v *Application) SetAccelsForAction(act string, accels []string) {
	cstr1 := (*C.gchar)(C.CString(act))
	defer C.free(unsafe.Pointer(cstr1))

	caccels := C.make_strings(C.int(len(accels) + 1))
	defer C.destroy_strings(caccels)

	for i, accel := range accels {
		cstr := C.CString(accel)
		defer C.free(unsafe.Pointer(cstr))
		C.set_string(caccels, C.int(i), (*C.gchar)(cstr))
	}

	C.set_string(caccels, C.int(len(accels)), nil)

	C.gtk_application_set_accels_for_action(v.native(), cstr1, caccels)
}

// ListActionDescriptions is a wrapper around gtk_application_list_action_descriptions().
func (v *Application) ListActionDescriptions() []string {
	var descs []string
	c := C.gtk_application_list_action_descriptions(v.native())
	originalc := c
	defer C.g_strfreev(originalc)

	for *c != nil {
		descs = append(descs, C.GoString((*C.char)(*c)))
		c = C.next_gcharptr(c)
	}

	return descs
}
