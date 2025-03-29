package strutil

import "unsafe"

func ToReadonlyBytes(s string) []byte {
	type stringHeader struct {
		data uintptr
		len  int
	}
	type sliceHeader struct {
		data uintptr
		len  int
		cap  int
	}

	strHeader := (*stringHeader)(unsafe.Pointer(&s))
	slHeader := sliceHeader{
		data: strHeader.data,
		len:  strHeader.len,
		cap:  strHeader.len,
	}
	return *(*[]byte)(unsafe.Pointer(&slHeader))
}
