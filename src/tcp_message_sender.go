package src

import (
	"net"
)

type TcpMessageSender struct {
	conn net.Conn
}

func NewTCPSender(address string) (*TcpMessageSender, error) {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return nil, err
	}
	return &TcpMessageSender{conn: conn}, nil
}

func (s *TcpMessageSender) SendMsg(data []byte) error {
	_, err := s.conn.Write(data)
	return err
}

func (s *TcpMessageSender) Close() {
	s.conn.Close()
}
