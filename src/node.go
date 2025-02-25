package src

import (
	"fmt"
	"log"
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
	log.Printf("\033[33mNode: Initializing node with address: %d\033[0m", address)

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
		node.tcpConnectionManager.ConnectToNode(bootstrapAddress)
	}

	log.Printf("\033[33mNode: Node initialized with address: %d and bootstrap address: %d\033[0m", address, bootstrapAddress)
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
	log.Printf("\033[33mNode: Node starting on address: %d\033[0m", n.address)

	ln, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", n.address))
	if err != nil {
		log.Fatalf("\033[33mNode: Failed to listen on address %d: %v\033[0m", n.address, err)
	}
	defer ln.Close()

	// Check if node.address is not equal to bootstrapAddress
	if len(n.nodes) > 0 && n.nodes[0] != n.address {
		log.Printf("\033[33mNode: Broadcasting address to nodes: %v\033[0m", n.nodes)
		n.nodeHandler.BroadcastAddress(n.nodes, n.address)
	}

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Fatalf("\033[33mNode: Failed to accept connection: %v\033[0m", err)
		}
		log.Printf("\033[33mNode: Accepted connection from: %s\033[0m", conn.RemoteAddr().String())
		go n.handleConnection(conn)
	}
}

// handleConnection handles an incoming connection.
func (n *Node) handleConnection(conn net.Conn) {
	defer conn.Close()

	buf := make([]byte, 4096)
	nRead, err := conn.Read(buf)
	if err != nil {
		log.Println("\033[33mNode: Error reading from connection: \033[0m", err)
		return
	}

	var mainMessage block_chain.MainMessage
	err = proto.Unmarshal(buf[:nRead], &mainMessage)
	if err != nil {
		log.Printf("\033[33mNode: Failed to unmarshal main message: %v\033[0m", err)
		return
	}

	switch msg := mainMessage.MessageType.(type) {
	case *block_chain.MainMessage_BlockMessage:
		log.Printf("\033[33mNode: Received block message\033[0m")
		n.blockHandler.HandleBlockMessage(msg.BlockMessage)
	case *block_chain.MainMessage_NodeMessage:
		log.Printf("\033[33mNode: Received node message\033[0m")
		n.nodeHandler.HandleNodeMessage(msg.NodeMessage)
	default:
		log.Printf("\033[33mNode: Unknown message type: %T\033[0m", msg)
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
	log.Println("\033[33mNode: Starting to find a new block...\033[0m")

	n.mux.Lock() // Lock the mutex before generating the new block

	// Get the latest approved block or the latest block if no approved block exists
	parentBlock := n.blockchain.GetLatestBlock()
	log.Printf("\033[33mNode: Latest block index: %d\033[0m", parentBlock.Index)

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
		log.Printf("\033[33mNode: Failed to find the latest block node\033[0m")
		n.mux.Unlock()
		return
	}

	log.Printf("\033[33mNode: Attempting to add block with index %d\033[0m", newBlock.Index)
	err := n.blockchain.AddBlock(latestBlockNode, newBlock)
	if err != nil {
		log.Printf("\033[33mNode: Failed to add block: %v\033[0m", err)
		n.mux.Unlock()
		return
	} else {
		log.Printf("\033[33mNode: Block with index %d added successfully\033[0m", newBlock.Index)
		// Broadcast the new block to other nodes if it has a checkpoint
		if newBlock.Checkpoint {
			log.Println("\033[33mNode: Broadcasting latest block to nodes\033[0m")
			n.blockHandler.BroadcastApprovedBlock(newBlock, n.nodes)
		}
	}

	n.mux.Unlock() // Unlock the mutex after adding the block
}

func (n *Node) SyncNodes(address int) error {
	log.Printf("\033[33mNode: Synchronizing with node at address: %d from node: %d\033[0m", address, n.address)
	latestBlockHash := n.blockchain.GetLatestApprovedBlock().CalculateHash()
	mainMessage := &block_chain.MainMessage{
		MessageType: &block_chain.MainMessage_BlockMessage{
			BlockMessage: &block_chain.BlockMessage{
				BlockMessageType: &block_chain.BlockMessage_BlockchainSyncRequest{
					BlockchainSyncRequest: &block_chain.BlockchainSyncRequest{
						Hash:          latestBlockHash,
						SenderAddress: []byte(fmt.Sprintf("%d", n.address)),
					},
				},
			},
		},
	}
	log.Printf("\033[33mNode: Created BlockMessage_BlockchainSyncRequest: %v\033[0m", mainMessage)

	data, err := EncodeMessage(mainMessage)
	if err != nil {
		return fmt.Errorf("\033[33mNode: failed to marshal MainMessage: %v\033[0m", err)
	}

	if len(data) == 0 {
		log.Println("\033[33mNode: Encoded data is empty\033[0m")
	} else {
		log.Printf("\033[33mNode: Encoded MainMessage: %x\033[0m", data)
	}

	// Send the message to the other node
	log.Printf("\033[33mNode: Sending MainMessage to address: %d from node: %d with payload: %x\033[0m", address, n.address, data)
	err = n.tcpMessageSender.SendMsgToAddress(address, data, n.address)
	if err != nil {
		return fmt.Errorf("\033[33mNode: failed to send message: %v\033[0m", err)
	}

	return nil
}

// GetPortMap returns the port map from TcpConnectionManager.
func (n *Node) GetPortMap() map[int]map[int]net.Conn {
	return n.tcpConnectionManager.GetPortMap()
}
