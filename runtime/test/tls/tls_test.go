package tls_test

import (
	"fmt"
	"testing"
	"time"

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

// see https://github.com/xhd2015/xgo/issues/96
func TestTimeAfterShouldNotInherit(t *testing.T) {
	b.Set(1)
	goDone := make(chan struct{})
	go func() {
		defer close(goDone)
		v, ok := b.GetOK()
		if !ok {
			panic("expect inherited in new proc")
		}
		if v.(int) != 1 {
			panic(fmt.Errorf("expect inherited to be: %d, actual: %v", 1, v))
		}
	}()
	<-goDone

	timerDone := make(chan struct{})
	time.AfterFunc(10*time.Millisecond, func() {
		defer close(timerDone)
		_, ok := b.GetOK()
		if ok {
			t.Fatalf("expect not inherited in timer")
		}
	})
	<-timerDone
}
