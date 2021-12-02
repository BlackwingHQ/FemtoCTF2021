package admin

/*
#cgo CFLAGS: -I${SRCDIR}/libflag
#cgo LDFLAGS: -Wl,-rpath,${SRCDIR}/libflag
#cgo LDFLAGS: -L${SRCDIR}/libflag
#cgo LDFLAGS: -lflag

#include <flag.h>
*/
import "C"
import (
	"fmt"
	"net/http"
	"strconv"
	"unsafe"
)

func Admin() {
	mux := http.NewServeMux()
	mux.HandleFunc("/admin", func(w http.ResponseWriter, r *http.Request) {
		input := r.URL.Query().Get("input")
		dbg := r.URL.Query().Get("dbg")
		i, err := strconv.Atoi(dbg)
		if err != nil {
			i = 0
		}
		cS := C.CString(input)
		defer C.free(unsafe.Pointer(cS))
		var flagOut *C.char = C.get_flag(cS, C.int(i))
		getString := C.GoString(flagOut)
		fmt.Println(getString)
		fmt.Fprintf(w, getString)
	})

	http.ListenAndServe("127.0.0.1:1337", mux)
}
