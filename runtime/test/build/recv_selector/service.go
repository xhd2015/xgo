package recv_selector

import "net"

type Service struct {
	conn net.Conn
}

func (s *Service) Handle() string {
	return s.conn.LocalAddr().String()
}
