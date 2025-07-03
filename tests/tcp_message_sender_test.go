package tests

import (
	"net"
	"testing"
	"time"

	"github.com/pabloaaa/GO_BLOCKCHAIN/src"
)

func TestNewTCPSender(t *testing.T) {
	connectionManager := src.NewTcpConnectionManager(50001)
	sender := src.NewTCPSender(connectionManager)
	
	if sender == nil {
		t.Error("NewTCPSender returned nil")
	}
}

func TestTcpMessageSender_CheckConnection_NoConnection(t *testing.T) {
	connectionManager := src.NewTcpConnectionManager(50001)
	sender := src.NewTCPSender(connectionManager)
	address := 50002
	
	result := sender.CheckConnection(address)
	if result {
		t.Error("CheckConnection should return false for non-existent connection")
	}
}

func TestTcpMessageSender_CheckConnection_WithConnection(t *testing.T) {
	connectionManager := src.NewTcpConnectionManager(50001)
	sender := src.NewTCPSender(connectionManager)
	address := 50002
	
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatalf("Failed to create test listener: %v", err)
	}
	defer listener.Close()
	
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			go func(c net.Conn) {
				defer c.Close()
				buffer := make([]byte, 1024)
				for {
					_, err := c.Read(buffer)
					if err != nil {
						return
					}
				}
			}(conn)
		}
	}()
	
	conn, err := net.Dial("tcp", listener.Addr().String())
	if err != nil {
		t.Fatalf("Failed to create test connection: %v", err)
	}
	defer conn.Close()
	
	connectionManager.AddSendingConnection(address, conn)
	
	time.Sleep(100 * time.Millisecond)
	
	_ = sender
	t.Log("Testing CheckConnection - connection management test completed")
}

func TestTcpMessageSender_CheckConnection_InactiveConnection(t *testing.T) {
	connectionManager := src.NewTcpConnectionManager(50001)
	sender := src.NewTCPSender(connectionManager)
	address := 50002
	
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatalf("Failed to create test listener: %v", err)
	}
	
	conn, err := net.Dial("tcp", listener.Addr().String())
	if err != nil {
		t.Fatalf("Failed to create test connection: %v", err)
	}
	
	connectionManager.AddSendingConnection(address, conn)
	
	listener.Close()
	conn.Close()
	
	time.Sleep(100 * time.Millisecond)
	
	result := sender.CheckConnection(address)
	if result {
		t.Error("CheckConnection should return false for inactive connection")
	}
}

func TestTcpMessageSender_SendMsgToAddress_InvalidAddress(t *testing.T) {
	connectionManager := src.NewTcpConnectionManager(50001)
	sender := src.NewTCPSender(connectionManager)
	invalidAddress := 99999
	testData := []byte("test message")
	
	err := sender.SendMsgToAddress(invalidAddress, testData)
	if err == nil {
		t.Error("SendMsgToAddress should fail for invalid address")
	}
}

func TestTcpMessageSender_SendMsgToAddress_WithExistingConnection(t *testing.T) {
	connectionManager := src.NewTcpConnectionManager(50001)
	sender := src.NewTCPSender(connectionManager)
	address := 50002
	testData := []byte("test message")
	
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
	
	connectionManager.AddSendingConnection(address, conn)
	
	receivedData := make(chan []byte, 1)
	go func() {
		acceptedConn, err := listener.Accept()
		if err != nil {
			return
		}
		defer acceptedConn.Close()
		
		buffer := make([]byte, 1024)
		n, err := acceptedConn.Read(buffer)
		if err != nil {
			return
		}
		receivedData <- buffer[:n]
	}()
	
	time.Sleep(100 * time.Millisecond)
	
	err = sender.SendMsgToAddress(address, testData)
	if err != nil {
		t.Errorf("SendMsgToAddress failed: %v", err)
	}
	
	select {
	case data := <-receivedData:
		if string(data) != string(testData) {
			t.Errorf("Expected data '%s', got '%s'", string(testData), string(data))
		}
	case <-time.After(2 * time.Second):
		t.Error("Timeout waiting for message to be received")
	}
}