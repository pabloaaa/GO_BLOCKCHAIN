package src

import (
	"fmt"
	"log"
	"math/rand"
	"net"
	"sync"
)

// TcpConnectionManager manages TCP connections.
type TcpConnectionManager struct {
	portMap   map[int]map[int]net.Conn
	mux       sync.Mutex
	localPort int
}

// NewTcpConnectionManager creates a new TcpConnectionManager.
func NewTcpConnectionManager(localPort int) *TcpConnectionManager {
	return &TcpConnectionManager{
		portMap:   make(map[int]map[int]net.Conn),
		localPort: localPort,
	}
}

// ConnectToNode establishes a TCP connection to the specified address using a random available port.
func (m *TcpConnectionManager) ConnectToNode(address int) (net.Conn, error) {
	if address == m.localPort {
		log.Printf("TcpConnectionManager: Skipping connection to self at address %d", address)
		return nil, fmt.Errorf("cannot connect to self")
	}

	log.Printf("TcpConnectionManager: Attempting to connect to node at address %d", address)
	tcpAddr, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("localhost:%d", address))
	if err != nil {
		log.Printf("TcpConnectionManager: Failed to resolve TCP address %d: %v", address, err)
		return nil, err
	}

	localAddr, err := m.getRandomAvailableLocalAddr()
	if err != nil {
		log.Printf("TcpConnectionManager: Failed to get available local address: %v", err)
		return nil, err
	}

	conn, err := net.DialTCP("tcp", localAddr, tcpAddr)
	if err != nil {
		log.Printf("TcpConnectionManager: Failed to connect to node at address %d: %v", address, err)
		return nil, err
	}
	log.Printf("TcpConnectionManager: Successfully connected to node at address %d using port %d", address, localAddr.Port)
	m.AddSendingConnection(address, localAddr.Port, conn)
	return conn, nil
}

// getRandomAvailableLocalAddr returns a random available local address.
func (m *TcpConnectionManager) getRandomAvailableLocalAddr() (*net.TCPAddr, error) {
	for {
		randomPort := rand.Intn(65535-1024) + 1024
		localAddr, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("localhost:%d", randomPort))
		if err != nil {
			log.Printf("TcpConnectionManager: Failed to resolve local TCP address: %v", err)
			continue
		}

		listener, err := net.ListenTCP("tcp", localAddr)
		if err == nil {
			listener.Close()
			return localAddr, nil
		}
	}
}

// AddSendingConnection adds a TCP connection to the sending connections map.
func (m *TcpConnectionManager) AddSendingConnection(addressPort, localPort int, conn net.Conn) {
	if addressPort == m.localPort || localPort == m.localPort {
		log.Printf("TcpConnectionManager: Skipping adding connection to self from address %d or local port %d", addressPort, localPort)
		return
	}

	m.mux.Lock()
	defer m.mux.Unlock()
	if _, exists := m.portMap[addressPort]; !exists {
		m.portMap[addressPort] = make(map[int]net.Conn)
	}
	m.portMap[addressPort][localPort] = conn
	log.Printf("TcpConnectionManager: Added sending connection to address %d on port %d", addressPort, localPort)
}

// RemoveConnection removes a TCP connection from the manager and closes it.
func (m *TcpConnectionManager) RemoveConnection(addressPort, localPort int) {
	m.mux.Lock()
	defer m.mux.Unlock()
	if connections, exists := m.portMap[addressPort]; exists {
		if conn, exists := connections[localPort]; exists {
			conn.Close()
			delete(connections, localPort)
			log.Printf("TcpConnectionManager: Removed connection to address %d on port %d", addressPort, localPort)
		}
	}
}

// GetSendingConnection retrieves a sending TCP connection by address.
func (m *TcpConnectionManager) GetSendingConnection(addressPort int) (net.Conn, bool) {
	m.mux.Lock()
	defer m.mux.Unlock()
	for _, connections := range m.portMap {
		if conn, exists := connections[addressPort]; exists {
			return conn, true
		}
	}
	return nil, false
}

// PortMap returns the port map.
func (m *TcpConnectionManager) GetPortMap() map[int]map[int]net.Conn {
	m.mux.Lock()
	defer m.mux.Unlock()
	return m.portMap
}
