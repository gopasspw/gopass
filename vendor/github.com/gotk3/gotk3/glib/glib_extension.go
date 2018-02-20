//glib_extension contains definitions and functions to interface between glib/gtk/gio and go universe

package glib

import (
	"reflect"
)

// Should be implemented by  any class which need special conversion like
// gtk.Application -> gio.Application
type IGlibConvert interface {
	//  If convertion can't be done, function have to panic with a message that it can't convert to type
	Convert(reflect.Type) reflect.Value
}

var (
	IGlibConvertType reflect.Type
)
