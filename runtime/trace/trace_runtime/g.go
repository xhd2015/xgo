package trace_runtime

// keep in sync with runtime.__xgo_g
type G struct {
	goid       uint64
	parentGoID uint64

	gls                 map[interface{}]interface{}
	looseJsonMarshaling bool
}

func GetG() *G {
	return (*G)(Runtime_XgoGetCurG())
}

func (g *G) GoID() uint64 {
	return g.goid
}

func (g *G) ParentGoID() uint64 {
	return g.parentGoID
}

func (g *G) Set(key, value interface{}) {
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
