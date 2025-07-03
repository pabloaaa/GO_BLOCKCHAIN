package tests

import (
	"net"
	"testing"
	"time"

	"github.com/pabloaaa/GO_BLOCKCHAIN/src"
)

func TestNewTcpConnectionManager(t *testing.T) {
	localPort := 50001
	manager := src.NewTcpConnectionManager(localPort)
	
	if manager == nil {
		t.Error("NewTcpConnectionManager returned nil")
	}
}

func TestCheckConnection_NoConnection(t *testing.T) {
	manager := src.NewTcpConnectionManager(50001)
	address := 50002
	
	result := manager.CheckConnection(address)
	if result {
		t.Error("CheckConnection should return false for non-existent connection")
	}
}

func TestConnectToNode_SelfConnection(t *testing.T) {
	localPort := 50001
	manager := src.NewTcpConnectionManager(localPort)
	
	conn, err := manager.ConnectToNode(localPort)
	if err == nil {
		t.Error("ConnectToNode should fail when trying to connect to self")
	}
	if conn != nil {
		t.Error("ConnectToNode should return nil connection when connecting to self")
	}
	
	expectedError := "cannot connect to self"
	if err.Error() != expectedError {
		t.Errorf("Expected error '%s', got '%s'", expectedError, err.Error())
	}
}

func TestConnectToNode_InvalidAddress(t *testing.T) {
	manager := src.NewTcpConnectionManager(50001)
	invalidAddress := 99999
	
	conn, err := manager.ConnectToNode(invalidAddress)
	if err == nil {
		t.Error("ConnectToNode should fail for invalid address")
	}
	if conn != nil {
		t.Error("ConnectToNode should return nil connection for invalid address")
	}
}

func TestAddSendingConnection_SelfConnection(t *testing.T) {
	localPort := 50001
	manager := src.NewTcpConnectionManager(localPort)
	
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatalf("Failed to create test listener: %v", err)
	}
	defer listener.Close()
	
	conn, err := net.Dial("tcp", listener.Addr().String())
	if err != nil {
		t.Fatalf("Failed to create test connection: %v", err)
	}
	defer conn.Close()
	
	manager.AddSendingConnection(localPort, conn)
	
	portMap := manager.GetPortMap()
	if len(portMap) != 0 {
		t.Error("AddSendingConnection should not add connection to self")
	}
}

func TestAddSendingConnection_ValidConnection(t *testing.T) {
	localPort := 50001
	remotePort := 50002
	manager := src.NewTcpConnectionManager(localPort)
	
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatalf("Failed to create test listener: %v", err)
	}
	defer listener.Close()
	
	conn, err := net.Dial("tcp", listener.Addr().String())
	if err != nil {
		t.Fatalf("Failed to create test connection: %v", err)
	}
	defer conn.Close()
	
	manager.AddSendingConnection(remotePort, conn)
	
	retrievedConn, exists := manager.GetSendingConnection(remotePort)
	if !exists {
		t.Error("AddSendingConnection should add the connection")
	}
	if retrievedConn != conn {
		t.Error("Retrieved connection should match the added connection")
	}
}

func TestRemoveConnection(t *testing.T) {
	localPort := 50001
	remotePort := 50002
	manager := src.NewTcpConnectionManager(localPort)
	
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatalf("Failed to create test listener: %v", err)
	}
	defer listener.Close()
	
	conn, err := net.Dial("tcp", listener.Addr().String())
	if err != nil {
		t.Fatalf("Failed to create test connection: %v", err)
	}
	
	manager.AddSendingConnection(remotePort, conn)
	manager.RemoveConnection(remotePort)
	
	_, exists := manager.GetSendingConnection(remotePort)
	if exists {
		t.Error("RemoveConnection should remove the connection")
	}
}

func TestGetSendingConnection_NonExistent(t *testing.T) {
	manager := src.NewTcpConnectionManager(50001)
	address := 50002
	
	conn, exists := manager.GetSendingConnection(address)
	if exists {
		t.Error("GetSendingConnection should return false for non-existent connection")
	}
	if conn != nil {
		t.Error("GetSendingConnection should return nil connection for non-existent address")
	}
}

func TestGetPortMap(t *testing.T) {
	localPort := 50001
	manager := src.NewTcpConnectionManager(localPort)
	
	portMap := manager.GetPortMap()
	if portMap == nil {
		t.Error("GetPortMap should not return nil")
	}
	
	if len(portMap) != 0 {
		t.Error("Initial port map should be empty")
	}
}

func TestCheckConnection_WithConnection(t *testing.T) {
	localPort := 50001
	remotePort := 50002
	manager := src.NewTcpConnectionManager(localPort)
	
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatalf("Failed to create test listener: %v", err)
	}
	defer listener.Close()
	
	conn, err := net.Dial("tcp", listener.Addr().String())
	if err != nil {
		t.Fatalf("Failed to create test connection: %v", err)
	}
	defer conn.Close()
	
	manager.AddSendingConnection(remotePort, conn)
	
	go func() {
		for {
			_, err := listener.Accept()
			if err != nil {
				return
			}
		}
	}()
	
	time.Sleep(100 * time.Millisecond)
	
	result := manager.CheckConnection(remotePort)
	if !result {
		t.Error("CheckConnection should return true for active connection")
	}
}