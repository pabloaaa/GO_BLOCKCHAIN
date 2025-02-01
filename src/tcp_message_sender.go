package src

import (
	"log"
	"net"
)

// TcpMessageSender sends messages over TCP.
type TcpMessageSender struct {
}

// NewTCPSender creates a new TcpMessageSender.
func NewTCPSender() *TcpMessageSender {
	return &TcpMessageSender{}
}

// SendMsgToAddress sends a message over TCP to the specified address.
func (s *TcpMessageSender) SendMsgToAddress(address []byte, data []byte) error {
	conn, err := net.Dial("tcp", string(address))
	if err != nil {
		log.Printf("Failed to dial TCP address %s: %v", address, err)
		return err
	}
	defer conn.Close()

	_, err = conn.Write(data)
	if err != nil {
		log.Printf("Failed to send message: %v", err)
	}
	return err
}

// SendMsg sends a message over TCP without specifying an address.
func (s *TcpMessageSender) SendMsg(data []byte) error {
	// Implementation for sending messages without specifying an address
	log.Println("SendMsg method called without specifying an address")
	return nil
}
