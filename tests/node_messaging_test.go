package tests

import (
	"testing"

	"github.com/pabloaaa/GO_BLOCKCHAIN/src"
)

func TestAddPendingMessage(t *testing.T) {
	// Setup
	blockchain := src.NewBlockchain()
	connectionManager := src.NewTcpConnectionManager(50001)
	tcpMessageSender := src.NewTCPSender(connectionManager)
	node := src.NewNode(blockchain, 50001, tcpMessageSender, 50001)

	// Test adding a message
	testMessage := "Test message for blockchain"
	node.AddPendingMessage(testMessage)

	// Get pending messages
	messages := node.GetPendingMessages()

	// Verify
	if len(messages) != 1 {
		t.Errorf("Expected 1 pending message, got %d", len(messages))
	}

	if messages[0] != testMessage {
		t.Errorf("Expected message '%s', got '%s'", testMessage, messages[0])
	}
}

func TestGetPendingMessages_ClearsQueue(t *testing.T) {
	// Setup
	blockchain := src.NewBlockchain()
	connectionManager := src.NewTcpConnectionManager(50001)
	tcpMessageSender := src.NewTCPSender(connectionManager)
	node := src.NewNode(blockchain, 50001, tcpMessageSender, 50001)

	// Add multiple messages
	node.AddPendingMessage("Message 1")
	node.AddPendingMessage("Message 2")
	node.AddPendingMessage("Message 3")

	// Get pending messages (should clear the queue)
	messages := node.GetPendingMessages()

	// Verify messages were retrieved
	if len(messages) != 3 {
		t.Errorf("Expected 3 pending messages, got %d", len(messages))
	}

	// Verify queue is cleared
	remainingMessages := node.GetPendingMessages()
	if len(remainingMessages) != 0 {
		t.Errorf("Expected queue to be cleared, but got %d messages", len(remainingMessages))
	}
}

func TestGetPendingMessages_EmptyQueue(t *testing.T) {
	// Setup
	blockchain := src.NewBlockchain()
	connectionManager := src.NewTcpConnectionManager(50001)
	tcpMessageSender := src.NewTCPSender(connectionManager)
	node := src.NewNode(blockchain, 50001, tcpMessageSender, 50001)

	// Get pending messages from empty queue
	messages := node.GetPendingMessages()

	// Verify
	if len(messages) != 0 {
		t.Errorf("Expected 0 pending messages from empty queue, got %d", len(messages))
	}
}

func TestMultiplePendingMessages(t *testing.T) {
	// Setup
	blockchain := src.NewBlockchain()
	connectionManager := src.NewTcpConnectionManager(50001)
	tcpMessageSender := src.NewTCPSender(connectionManager)
	node := src.NewNode(blockchain, 50001, tcpMessageSender, 50001)

	// Add multiple messages
	expectedMessages := []string{
		"First message",
		"Second message",
		"Third message",
	}

	for _, msg := range expectedMessages {
		node.AddPendingMessage(msg)
	}

	// Get pending messages
	messages := node.GetPendingMessages()

	// Verify count
	if len(messages) != len(expectedMessages) {
		t.Errorf("Expected %d pending messages, got %d", len(expectedMessages), len(messages))
	}

	// Verify order and content
	for i, expectedMsg := range expectedMessages {
		if messages[i] != expectedMsg {
			t.Errorf("Expected message at index %d to be '%s', got '%s'", i, expectedMsg, messages[i])
		}
	}
}

func TestGetNextPendingMessage(t *testing.T) {
	// Setup
	blockchain := src.NewBlockchain()
	connectionManager := src.NewTcpConnectionManager(50001)
	tcpMessageSender := src.NewTCPSender(connectionManager)
	node := src.NewNode(blockchain, 50001, tcpMessageSender, 50001)

	// Add messages in order
	node.AddPendingMessage("Message 1")
	node.AddPendingMessage("Message 2")
	node.AddPendingMessage("Message 3")

	// Get messages one by one
	msg1 := node.GetNextPendingMessage()
	if msg1 != "Message 1" {
		t.Errorf("Expected 'Message 1', got '%s'", msg1)
	}

	msg2 := node.GetNextPendingMessage()
	if msg2 != "Message 2" {
		t.Errorf("Expected 'Message 2', got '%s'", msg2)
	}

	msg3 := node.GetNextPendingMessage()
	if msg3 != "Message 3" {
		t.Errorf("Expected 'Message 3', got '%s'", msg3)
	}

	// Queue should be empty now
	emptyMsg := node.GetNextPendingMessage()
	if emptyMsg != "" {
		t.Errorf("Expected empty string, got '%s'", emptyMsg)
	}
}

func TestGetPendingMessageCount(t *testing.T) {
	// Setup
	blockchain := src.NewBlockchain()
	connectionManager := src.NewTcpConnectionManager(50001)
	tcpMessageSender := src.NewTCPSender(connectionManager)
	node := src.NewNode(blockchain, 50001, tcpMessageSender, 50001)

	// Initially should be 0
	if count := node.GetPendingMessageCount(); count != 0 {
		t.Errorf("Expected 0 pending messages initially, got %d", count)
	}

	// Add one message
	node.AddPendingMessage("Message 1")
	if count := node.GetPendingMessageCount(); count != 1 {
		t.Errorf("Expected 1 pending message, got %d", count)
	}

	// Add more messages
	node.AddPendingMessage("Message 2")
	node.AddPendingMessage("Message 3")
	if count := node.GetPendingMessageCount(); count != 3 {
		t.Errorf("Expected 3 pending messages, got %d", count)
	}

	// Remove one message
	node.GetNextPendingMessage()
	if count := node.GetPendingMessageCount(); count != 2 {
		t.Errorf("Expected 2 pending messages after removing one, got %d", count)
	}

	// Clear all messages
	node.GetPendingMessages()
	if count := node.GetPendingMessageCount(); count != 0 {
		t.Errorf("Expected 0 pending messages after clearing, got %d", count)
	}
}