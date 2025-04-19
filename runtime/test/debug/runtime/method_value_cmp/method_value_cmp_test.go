package method_value_cmp

import (
	"reflect"
	"testing"
	"unsafe"
)

const N = 512

// mem equal compares
func __xgo_link_mem_equal(a, b unsafe.Pointer, size uintptr) bool {
	return false
}

type large [N]int

var lg large
var lgM func() string

func (c large) String() string {
	if isSame(&c, lgM) {
		return "same large"
	}
	return "large"
}

func TestMethodValueCompare(t *testing.T) {
	t.Skip("Maybe supported later")
	// size should be 4096
	expectedSize := N * unsafe.Sizeof(int(0))
	reflectSize := reflect.TypeOf(&lg).Elem().Size()
	if reflectSize != expectedSize {
		t.Fatalf("expect size to be %d, actual: %d", expectedSize, reflectSize)
	}

	lg[0] = 7
	lgM = lg.String

	var lgLocal large
	lgLocal[0] = 9

	// fmt.Printf("&lg = 0x%x\n", unsafe.Pointer(&lg))
	// fmt.Printf("&lgLocal = 0x%x\n", unsafe.Pointer(&lgLocal))

	localStr := lgLocal.String
	localStrExpect := "large"
	localStrVal := localStr()
	if localStrVal != localStrExpect {
		t.Fatalf("expect localStr to be %s, actual: %s", localStrExpect, localStrVal)
	}

	lgStrExpect := "same large"
	lgStr := lgM()
	if lgStr != lgStrExpect {
		t.Fatalf("expect lgStr to be %s, actual: %s", lgStrExpect, lgStr)
	}

	// then change lgLocal to lg
	lgLocal = lg
	localStrResetExpect := "same large"
	localStrReset := lgLocal.String()
	if localStrReset != localStrResetExpect {
		t.Fatalf("expect localStrReset to be %s, actual: %s", localStrResetExpect, localStrReset)
	}
}

// this function interprets
func isSame(recvPtr interface{}, methodValue interface{}) bool {
	// can also be a constant
	// size := unsafe.Sizeof(*(*large)(nil))
	size := reflect.TypeOf(recvPtr).Elem().Size()
	type _intfRecv struct {
		_    uintptr // type word
		data *byte   // data word
	}

	a := (*_intfRecv)(unsafe.Pointer(&recvPtr))
	type _methodValue struct {
		_    uintptr // pc
		recv byte
	}
	type _intf struct {
		_    uintptr // type word
		data *_methodValue
	}
	ppb := (*_intf)(unsafe.Pointer(&methodValue))
	pb := *ppb
	b := unsafe.Pointer(&pb.data.recv)

	return __xgo_link_mem_equal(unsafe.Pointer(a.data), b, size)
}
