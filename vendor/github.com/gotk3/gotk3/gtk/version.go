package gtk

// #cgo pkg-config: gtk+-3.0
// #include <gtk/gtk.h>
import "C"
import "errors"

func CheckVersion(major, minor, micro uint) error {
	errChar := C.gtk_check_version(C.guint(major), C.guint(minor), C.guint(micro))
	if errChar == nil {
		return nil
	}

	return errors.New(C.GoString((*C.char)(errChar)))
}

func GetMajorVersion() uint {
	v := C.gtk_get_major_version()
	return uint(v)
}

func GetMinorVersion() uint {
	v := C.gtk_get_minor_version()
	return uint(v)
}

func GetMicroVersion() uint {
	v := C.gtk_get_micro_version()
	return uint(v)
}
