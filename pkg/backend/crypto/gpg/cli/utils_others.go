// +build !windows

package cli

// #include <unistd.h>
import "C"

// see https://www.gnupg.org/documentation/manuals/gnupg/Invoking-GPG_002dAGENT.html
func tty() string {
	name, err := C.ttyname(C.int(0))
	if err != nil {
		return ""
	}
	return C.GoString(name)
}
