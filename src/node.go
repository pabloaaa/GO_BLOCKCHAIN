package src

import (
	"fmt"
	"net"
	"sync"

	"github.com/pabloaaa/GO_BLOCKCHAIN/interfaces"
	block_chain "github.com/pabloaaa/GO_BLOCKCHAIN/protos"
	"google.golang.org/protobuf/proto"
)

// Node represents a node in the blockchain network.
type Node struct {
	blockchain           interfaces.BlockchainInterface
	nodes                []int
	blockHandler         interfaces.BlockMessageHandlerInterface
	nodeHandler          interfaces.NodeMessageHandlerInterface
	tcpMessageSender     *TcpMessageSender
	tcpConnectionManager *TcpConnectionManager
	address              int
	mux                  sync.Mutex
	messageFactory       *MessageFactory
	pendingMessages      []string // Queue for messages to be included in blocks
	messageMux           sync.Mutex // Mutex for pendingMessages
}

// NewNode creates a new Node.
func NewNode(blockchain interfaces.BlockchainInterface, address int, tcpMessageSender *TcpMessageSender, bootstrapAddress int) *Node {
	Info(fmt.Sprintf("Node: Initializing node with address: %d", address))

	connectionManager := NewTcpConnectionManager(address)
	node := &Node{
		blockchain:           blockchain,
		nodes:                make([]int, 0),
		tcpMessageSender:     tcpMessageSender,
		tcpConnectionManager: connectionManager,
		address:              address,
		pendingMessages:      make([]string, 0),
	}
	node.blockHandler = NewBlockMessageHandler(blockchain, node.tcpMessageSender, address)
	node.nodeHandler = NewNodeMessageHandler(node.tcpMessageSender, &node.nodes, address)

	// If bootstrapAddress is provided and not the same as node address, add it to the list of nodes
	if bootstrapAddress > 0 && bootstrapAddress != address {
		node.nodes = append(node.nodes, bootstrapAddress)
	}
	return node
}

// GetBlockchain returns the blockchain interface.
func (n *Node) GetBlockchain() interfaces.BlockchainInterface {
	return n.blockchain
}

// GetNodes returns the list of connected nodes.
func (n *Node) GetNodes() []int {
	return n.nodes
}

// GetAddress returns the address of the node.
func (n *Node) GetAddress() int {
	return n.address
}

// GetPortMap returns the port map from TcpConnectionManager.
func (n *Node) GetPortMap() map[int]net.Conn {
	return n.tcpConnectionManager.GetPortMap()
}

// Start starts the node and listens for incoming connections.
func (n *Node) Start() {

	ln, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", n.address))
	if err != nil {
		Error(fmt.Sprintf("Node: Failed to listen on address %d: %v", n.address, err))
		return
	}
	defer ln.Close()

	// Check if node.address is not equal to bootstrapAddress
	if len(n.nodes) > 0 && n.nodes[0] != n.address {
		Debug(fmt.Sprintf("Node: Broadcasting address to nodes: %v", n.nodes))
		n.nodeHandler.BroadcastAddress(n.nodes, n.address)
	}

	for {
		Debug("Node: Waiting for incoming connections...")
		conn, err := ln.Accept()
		if err != nil {
			Error(fmt.Sprintf("Node: Failed to accept connection: %v", err))
			return
		}
		Debug(fmt.Sprintf("Node: Accepted connection from: %s", conn.RemoteAddr().String()))
		go n.handleConnection(conn)
	}
}

// handleConnection handles Protobuf communication between nodes (independent of chat).
func (n *Node) handleConnection(conn net.Conn) {
	for {
		buf := make([]byte, 4096)
		nRead, err := conn.Read(buf)
		if err != nil {
			Debug(fmt.Sprintf("Node: Error reading from connection: %v", err))
			conn.Close()
			return
		}

		// Log the raw data received
		Debug(fmt.Sprintf("Node: Raw data received (hex): %x", buf[:nRead]))
		Debug(fmt.Sprintf("Node: Raw data received (string): %s", string(buf[:nRead])))

		var mainMessage block_chain.MainMessage
		err = proto.Unmarshal(buf[:nRead], &mainMessage)
		if err != nil {
			Error(fmt.Sprintf("Node: Failed to unmarshal main message: %v", err))
			Error(fmt.Sprintf("Node: Data causing error (hex): %x", buf[:nRead]))
			Error(fmt.Sprintf("Node: Data causing error (string): %s", string(buf[:nRead])))
			conn.Close()
			return
		}

		Debug(fmt.Sprintf("Node: Received message of type %T", mainMessage.MessageType))

		switch msg := mainMessage.MessageType.(type) {
		case *block_chain.MainMessage_BlockMessage:
			Debug("Node: Handling BlockMessage")
			n.blockHandler.HandleBlockMessage(msg.BlockMessage)
		case *block_chain.MainMessage_NodeMessage:
			Debug("Node: Handling NodeMessage")
			n.nodeHandler.HandleNodeMessage(msg.NodeMessage)
		default:
			Error(fmt.Sprintf("Node: Unknown message type: %T", msg))
			conn.Close()
			return
		}
	}
}

