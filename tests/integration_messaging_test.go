package tests

import (
	"fmt"
	"testing"
	"time"

	"github.com/pabloaaa/GO_BLOCKCHAIN/src"
	"github.com/pabloaaa/GO_BLOCKCHAIN/types"
)

func TestTwoNodesMessaging(t *testing.T) {
	src.SetLogLevel("info")

	// Setup dwóch nodów
	blockchain1 := src.NewBlockchain()
	connectionManager1 := src.NewTcpConnectionManager(50001)
	tcpMessageSender1 := src.NewTCPSender(connectionManager1)
	node1 := src.NewNode(blockchain1, 50001, tcpMessageSender1, 50001)

	blockchain2 := src.NewBlockchain()
	connectionManager2 := src.NewTcpConnectionManager(50002)
	tcpMessageSender2 := src.NewTCPSender(connectionManager2)
	node2 := src.NewNode(blockchain2, 50002, tcpMessageSender2, 50001) // Connect to node1

	// Start TCP servers
	go node1.Start()
	go node2.Start()

	// Wait for nodes to connect
	time.Sleep(2 * time.Second)

	// Synchronize nodes
	err := node2.SyncNodes(50001)
	if err != nil {
		t.Logf("Node sync warning: %v", err)
	}

	time.Sleep(1 * time.Second)

	// Add 3 messages to node1
	testMessages := []string{
		"NODE 50001 said: Message 1",
		"NODE 50001 said: Message 2", 
		"NODE 50001 said: Message 3",
	}

	for _, msg := range testMessages {
		node1.AddPendingMessage(msg)
		src.Info(fmt.Sprintf("Test: Added message to node1: %s", msg))
	}

	// Let node1 mine blocks with messages
	src.Info("Test: Starting block mining for node1...")
	
	var foundBlocks []*types.Block
	maxAttempts := 50
	
	for attempt := 0; attempt < maxAttempts && len(foundBlocks) < 3; attempt++ {
		// Try to find a new block
		node1.TryToFindNewBlock()
		
		// Check if a new block was added
		var currentBlocks []*types.Block
		node1.GetBlockchain().TraverseTree(func(blockNode *types.BlockNode) bool {
			if len(blockNode.Block.Messages) > 0 {
				currentBlocks = append(currentBlocks, blockNode.Block)
			}
			return false
		})
		
		if len(currentBlocks) > len(foundBlocks) {
			foundBlocks = currentBlocks
			src.Info(fmt.Sprintf("Test: Found block with message, total blocks with messages: %d", len(foundBlocks)))
		}
		
		time.Sleep(100 * time.Millisecond)
	}

	// Verify that 3 blocks with messages were created
	if len(foundBlocks) != 3 {
		t.Errorf("Expected 3 blocks with messages, got %d", len(foundBlocks))
		return
	}

	// Verify messages are in correct order and only one per block
	for i, block := range foundBlocks {
		if len(block.Messages) != 1 {
			t.Errorf("Block %d should have exactly 1 message, got %d", i, len(block.Messages))
			continue
		}
		
		expectedMessage := testMessages[i]
		if block.Messages[0] != expectedMessage {
			t.Errorf("Block %d message mismatch. Expected '%s', got '%s'", 
				i, expectedMessage, block.Messages[0])
		}
		
		src.Info(fmt.Sprintf("Test: Block %d confirmed with message: %s", i, block.Messages[0]))
	}

	// Now sync node2 to get the blocks from node1
	src.Info("Test: Synchronizing node2 with node1...")
	err = node2.SyncNodes(50001)
	if err != nil {
		t.Logf("Node sync warning: %v", err)
	}

	time.Sleep(2 * time.Second)

	// Verify node2 has the same messages in blockchain
	var node2Blocks []*types.Block
	node2.GetBlockchain().TraverseTree(func(blockNode *types.BlockNode) bool {
		if len(blockNode.Block.Messages) > 0 {
			node2Blocks = append(node2Blocks, blockNode.Block)
		}
		return false
	})

	if len(node2Blocks) != 3 {
		t.Errorf("Node2 should have 3 blocks with messages after sync, got %d", len(node2Blocks))
		return
	}

	// Verify messages are the same on both nodes
	for i := 0; i < 3; i++ {
		if node2Blocks[i].Messages[0] != foundBlocks[i].Messages[0] {
			t.Errorf("Message %d differs between nodes. Node1: '%s', Node2: '%s'",
				i, foundBlocks[i].Messages[0], node2Blocks[i].Messages[0])
		}
	}

	src.Info("Test: Successfully verified that both nodes have the same messages in blockchain!")

	// Test frontend visibility by simulating HTTP requests
	testFrontendVisibility(t, node1, node2, testMessages)
}

func testFrontendVisibility(t *testing.T, node1, node2 *src.Node, expectedMessages []string) {
	src.Info("Test: Testing frontend visibility...")

	// Function to get chat history from node (simulate HTTP call)
	getChatHistory := func(node *src.Node, nodeName string) []string {
		var blockchainMessages []string
		node.GetBlockchain().TraverseTree(func(blockNode *types.BlockNode) bool {
			for _, msg := range blockNode.Block.Messages {
				blockchainMessages = append(blockchainMessages, msg + " (confirmed)")
			}
			return false
		})
		
		src.Info(fmt.Sprintf("Test: %s chat history: %v", nodeName, blockchainMessages))
		return blockchainMessages
	}

	// Get chat history from both nodes
	node1Messages := getChatHistory(node1, "Node1")
	node2Messages := getChatHistory(node2, "Node2")

	// Verify both nodes show the same messages
	if len(node1Messages) != len(node2Messages) {
		t.Errorf("Frontend visibility test failed: Node1 has %d messages, Node2 has %d messages",
			len(node1Messages), len(node2Messages))
		return
	}

	if len(node1Messages) != 3 {
		t.Errorf("Frontend visibility test failed: Expected 3 messages on frontend, got %d", len(node1Messages))
		return
	}

	// Verify message content (they should be marked as confirmed)
	for i, expectedMsg := range expectedMessages {
		expectedFrontendMsg := expectedMsg + " (confirmed)"
		
		if node1Messages[i] != expectedFrontendMsg {
			t.Errorf("Node1 frontend message %d mismatch. Expected '%s', got '%s'",
				i, expectedFrontendMsg, node1Messages[i])
		}
		
		if node2Messages[i] != expectedFrontendMsg {
			t.Errorf("Node2 frontend message %d mismatch. Expected '%s', got '%s'",
				i, expectedFrontendMsg, node2Messages[i])
		}
	}

	src.Info("Test: Frontend visibility test passed! Both nodes show the same confirmed messages.")
}