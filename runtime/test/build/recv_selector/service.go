// Integration test for ParseReceiverType fix: verifies xgo can instrument
// Go modules that reference types from other packages (e.g., net.Conn) without
// panicking. The fix handles *ast.SelectorExpr in method receiver parsing,
// which occurs when a receiver type is *otherpkg.Type.
package recv_selector

import "net"

type Service struct {
	conn net.Conn
}

func (s *Service) Handle() string {
	return s.conn.LocalAddr().String()
}
