package src

import (
	"fmt"
	"net"
)

// MessageSender is an interface for sending messages.
type MessageSender interface {
	SendMsgToAddress(address int, data []byte) error
}

// TcpMessageSender sends messages over TCP.
type TcpMessageSender struct {
	connectionManager *TcpConnectionManager
	connections       map[int]net.Conn
}

// NewTCPSender creates a new TcpMessageSender.
func NewTCPSender(connectionManager *TcpConnectionManager) *TcpMessageSender {
	return &TcpMessageSender{
		connectionManager: connectionManager,
		connections:       make(map[int]net.Conn),
	}
}

// CheckConnection checks if the connection is still active.
func (s *TcpMessageSender) CheckConnection(address int) bool {
	conn, exists := s.connections[address]
	if !exists {
		Debug(fmt.Sprintf("TcpMessageSender: No existing connection for address %d", address))
		return false
	}
	if _, err := conn.Write([]byte{}); err != nil {
		Debug(fmt.Sprintf("TcpMessageSender: Connection to address %d is inactive: %v", address, err))
		delete(s.connections, address)
		return false
	}
	Debug(fmt.Sprintf("TcpMessageSender: Connection to address %d is active", address))
	return true
}

// SendMsgToAddress sends a message over TCP to the specified address.
func (s *TcpMessageSender) SendMsgToAddress(address int, data []byte) error {
	Debug(fmt.Sprintf("TcpMessageSender: Attempting to send message to address: %d", address))

	if !s.CheckConnection(address) {
		Debug(fmt.Sprintf("TcpMessageSender: No active connection for address %d, attempting to create a new one", address))
		conn, err := s.connectionManager.ConnectToNode(address)
		if err != nil {
			return fmt.Errorf("TcpMessageSender: Failed to create a new connection for address %d: %v", address, err)
		}
		s.connections[address] = conn
		Debug(fmt.Sprintf("TcpMessageSender: New connection created for address %d", address))
	}

	conn := s.connections[address]
	Debug(fmt.Sprintf("TcpMessageSender: Connection found, sending data to address %d", address))
	_, err := conn.Write(data)
	if err != nil {
		Error(fmt.Sprintf("TcpMessageSender: Failed to send message: %v", err))
		delete(s.connections, address)
	} else {
		Debug(fmt.Sprintf("TcpMessageSender: Message sent successfully to address %d", address))
	}
	return err
}
