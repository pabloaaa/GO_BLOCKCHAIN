package src

import (
	"log"
	"math/rand"
	"net"
	"time"

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
	tcpMessageSender interfaces.MessageSender
	address          string
}

func NewNode(blockchain interfaces.BlockchainInterface, address string) *Node {
	messageSender, _ := NewTCPSender(address)
	node := &Node{
		blockchain:       blockchain,
		nodes:            make([][]byte, 0),
		tcpMessageSender: messageSender,
		address:          address,
	}
	node.blockHandler = NewBlockMessageHandler(blockchain, node.tcpMessageSender)
	node.nodeHandler = NewNodeMessageHandler(node)
	return node
}

func (n *Node) GetBlockchain() interfaces.BlockchainInterface {
	return n.blockchain
}

func (n *Node) GetNodes() [][]byte {
	return n.nodes
}

func (n *Node) GetAddress() string {
	return n.address
}

func (n *Node) GetMessageSender() interfaces.MessageSender {
	return n.tcpMessageSender
}

func (n *Node) Start() {
	ln, err := net.Listen("tcp", n.address)
	if err != nil {
		log.Fatalf("Failed to listen on address %s: %v", n.address, err)
	}
	defer ln.Close()

	n.BroadcastAddress([]byte(n.address))
	go n.TryToFindNewBlock()

	go n.blockHandler.BroadcastLatestBlock(n.nodes) // Implement this method

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

	var blockMessage block_chain.BlockMessage
	err = proto.Unmarshal(buf[:nRead], &blockMessage)
	if err == nil {
		n.blockHandler.HandleBlockMessage(&blockMessage)
		return
	}

	var nodeMessage block_chain.NodeMessage
	err = proto.Unmarshal(buf[:nRead], &nodeMessage)
	if err == nil {
		n.nodeHandler.HandleNodeMessage(&nodeMessage)
		return
	}

	log.Println("Failed to unmarshal message")
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
	for {
		transaction := []types.Transaction{
			{Sender: []byte("Alice"), Receiver: []byte("Bob"), Amount: 10},
		}
		newBlock := n.blockchain.GenerateNewBlock(transaction)
		nonce := uint64(0)

		for {
			newBlock.Data = nonce
			parentBlock := n.blockchain.GetLatestBlock()
			if err := n.blockchain.ValidateBlock(newBlock, parentBlock); err == nil {
				break
			}
			nonce++
		}

		n.blockchain.AddBlock(n.blockchain.GetRoot(), newBlock)
		time.Sleep(10 * time.Second)
	}
}
