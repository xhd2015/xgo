// Verifies that xgo can instrument a package whose source references types
// from another package (net.Conn) in struct fields — this exercises the
// CollectDecls → ParseReceiverType code path without hitting the
// *ast.SelectorExpr panic that was fixed in instrument/ast/recv.go.
package recv_selector

import (
	"testing"
)

func TestMethodWithExternalTypeFields(t *testing.T) {
	s := &Service{}
	_ = s
}

