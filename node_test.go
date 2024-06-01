package main

import (
	"net"
	"testing"
	"time"
)

func TestNewNode(t *testing.T) {
	bc := NewBlockchain()
	newBlockReceiver := make(chan Block)

	node := NewNode(bc, newBlockReceiver)

	if node.blockchain != bc {
		t.Errorf("Expected node blockchain to be same as bc, but got different value")
	}

	if len(node.clients) != 0 {
		t.Errorf("Expected node clients length to be 0, but got %d", len(node.clients))
	}

	if node.newBlockReceiver != newBlockReceiver {
		t.Errorf("Expected node newBlockReceiver to be same as newBlockReceiver, but got different value")
	}
}

func TestGetBlockchain(t *testing.T) {
	bc := NewBlockchain()
	newBlockReceiver := make(chan Block)

	node := NewNode(bc, newBlockReceiver)

	if node.GetBlockchain() != bc {
		t.Errorf("Expected GetBlockchain to return initial blockchain, but got different value")
	}
}

// This is a mock test for Start method. In real scenario, you should use interface and mock the dependencies.
func TestStart(t *testing.T) {
	bc := NewBlockchain()
	newBlockReceiver := make(chan Block)

	node := NewNode(bc, newBlockReceiver)

	go func() {
		err := node.Start("localhost:8000")
		if err != nil {
			t.Errorf("Expected Start to not return error, but got %v", err)
		}
	}()

	// Wait for the server to start
	time.Sleep(time.Second)

	// Connect to the server
	conn, err := net.Dial("tcp", "localhost:8000")
	if err != nil {
		t.Errorf("Expected to connect to node, but got error: %v", err)
	}
	defer conn.Close()

	// Wait for the connection to be added
	time.Sleep(time.Second)

	if len(node.clients) != 1 {
		t.Errorf("Expected node clients length to be 1, but got %d", len(node.clients))
	}
}
