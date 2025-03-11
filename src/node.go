package src

import (
	"fmt"
	"math/rand"
	"net"
	"sync"

	"github.com/pabloaaa/GO_BLOCKCHAIN/interfaces"
	block_chain "github.com/pabloaaa/GO_BLOCKCHAIN/protos"
	"github.com/pabloaaa/GO_BLOCKCHAIN/types"
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

// GetMessageSender returns the TCP message sender.
func (n *Node) GetMessageSender() *TcpMessageSender {
	return n.tcpMessageSender
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

// handleConnection handles an incoming connection.
func (n *Node) handleConnection(conn net.Conn) {
	for {
		buf := make([]byte, 4096)
		nRead, err := conn.Read(buf)
		if err != nil {
			Debug(fmt.Sprintf("Node: Error reading from connection: %v", err))
			conn.Close()
			return
		}

		var mainMessage block_chain.MainMessage
		err = proto.Unmarshal(buf[:nRead], &mainMessage)
		if err != nil {
			Error(fmt.Sprintf("Node: Failed to unmarshal main message: %v", err))
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

// getRandomNodes returns a random subset of nodes.
func (n *Node) getRandomNodes(count int) []int {
	if count > len(n.nodes) {
		count = len(n.nodes)
	}

	rand.Shuffle(len(n.nodes), func(i, j int) {
		n.nodes[i], n.nodes[j] = n.nodes[j], n.nodes[i]
	})

	return n.nodes[:count]
}

// TryToFindNewBlock attempts to find a new block.
func (n *Node) TryToFindNewBlock() {
	Debug("Node: Starting to find a new block...")

	n.mux.Lock() // Lock the mutex before generating the new block

	// Get the latest approved block or the latest block if no approved block exists
	parentBlock := n.blockchain.GetLatestBlock()
	Debug(fmt.Sprintf("Node: Latest block index: %d", parentBlock.Index))

	// Generate a new block with the correct index
	transaction := []types.Transaction{
		{Sender: []byte("Alice"), Receiver: []byte("Bob"), Amount: 10},
	}
	newBlock := n.blockchain.GenerateNewBlock(transaction)
	newBlock.Index = parentBlock.Index + 1
	newBlock.PreviousHash = parentBlock.CalculateHash()

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
		// Broadcast the new block to other nodes if it has a checkpoint
		if newBlock.Checkpoint {
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

// GetPortMap returns the port map from TcpConnectionManager.
func (n *Node) GetPortMap() map[int]net.Conn {
	return n.tcpConnectionManager.GetPortMap()
}
