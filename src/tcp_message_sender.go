package src

import (
	"fmt"
	"log"
	"net"
)

// TcpMessageSender sends messages over TCP.
type TcpMessageSender struct {
	conn net.Conn
}

// NewTCPSender creates a new TcpMessageSender.
func NewTCPSender(address string) (*TcpMessageSender, error) {
	// Check if the address contains a port
	host, port, err := net.SplitHostPort(address)
	if err != nil {
		log.Printf("Invalid address format: %v", err)
		return nil, err
	}

	// If the port is 0, find an available port
	if port == "0" {
		ln, err := net.Listen("tcp", net.JoinHostPort(host, "0"))
		if err != nil {
			log.Printf("Failed to find an available port: %v", err)
			return nil, err
		}
		defer ln.Close()
		address = ln.Addr().String()
	}

	conn, err := net.Dial("tcp", address)
	if err != nil {
		log.Printf("Failed to dial TCP address %s: %v", address, err)
		return nil, err
	}
	return &TcpMessageSender{conn: conn}, nil
}

// SendMsg sends a message over TCP.
func (s *TcpMessageSender) SendMsg(data []byte) error {
	if s.conn == nil {
		log.Println("TCP connection is nil")
		return fmt.Errorf("TCP connection is nil")
	}
	log.Printf("Sending message: %x", data)
	_, err := s.conn.Write(data)
	if err != nil {
		log.Printf("Failed to send message: %v", err)
	}
	return err
}

func (s *TcpMessageSender) Close() {
	if s.conn != nil {
		s.conn.Close()
	}
}
