package strutil

import "unsafe"

type stringHeader struct {
	data uintptr
	len  int
}
type sliceHeader struct {
	data uintptr
	len  int
	cap  int
}

func ToReadonlyBytes(s string) []byte {
	strHeader := (*stringHeader)(unsafe.Pointer(&s))
	slHeader := sliceHeader{
		data: strHeader.data,
		len:  strHeader.len,
		cap:  strHeader.len,
	}
	return *(*[]byte)(unsafe.Pointer(&slHeader))
}

func ToReadonlyString(bytes []byte) string {
	sliceHeader := (*sliceHeader)(unsafe.Pointer(&bytes))
	strHeader := stringHeader{
		data: sliceHeader.data,
		len:  sliceHeader.len,
	}
	return *(*string)(unsafe.Pointer(&strHeader))
}
