package src

import (
	"fmt"
	"log"
	"net"
	"sync"
)

// TcpConnectionManager manages TCP connections.
type TcpConnectionManager struct {
	connections map[int]net.Conn
	mux         sync.Mutex
	localPort   int
}

// NewTcpConnectionManager creates a new TcpConnectionManager.
func NewTcpConnectionManager(localPort int) *TcpConnectionManager {
	return &TcpConnectionManager{
		connections: make(map[int]net.Conn),
		localPort:   localPort,
	}
}

// CheckConnection checks if the connection is still active.
func (m *TcpConnectionManager) CheckConnection(address int) bool {
	m.mux.Lock()
	defer m.mux.Unlock()
	conn, exists := m.connections[address]
	if !exists {
		log.Printf("TcpConnectionManager: No existing connection for address %d", address)
		return false
	}
	if _, err := conn.Write([]byte{}); err != nil {
		log.Printf("TcpConnectionManager: Connection to address %d is inactive: %v", address, err)
		delete(m.connections, address)
		return false
	}
	log.Printf("TcpConnectionManager: Connection to address %d is active", address)
	return true
}

// ConnectToNode establishes a TCP connection to the specified address using a random available port.
func (m *TcpConnectionManager) ConnectToNode(address int) (net.Conn, error) {
	if address == m.localPort {
		log.Printf("TcpConnectionManager: Skipping connection to self at address %d", address)
		return nil, fmt.Errorf("cannot connect to self")
	}

	m.mux.Lock()
	defer m.mux.Unlock()

	// Sprawdzenie, czy istnieje aktywne połączenie
	if conn, exists := m.connections[address]; exists {
		if _, err := conn.Write([]byte{}); err == nil {
			log.Printf("TcpConnectionManager: Existing connection to address %d is active", address)
			return conn, nil
		}
		log.Printf("TcpConnectionManager: Existing connection to address %d is inactive, removing it", address)
		conn.Close()
		delete(m.connections, address)
	}

	log.Printf("TcpConnectionManager: Attempting to connect to node at address %d", address)
	tcpAddr, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("localhost:%d", address))
	if err != nil {
		log.Printf("TcpConnectionManager: Failed to resolve TCP address %d: %v", address, err)
		return nil, err
	}

	localAddr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		log.Printf("TcpConnectionManager: Failed to resolve local TCP address: %v", err)
		return nil, err
	}

	log.Printf("TcpConnectionManager: Local address resolved: %v", localAddr)
	log.Printf("TcpConnectionManager: Remote address resolved: %v", tcpAddr)

	conn, err := net.DialTCP("tcp", localAddr, tcpAddr)
	if err != nil {
		log.Printf("TcpConnectionManager: Failed to connect to node at address %d: %v", address, err)
		return nil, err
	}
	log.Printf("TcpConnectionManager: Successfully connected to node at address %d", address)
	m.connections[address] = conn
	return conn, nil
}

// AddSendingConnection adds a TCP connection to the sending connections map.
func (m *TcpConnectionManager) AddSendingConnection(address int, conn net.Conn) {
	if address == m.localPort {
		log.Printf("TcpConnectionManager: Skipping adding connection to self from address %d", address)
		return
	}

	m.mux.Lock()
	defer m.mux.Unlock()
	m.connections[address] = conn
	log.Printf("TcpConnectionManager: Added sending connection to address %d", address)
}

// RemoveConnection removes a TCP connection from the manager and closes it.
func (m *TcpConnectionManager) RemoveConnection(address int) {
	m.mux.Lock()
	defer m.mux.Unlock()
	if conn, exists := m.connections[address]; exists {
		conn.Close()
		delete(m.connections, address)
		log.Printf("TcpConnectionManager: Removed connection to address %d", address)
	}
}

// GetSendingConnection retrieves a sending TCP connection by address.
func (m *TcpConnectionManager) GetSendingConnection(address int) (net.Conn, bool) {
	m.mux.Lock()
	defer m.mux.Unlock()
	conn, exists := m.connections[address]
	return conn, exists
}

// GetPortMap returns the connections map.
func (m *TcpConnectionManager) GetPortMap() map[int]net.Conn {
	m.mux.Lock()
	defer m.mux.Unlock()
	return m.connections
}
