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

type Message struct {
	Type []byte
	Data []byte
}

type Node struct {
	blockchain       interfaces.BlockchainInterface
	nodes            [][]byte
	blockHandler     interfaces.BlockMessageHandlerInterface
	nodeHandler      interfaces.NodeMessageHandlerInterface
	tcpMessageSender *TcpMessageSender
	address          []byte
	mux              sync.Mutex
}

func NewNode(blockchain interfaces.BlockchainInterface, address []byte, tcpMessageSender *TcpMessageSender, bootstrapAddress []byte) *Node {
	log.Printf("Initializing node with address: %s", address)

	node := &Node{
		blockchain:       blockchain,
		nodes:            make([][]byte, 0),
		tcpMessageSender: tcpMessageSender,
		address:          address,
	}
	node.blockHandler = NewBlockMessageHandler(blockchain, node.tcpMessageSender)
	node.nodeHandler = NewNodeMessageHandler(node.tcpMessageSender, &node.nodes)

	// If bootstrapAddress is provided and not the same as node address, add it to the list of nodes
	if len(bootstrapAddress) > 0 && string(bootstrapAddress) != string(address) {
		node.nodes = append(node.nodes, bootstrapAddress)
	}

	return node
}

func (n *Node) GetBlockchain() interfaces.BlockchainInterface {
	return n.blockchain
}

func (n *Node) GetNodes() [][]byte {
	return n.nodes
}

func (n *Node) GetAddress() []byte {
	return n.address
}

func (n *Node) GetMessageSender() *TcpMessageSender {
	return n.tcpMessageSender
}

func (n *Node) Start() {
	log.Printf("Node starting on address: %s", n.address)

	ln, err := net.Listen("tcp", string(n.address))
	if err != nil {
		log.Fatalf("Failed to listen on address %s: %v", n.address, err)
	}
	defer ln.Close()

	// Sprawdź, czy node.address nie jest równy bootstrapAddress
	if len(n.nodes) > 0 && string(n.nodes[0]) != string(n.address) {
		n.nodeHandler.BroadcastAddress(n.nodes, n.address)
	}

	// go n.blockHandler.BroadcastLatestBlock(n.nodes) // trzeba doimplementowac

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Fatalf("Failed to accept connection: %v", err)
		}
		go n.handleConnection(conn)
	}
}

func (n *Node) handleConnection(conn net.Conn) {
	defer conn.Close()

	buf := make([]byte, 4096)
	nRead, err := conn.Read(buf)
	if err != nil {
		log.Println(err)
		return
	}

	var mainMessage block_chain.MainMessage
	err = proto.Unmarshal(buf[:nRead], &mainMessage)
	if err != nil {
		log.Printf("Failed to unmarshal main message: %v", err)
		return
	}

	switch msg := mainMessage.MessageType.(type) {
	case *block_chain.MainMessage_BlockMessage:
		log.Printf("Received block message")
		n.blockHandler.HandleBlockMessage(msg.BlockMessage)
	case *block_chain.MainMessage_NodeMessage:
		log.Printf("Received node message")
		n.nodeHandler.HandleNodeMessage(msg.NodeMessage)
	default:
		log.Printf("Unknown message type: %T", msg)
	}
}

func (n *Node) getRandomNodes(count int) [][]byte {
	if count > len(n.nodes) {
		count = len(n.nodes)
	}

	rand.Shuffle(len(n.nodes), func(i, j int) {
		n.nodes[i], n.nodes[j] = n.nodes[j], n.nodes[i]
	})

	return n.nodes[:count]
}

func (n *Node) TryToFindNewBlock() {
	log.Println("Starting to find a new block...")

	n.mux.Lock() // Lock the mutex before generating the new block

	// Get the latest approved block or the latest block if no approved block exists
	parentBlock := n.blockchain.GetLatestBlock()
	log.Printf("Latest block index: %d", parentBlock.Index)

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
		log.Printf("Failed to find the latest block node")
		n.mux.Unlock()
		return
	}

	err := n.blockchain.AddBlock(latestBlockNode, newBlock)
	if err != nil {
		log.Printf("Failed to add block: %v", err)
		n.mux.Unlock()
		return
	} else {
		log.Printf("Added new block: %v", newBlock)
		// Broadcast the new block to other nodes if it has a checkpoint
		if newBlock.Checkpoint {
			log.Println("Broadcasting latest block to nodes")
			// n.blockHandler.BroadcastLatestBlock(n.nodes)  trzeba doimplementowac
		}
	}

	n.mux.Unlock() // Unlock the mutex after adding the block
}

func (n *Node) SyncNodes(address []byte) error {
	log.Printf("Synchronizing with node at address: %s from node: %s", address, n.address)
	latestBlockHash := n.blockchain.GetLatestBlock().CalculateHash()
	mainMessage := &block_chain.MainMessage{
		MessageType: &block_chain.MainMessage_BlockMessage{
			BlockMessage: &block_chain.BlockMessage{
				BlockMessageType: &block_chain.BlockMessage_BlockchainSyncRequest{
					BlockchainSyncRequest: &block_chain.BlockchainSyncRequest{
						Hash:          latestBlockHash,
						SenderAddress: n.address,
					},
				},
			},
		},
	}
	log.Printf("Created BlockMessage_BlockchainSyncRequest: %v", mainMessage)

	data, err := EncodeMessage(mainMessage)
	if err != nil {
		return fmt.Errorf("failed to marshal MainMessage: %v", err)
	}

	if len(data) == 0 {
		log.Println("Encoded data is empty")
	} else {
		log.Printf("Encoded MainMessage: %x", data)
	}

	// Send the message to the other node
	log.Printf("Sending MainMessage to address: %s from node: %s with payload: %x", address, n.address, data)
	err = n.tcpMessageSender.SendMsgToAddress(address, data)
	if err != nil {
		return fmt.Errorf("failed to send message: %v", err)
	}

	return nil
}
