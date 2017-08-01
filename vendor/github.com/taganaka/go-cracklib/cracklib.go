// Package cracklib provides a Golang binding for cracklib
// https://github.com/cracklib/cracklib
package cracklib

// #cgo LDFLAGS: -lcrack
// #include <stdlib.h>
// #include <crack.h>
import "C"
import "unsafe"

// FascistCheck checks a potential password for guessability
// It returns an error message and a boolean value
// The error message will be "" if ok is true
func FascistCheck(pw string) (message string, ok bool) {
	path := C.GetDefaultCracklibDict()
	pwptr := C.CString(pw)
	defer C.free(unsafe.Pointer(pwptr))
	v := C.FascistCheck(pwptr, path)
	message = C.GoString(v)
	if message != "" {
		return message, false
	}
	return "", true
}

// FascistCheckUser executes tests against an arbitrary user
// It returns an error message and a boolean value
// The error message will be "" if ok is true
func FascistCheckUser(pw string, user string) (message string, ok bool) {
	path := C.GetDefaultCracklibDict()
	pwptr := C.CString(pw)
	defer C.free(unsafe.Pointer(pwptr))
	userptr := C.CString(user)
	defer C.free(unsafe.Pointer(userptr))
	v := C.FascistCheckUser(pwptr, path, userptr, nil)
	message = C.GoString(v)
	if message != "" {
		return message, false
	}
	return "", true
}

// extern const char *FascistCheck(const char *pw, const char *dictpath);
// extern const char *FascistCheckUser(const char *pw, const char *dictpath,
// 				    const char *user, const char *gecos);
//
// /* This function returns the compiled in value for DEFAULT_CRACKLIB_DICT.
//  */
// extern const char *GetDefaultCracklibDict(void);
