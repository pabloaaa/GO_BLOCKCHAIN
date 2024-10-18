package src

import (
	"log"
	"math/rand"
	"net"
	"time"

	"github.com/pabloaaa/GO_BLOCKCHAIN/interfaces"
	block_chain "github.com/pabloaaa/GO_BLOCKCHAIN/protos"
	"google.golang.org/protobuf/proto"
)

type Message struct {
	Type []byte
	Data []byte
}

type Node struct {
	blockchain   *Blockchain
	nodes        [][]byte
	blockHandler interfaces.BlockMessageHandlerInterface
	nodeHandler  interfaces.NodeMessageHandlerInterface
}

func NewNode(blockchain *Blockchain) *Node {
	node := &Node{
		blockchain: blockchain,
		nodes:      make([][]byte, 0),
	}
	node.blockHandler = NewBlockMessageHandler(blockchain)
	node.nodeHandler = NewNodeMessageHandler(node)
	return node
}

func (n *Node) GetBlockchain() *Blockchain {
	return n.blockchain
}

func (n *Node) GetNodes() [][]byte {
	return n.nodes
}

func (n *Node) Start(address []byte) {
	ln, err := net.Listen("tcp", string(address))
	if err != nil {
		log.Fatalf("Failed to listen on address %s: %v", address, err)
	}
	defer ln.Close()

	n.BroadcastAddress(address)
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
		n.blockHandler.HandleBlockMessage(&blockMessage, conn)
		return
	}

	var nodeMessage block_chain.NodeMessage
	err = proto.Unmarshal(buf[:nRead], &nodeMessage)
	if err == nil {
		n.nodeHandler.HandleNodeMessage(&nodeMessage, conn)
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
		transaction := []Transaction{
			{Sender: []byte("Alice"), Receiver: []byte("Bob"), Amount: 10},
		}
		newBlock := n.blockchain.GenerateNewBlock(transaction)
		nonce := uint64(0)

		for {
			newBlock.SetData(nonce)
			parentBlock := n.blockchain.GetLatestBlock()
			if err := n.blockchain.ValidateBlock(newBlock, parentBlock); err == nil {
				break
			}
			nonce++
		}

		n.blockchain.AddBlock(n.blockchain.root, newBlock)
		time.Sleep(10 * time.Second)
	}
}
