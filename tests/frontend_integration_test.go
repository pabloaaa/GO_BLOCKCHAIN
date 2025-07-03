package tests

import (
	"fmt"
	"testing"
	"time"

	"github.com/pabloaaa/GO_BLOCKCHAIN/src"
	"github.com/pabloaaa/GO_BLOCKCHAIN/types"
)

func TestFrontendIntegration(t *testing.T) {
	src.SetLogLevel("info")

	// Setup dwóch nodów
	blockchain1 := src.NewBlockchain()
	connectionManager1 := src.NewTcpConnectionManager(50001)
	tcpMessageSender1 := src.NewTCPSender(connectionManager1)
	node1 := src.NewNode(blockchain1, 50001, tcpMessageSender1, 50001)

	blockchain2 := src.NewBlockchain()
	connectionManager2 := src.NewTcpConnectionManager(50002)
	tcpMessageSender2 := src.NewTCPSender(connectionManager2)
	node2 := src.NewNode(blockchain2, 50002, tcpMessageSender2, 50001)

	// Start TCP servers
	go node1.Start()
	go node2.Start()

	// Wait for connection
	time.Sleep(2 * time.Second)

	// Test 1: Symulacja wysyłania wiadomości przez frontend node1
	src.Info("=== TEST 1: Node1 sends message through 'frontend' ===")
	
	// Check initial reward
	initialReward1 := node1.GetBlockchain().GetReward()
	src.Info(fmt.Sprintf("Node1 initial reward: %d", initialReward1))
	
	// Ensure node has enough reward
	if initialReward1 < 10 {
		node1.GetBlockchain().RewardNode(node1.GetAddress(), 50) // Add reward for testing
	}
	
	// Simulate frontend message sending
	testMessage := "Hello from Node 50001!"
	formattedMessage := fmt.Sprintf("NODE %d said: %s", node1.GetAddress(), testMessage)
	
	// Deduct reward (like frontend would do)
	err := node1.GetBlockchain().RewardNode(node1.GetAddress(), -10)
	if err != nil {
		t.Fatalf("Failed to deduct reward: %v", err)
	}
	
	// Add message to pending queue (like frontend would do)
	node1.AddPendingMessage(formattedMessage)
	src.Info(fmt.Sprintf("Message added to node1 pending queue: %s", formattedMessage))
	
	// Check pending count
	pendingCount := node1.GetPendingMessageCount()
	if pendingCount != 1 {
		t.Errorf("Expected 1 pending message, got %d", pendingCount)
	}
	src.Info(fmt.Sprintf("Node1 pending messages: %d", pendingCount))

	// Test 2: Mine block with message
	src.Info("=== TEST 2: Mining block with message ===")
	
	// Mine a block that should include the message
	for attempt := 0; attempt < 20; attempt++ {
		node1.TryToFindNewBlock()
		
		// Check if message appeared in blockchain
		var foundMessage bool
		node1.GetBlockchain().TraverseTree(func(blockNode *types.BlockNode) bool {
			for _, msg := range blockNode.Block.Messages {
				if msg == formattedMessage {
					foundMessage = true
					src.Info(fmt.Sprintf("Message found in blockchain: %s", msg))
				}
			}
			return false
		})
		
		if foundMessage {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	
	// Verify message is in blockchain
	var blockchainMessages []string
	node1.GetBlockchain().TraverseTree(func(blockNode *types.BlockNode) bool {
		for _, msg := range blockNode.Block.Messages {
			blockchainMessages = append(blockchainMessages, msg)
		}
		return false
	})
	
	if len(blockchainMessages) != 1 {
		t.Errorf("Expected 1 message in blockchain, got %d", len(blockchainMessages))
	}
	
	if blockchainMessages[0] != formattedMessage {
		t.Errorf("Expected message '%s', got '%s'", formattedMessage, blockchainMessages[0])
	}
	
	// Check pending count is now 0
	pendingCount = node1.GetPendingMessageCount()
	if pendingCount != 0 {
		t.Errorf("Expected 0 pending messages after mining, got %d", pendingCount)
	}

	// Test 3: Synchronize node2 and verify it sees the message
	src.Info("=== TEST 3: Node2 synchronization ===")
	
	time.Sleep(2 * time.Second) // Wait for automatic sync
	
	var node2Messages []string
	node2.GetBlockchain().TraverseTree(func(blockNode *types.BlockNode) bool {
		for _, msg := range blockNode.Block.Messages {
			node2Messages = append(node2Messages, msg)
		}
		return false
	})
	
	if len(node2Messages) != 1 {
		t.Errorf("Node2 should have 1 message after sync, got %d", len(node2Messages))
	}
	
	if len(node2Messages) > 0 && node2Messages[0] != formattedMessage {
		t.Errorf("Node2 message mismatch. Expected '%s', got '%s'", formattedMessage, node2Messages[0])
	}

	// Test 4: Simulate node2 sending a message
	src.Info("=== TEST 4: Node2 sends message ===")
	
	// Ensure node2 has reward
	if node2.GetBlockchain().GetReward() < 10 {
		node2.GetBlockchain().RewardNode(node2.GetAddress(), 50)
	}
	
	testMessage2 := "Hello from Node 50002!"
	formattedMessage2 := fmt.Sprintf("NODE %d said: %s", node2.GetAddress(), testMessage2)
	
	// Deduct reward and add message
	node2.GetBlockchain().RewardNode(node2.GetAddress(), -10)
	node2.AddPendingMessage(formattedMessage2)
	
	// Mine block on node2
	for attempt := 0; attempt < 20; attempt++ {
		node2.TryToFindNewBlock()
		
		var foundSecondMessage bool
		node2.GetBlockchain().TraverseTree(func(blockNode *types.BlockNode) bool {
			for _, msg := range blockNode.Block.Messages {
				if msg == formattedMessage2 {
					foundSecondMessage = true
				}
			}
			return false
		})
		
		if foundSecondMessage {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	// Test 5: Verify both nodes see both messages
	src.Info("=== TEST 5: Final verification ===")
	
	time.Sleep(2 * time.Second) // Wait for sync
	
	// Check node1 sees both messages
	var finalNode1Messages []string
	node1.GetBlockchain().TraverseTree(func(blockNode *types.BlockNode) bool {
		for _, msg := range blockNode.Block.Messages {
			finalNode1Messages = append(finalNode1Messages, msg)
		}
		return false
	})
	
	// Check node2 sees both messages  
	var finalNode2Messages []string
	node2.GetBlockchain().TraverseTree(func(blockNode *types.BlockNode) bool {
		for _, msg := range blockNode.Block.Messages {
			finalNode2Messages = append(finalNode2Messages, msg)
		}
		return false
	})
	
	src.Info(fmt.Sprintf("Node1 final messages: %v", finalNode1Messages))
	src.Info(fmt.Sprintf("Node2 final messages: %v", finalNode2Messages))
	
	// Both nodes should see 2 messages
	if len(finalNode1Messages) != 2 || len(finalNode2Messages) != 2 {
		t.Errorf("Both nodes should have 2 messages. Node1: %d, Node2: %d", 
			len(finalNode1Messages), len(finalNode2Messages))
	}

	src.Info("=== FRONTEND INTEGRATION TEST PASSED! ===")
}