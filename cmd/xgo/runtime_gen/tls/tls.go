package tls

import (
	"sync"
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
	store   sync.Map // <goroutine ptr> -> interface{}
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
	key := uintptr(__xgo_link_getcurg())
	val, ok := c.store.Load(key)
	if !ok {
		return nil
	}
	return val
}

func (c *tlsKey) GetOK() (interface{}, bool) {
	key := uintptr(__xgo_link_getcurg())
	return c.store.Load(key)
}

func (c *tlsKey) Set(v interface{}) {
	key := uintptr(__xgo_link_getcurg())
	c.store.Store(key, v)
}
