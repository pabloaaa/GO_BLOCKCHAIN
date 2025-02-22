package src

import (
	"fmt"
	"log"
)

// MessageSender is an interface for sending messages.
type MessageSender interface {
	SendMsgToAddress(address int, data []byte, senderPort int) error
	SendMsg(data []byte) error
}

// TcpMessageSender sends messages over TCP.
type TcpMessageSender struct {
	connectionManager *TcpConnectionManager
}

// NewTCPSender creates a new TcpMessageSender.
func NewTCPSender(connectionManager *TcpConnectionManager) *TcpMessageSender {
	return &TcpMessageSender{connectionManager: connectionManager}
}

// SendMsgToAddress sends a message over TCP to the specified address.
func (s *TcpMessageSender) SendMsgToAddress(address int, data []byte, senderPort int) error {
	log.Printf("TcpMessageSender: Attempting to send message to address: %d from port: %d", address, senderPort)

	conn, exists := s.connectionManager.GetSendingConnection(address)
	if exists {
		// Sprawdź, czy połączenie jest aktywne
		if _, err := conn.Write([]byte{}); err != nil {
			log.Printf("TcpMessageSender: Existing connection to address %d is inactive, creating a new connection", address)
			s.connectionManager.RemoveConnection(address, senderPort)
			exists = false
		}
	}

	if !exists {
		var err error
		conn, err = s.connectionManager.ConnectToNode(address)
		if err != nil {
			return fmt.Errorf("TcpMessageSender: Failed to create a new sending port for listening port %d: %v", address, err)
		}
	}

	log.Printf("TcpMessageSender: Connection found, sending data to address %d", address)
	_, err := conn.Write(data)
	if err != nil {
		log.Printf("TcpMessageSender: Failed to send message: %v", err)
		s.connectionManager.RemoveConnection(address, senderPort)
	}
	return err
}

// SendMsg sends a message over TCP without specifying an address.
func (s *TcpMessageSender) SendMsg(data []byte) error {
	log.Println("TcpMessageSender: SendMsg method called without specifying an address")
	return nil
}