// TryToFindNewBlock attempts to find a new block.
func (n *Node) TryToFindNewBlock() {
	Debug("Node: Starting to find a new block...")

	n.mux.Lock() // Lock the mutex before generating the new block

	// Get the block with the highest index
	parentBlock := n.blockchain.GetBlockWithHighestIndex()
	Debug(fmt.Sprintf("Node: Highest block index: %d", parentBlock.Index))

	// Generate a new block with the correct index
	newBlock := n.blockchain.GenerateNewBlock()
	newBlock.Index = parentBlock.Index + 1
	newBlock.PreviousHash = parentBlock.CalculateHash()
	
	// Add one pending message to the new block
	nextMessage := n.GetNextPendingMessage()
	if nextMessage != "" {
		newBlock.Messages = []string{nextMessage}
		Info(fmt.Sprintf("Node: Including message in new block: %s", nextMessage))
	}

	// Validate the new block
	nonce := uint64(0)
	for {
		newBlock.Data = nonce
		if err := n.blockchain.ValidateBlock(newBlock, parentBlock); err == nil {
			break
		}
		nonce++
	}

	// Add the new block to the blockchain
	latestBlockNode := n.blockchain.GetBlock(parentBlock.CalculateHash())
	if latestBlockNode == nil {
		Error("Node: Failed to find the latest block node")
		n.mux.Unlock()
		return
	}

	Debug(fmt.Sprintf("Node: Attempting to add block with index %d", newBlock.Index))
	err := n.blockchain.AddBlock(latestBlockNode, newBlock)
	if err != nil {
		Error(fmt.Sprintf("Node: Failed to add block: %v", err))
		n.mux.Unlock()
		return
	} else {
		Debug(fmt.Sprintf("Node: Block with index %d added successfully", newBlock.Index))
		// Reward the node for finding an approved block
		if newBlock.Checkpoint {
			n.blockchain.RewardNode(n.address, 100)
			Debug(fmt.Sprintf("Node: Reward increased to %d", n.blockchain.GetReward()))
			Debug("Node: Broadcasting latest block to nodes")
			n.blockHandler.BroadcastApprovedBlock(newBlock, n.nodes)
		}
	}

	n.mux.Unlock() // Unlock the mutex after adding the block
}

// SyncNodes synchronizes the node with another node.
func (n *Node) SyncNodes(address int) error {
	Debug(fmt.Sprintf("Node: Synchronizing with node at address: %d from node: %d", address, n.address))
	return n.nodeHandler.SyncNodes(address)
}

// AddPendingMessage adds a message to the pending messages queue.
func (n *Node) AddPendingMessage(message string) {
	n.messageMux.Lock()
	defer n.messageMux.Unlock()
	n.pendingMessages = append(n.pendingMessages, message)
	Info(fmt.Sprintf("Node: Added message to pending queue: %s", message))
}

// GetPendingMessages returns and clears the pending messages queue.
func (n *Node) GetPendingMessages() []string {
	n.messageMux.Lock()
	defer n.messageMux.Unlock()
	messages := make([]string, len(n.pendingMessages))
	copy(messages, n.pendingMessages)
	n.pendingMessages = n.pendingMessages[:0] // Clear the queue
	return messages
}

// GetNextPendingMessage returns and removes the first pending message.
func (n *Node) GetNextPendingMessage() string {
	n.messageMux.Lock()
	defer n.messageMux.Unlock()
	if len(n.pendingMessages) == 0 {
		return ""
	}
	message := n.pendingMessages[0]
	n.pendingMessages = n.pendingMessages[1:] // Remove first message
	return message
}

// GetPendingMessageCount returns the number of pending messages without removing them.
func (n *Node) GetPendingMessageCount() int {
	n.messageMux.Lock()
	defer n.messageMux.Unlock()
	return len(n.pendingMessages)
}
