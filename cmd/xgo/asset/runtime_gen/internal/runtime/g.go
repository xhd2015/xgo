package runtime

import "unsafe"

// keep in sync with runtime.__xgo_g
type G struct {
	gls                 map[interface{}]interface{}
	looseJsonMarshaling bool

	// when handling trapping, prohibit
	// another trapping from the handler
	trappingDepth int
}

func GetG() *G {
	return (*G)(XgoGetCurG())
}

func AsG(g unsafe.Pointer) *G {
	return (*G)(g)
}

func (g *G) Set(key, value interface{}) {
	if g.gls == nil {
		g.gls = make(map[interface{}]interface{})
	}
	g.gls[key] = value
}

func (g *G) Get(key interface{}) interface{} {
	return g.gls[key]
}

func (g *G) GetOK(key interface{}) (interface{}, bool) {
	v, ok := g.gls[key]
	return v, ok
}

func (g *G) Delete(key interface{}) {
	delete(g.gls, key)
}

func (g *G) IncTrappingDepth() int {
	g.trappingDepth++
	return g.trappingDepth
}

func (g *G) DecTrappingDepth() {
	g.trappingDepth--
}
