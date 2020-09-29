package stoken

// #cgo CFLAGS: -I/usr/include/libxml2
// #cgo LDFLAGS: -lstoken
// #include <stoken.h>
// #include <stdlib.h>
import "C"
import (
	"errors"
	"time"
	"unsafe"

	"github.com/gopasspw/gopass/pkg/gopass"
)

// Compute token for time given
func Compute(T time.Time, pin, devid, seedpw string, sec gopass.Secret) (string, int64, error) {
	ctx := C.stoken_new()
	if ctx == nil {
		return "", 0, errors.New("Failed to create stoken context")
	}
	defer C.stoken_destroy(ctx)
	seed := C.CString(sec.Get("Password"))
	defer C.free(unsafe.Pointer(seed))
	if rc := C.stoken_import_string(ctx, seed); rc == 22 { // 22 0x16 EINVAL
		return "", 0, errors.New("Invalid token string format")
	} else if rc != 0 {
		return "", 0, errors.New("Failed to import token string")
	}
	var cdevid *C.char
	if C.stoken_devid_required(ctx) != 0 {
		cdevid = C.CString(devid)
		defer C.free(unsafe.Pointer(cdevid))
	}
	var pass *C.char
	if C.stoken_pass_required(ctx) != 0 {
		pass = C.CString(seedpw)
		defer C.free(unsafe.Pointer(pass))
	}
	if rc := C.stoken_decrypt_seed(ctx, cdevid, pass); rc == 22 { // 22 0x16 EINVAL
		return "", 0, errors.New("Decrypt seed MAC failure (PASS or DEVID incorrect)")
	} else if rc != 0 {
		return "", 0, errors.New("Decrypt seed failed")
	}
	info := C.stoken_get_info(ctx)
	if info == nil {
		return "", 0, errors.New("Failed to parse token information")
	}
	var pincode *C.char
	if info.uses_pin != 0 {
		pincode = C.CString(pin)
		defer C.free(unsafe.Pointer(pincode))
	}
	code := C.malloc(C.STOKEN_MAX_TOKENCODE + 1)
	defer C.free(unsafe.Pointer(code))
	if rc := C.stoken_compute_tokencode(ctx, C.long(T.Unix()), pincode, (*C.char)(code)); rc == 22 { // 22 0x16 EINVAL
		return "", int64(info.interval), errors.New("Invalid PIN format")
	} else if rc != 0 {
		return "", int64(info.interval), errors.New("Token generation failed")
	}
	passcode := C.GoString((*C.char)(code))
	return passcode, int64(info.interval), nil
}
