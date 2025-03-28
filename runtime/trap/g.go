package trap

import (
	"github.com/xhd2015/xgo/runtime/internal/runtime"
)

// keep in sync with runtime.__xgo_g
type _G struct {
	gls                 map[interface{}]interface{}
	looseJsonMarshaling bool
}

func _GetG() *_G {
	return (*_G)(runtime.XgoGetCurG())
}

func (g *_G) Set(key, value interface{}) {
	if g.gls == nil {
		g.gls = make(map[interface{}]interface{})
	}
	g.gls[key] = value
}

func (g *_G) Get(key interface{}) interface{} {
	return g.gls[key]
}

func (g *_G) GetOK(key interface{}) (interface{}, bool) {
	v, ok := g.gls[key]
	return v, ok
}

func (g *_G) Delete(key interface{}) {
	delete(g.gls, key)
}
