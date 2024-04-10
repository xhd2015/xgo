package tls_test

import (
	"testing"

	"github.com/xhd2015/xgo/runtime/tls"
)

var a = tls.Declare("a")
var b = tls.DeclareInherit("b")

func TestDeclareLocal(t *testing.T) {
	a.Set(1)

	var v1 interface{}
	done := make(chan struct{})
	go func() {
		v1 = a.Get()
		close(done)
	}()
	<-done

	v2 := a.Get()

	if v1 != nil {
		t.Fatalf("expect sub goroutine get nil, actual: %v", v1)
	}

	if i := v2.(int); i != 1 {
		t.Fatalf("expect current goroutine get %d, actual: %d", 1, i)
	}
}

func TestInerhitLocal(t *testing.T) {
	b.Set(1)

	var v1 interface{}
	done := make(chan struct{})
	go func() {
		v1 = b.Get()
		close(done)
	}()
	<-done

	v2 := b.Get()

	if v1 == nil {
		t.Fatalf("expect sub goroutine inherit b, actual: %v", v1)
	}
	if i := v1.(int); i != 1 {
		t.Fatalf("expect sub goroutine get %d, actual: %d", 1, i)
	}
	if i := v2.(int); i != 1 {
		t.Fatalf("expect current goroutine get %d, actual: %d", 1, i)
	}
}
