package patch_var_struct_field_ptr

import "sync"

var pm ProviderManager

type ProviderManager struct {
	providers sync.Map // namespace->Provider
}

func getProvider(namespace string) interface{} {
	val, ok := pm.providers.Load(namespace)
	if ok {
		return val
	}

	p := newProvider(namespace)
	pm.providers.LoadOrStore(namespace, p)
	return p
}

func newProvider(namespace string) interface{} {
	return namespace
}

type ProviderFactory struct {
	ProviderManager ProviderManager
}

var pf ProviderFactory

func getFactoryprovider(namespace string) interface{} {
	val, _ := pf.ProviderManager.providers.Load(namespace)
	return val
}
