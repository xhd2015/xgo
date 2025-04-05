// package tls provides goroutine-local storage
// it is meant to be used only for tooling
// and instrumenting purpose,
// and should be avoided in production code
package tls

import (
	"sync"

	"github.com/xhd2015/xgo/runtime/internal/runtime"
)

var mut sync.Mutex
var keys []*tlsKey

type TLSKey interface {
	Get() interface{}
	GetOK() (interface{}, bool)
	Set(v interface{})
}

type tlsKey struct {
	name    string // for debugging purepose
	inherit bool
}

var _ TLSKey = (*tlsKey)(nil)

func Declare(name string) TLSKey {
	b := &TLSBuilder{
		name: name,
	}
	return b.Declare()
}

func DeclareInherit(name string) TLSKey {
	b := &TLSBuilder{
		name:    name,
		inherit: true,
	}
	return b.Declare()
}

type TLSBuilder struct {
	name    string
	inherit bool
}

func New() *TLSBuilder {
	return &TLSBuilder{}
}

func (c *TLSBuilder) Name(name string) *TLSBuilder {
	c.name = name
	return c
}

func (c *TLSBuilder) Inherit() *TLSBuilder {
	c.inherit = true
	return c
}

func (c *TLSBuilder) Declare() TLSKey {
	key := &tlsKey{
		name:    c.name,
		inherit: c.inherit,
	}
	mut.Lock()
	keys = append(keys, key)
	mut.Unlock()
	return key
}

func (c *tlsKey) Get() interface{} {
	return runtime.GetG().Get(c)
}

func (c *tlsKey) GetOK() (interface{}, bool) {
	return runtime.GetG().GetOK(c)
}

func (c *tlsKey) Set(v interface{}) {
	runtime.GetG().Set(c, v)
}
