package main

import (
	"bytes"
	"encoding/gob"
	"log"
	"math/rand"
	"net"
	"time"

	pb "github.com/pabloaaa/GO_BLOCKCHAIN/protos"
	"google.golang.org/protobuf/proto"
)

type Message struct {
	Type []byte
	Data []byte
}

type Node struct {
	blockchain *Blockchain
	nodes      [][]byte
}

func NewNode(blockchain *Blockchain) *Node {
	return &Node{
		blockchain: blockchain,
		nodes:      make([][]byte, 0),
	}
}

func (n *Node) Start(address []byte) {
	ln, err := net.Listen("tcp", string(address))
	if err != nil {
		log.Fatalf("Failed to listen on address %s: %v", address, err)
	}
	defer ln.Close()

	n.BroadcastAddress(address)
	go n.TryToFindNewBlock()

	go n.BroadcastLatestBlock()

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

	decoder := gob.NewDecoder(conn)

	var message Message
	err := decoder.Decode(&message)
	if err != nil {
		log.Println(err)
		return
	}

	// Handle the message based on its type
	switch string(message.Type) {
	case "WelcomeRequest":
		n.handleWelcomeRequest(message.Data, conn.LocalAddr().String())
	case "WelcomeResponse":
		n.handleWelcomeResponse(message.Data)
	case "GetLatestBlock":
		n.handleGetLatestBlock(message.Data, conn.LocalAddr().String())
	case "BlockResponse":
		n.handleBlockResponse(message.Data, conn.LocalAddr().String())
	case "GetBlockRequest":
		n.handleGetBlockRequest(message.Data, conn.LocalAddr().String())
	}
}

func (n *Node) handleWelcomeRequest(data []byte, address string) {
	welcomeRequest := &pb.WelcomeRequest{}
	err := proto.Unmarshal(data, welcomeRequest)
	if err != nil {
		log.Println(err)
		return
	}
	n.nodes = append(n.nodes, welcomeRequest.Message)
	n.SendAddressWelcomeResponse(address)
}

func (n *Node) handleWelcomeResponse(data []byte) {
	welcomeResponse := &pb.WelcomeResponse{}
	err := proto.Unmarshal(data, welcomeResponse)
	if err != nil {
		log.Println(err)
		return
	}
	n.AddNodes(welcomeResponse.Message)
}

func (n *Node) handleGetLatestBlock(data []byte, address string) {
	getLatestBlockRequest := &pb.GetLatestBlockRequest{}
	err := proto.Unmarshal(data, getLatestBlockRequest)
	if err != nil {
		log.Println(err)
		return
	}
	n.SendLatestBlock(address)
}

func (n *Node) handleBlockResponse(data []byte, address string) {
	blockResponse := &pb.BlockResponse{}
	err := proto.Unmarshal(data, blockResponse)
	if err != nil {
		log.Println(err)
		return
	}
	block := BlockFromProto(blockResponse.GetBlock())
	blockHash := block.calculateHash()
	if !n.blockchain.BlockExists(blockHash) {
		n.GetBlock(address, block.PreviousHash)
	} else {
		parent := n.blockchain.GetBlock(block.PreviousHash)
		if parent != nil {
			// Validate the block before adding it to the blockchain
			err := n.blockchain.ValidateBlock(block, parent.Block)
			if err != nil {
				log.Println("Received invalid block: ", err)
			} else {
				err := n.blockchain.AddBlock(parent, block)
				if err != nil {
					log.Println(err)
				}
			}
		}
	}
}

func (n *Node) handleGetBlockRequest(data []byte, address string) {
	getBlockRequest := &pb.GetBlockRequest{}
	err := proto.Unmarshal(data, getBlockRequest)
	if err != nil {
		log.Println(err)
		return
	}
	block := n.blockchain.GetBlock(getBlockRequest.Hash)
	if block != nil {
		n.SendBlock(address, block)
	}
}

func (n *Node) BroadcastAddress(address []byte) {
	for _, node := range n.nodes {
		conn, err := net.Dial("tcp", string(node))
		if err != nil {
			log.Printf("Nie udało się połączyć z węzłem o adresie %s: %v", node, err)
			continue
		}
		defer conn.Close()

		welcomeRequest := &pb.WelcomeRequest{
			Message: address,
		}

		err = encodeMessage(conn, "WelcomeRequest", welcomeRequest)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func encodeMessage(conn net.Conn, messageType string, messageData proto.Message) error {
	data, err := proto.Marshal(messageData)
	if err != nil {
		return err
	}

	message := &Message{
		Type: []byte(messageType),
		Data: data,
	}

	encoder := gob.NewEncoder(conn)
	err = encoder.Encode(message)
	if err != nil {
		return err
	}

	return nil
}

func (n *Node) SendAddressWelcomeResponse(address string) {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	// Convert the list of known nodes to a single string
	nodes := bytes.Join(n.nodes, []byte(", "))

	welcomeResponse := &pb.WelcomeResponse{
		Message: nodes,
	}

	err = encodeMessage(conn, "WelcomeResponse", welcomeResponse)
	if err != nil {
		log.Fatal(err)
	}
}

func (n *Node) GetLatestBlock(address string) {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		log.Printf("Failed to dial node at address %s: %v", address, err)
		return
	}
	defer conn.Close()

	emptyMessage := &pb.Empty{}
	err = encodeMessage(conn, "GetLatestBlock", emptyMessage)
	if err != nil {
		log.Printf("Failed to encode message: %v", err)
	}
}

func (n *Node) AddNodes(address []byte) {
	for _, node := range n.nodes {
		if bytes.Equal(node, address) {
			return
		}
	}
	n.nodes = append(n.nodes, address)
}

func (n *Node) BroadcastLatestBlock() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		nodes := n.getRandomNodes(3)
		for _, node := range nodes {
			n.GetLatestBlock(string(node))
		}
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

func (n *Node) SendLatestBlock(address string) {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	latestBlock := n.blockchain.GetLatestBlock()

	// Przekształć latestBlock do formatu pb.Block
	protoBlock := latestBlock.ToProto()

	blockResponse := &pb.BlockResponse{
		Success: true,
		Message: []byte("Latest block"),
		Block:   protoBlock,
	}

	err = encodeMessage(conn, "BlockResponse", blockResponse)
	if err != nil {
		log.Fatal(err)
	}
}

func (n *Node) GetBlock(address string, blockHash []byte) {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		log.Printf("Failed to dial node at address %s: %v", address, err)
		return
	}
	defer conn.Close()

	getBlockRequest := &pb.GetBlockRequest{
		Hash: blockHash,
	}

	err = encodeMessage(conn, "GetBlock", getBlockRequest)
	if err != nil {
		log.Printf("Failed to encode message: %v", err)
	}
}

func (n *Node) SendBlock(address string, blockNode *BlockNode) {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	// Convert block to pb.Block format
	protoBlock := blockNode.Block.ToProto()

	blockResponse := &pb.BlockResponse{
		Success: true,
		Message: []byte("Block"),
		Block:   protoBlock,
	}

	err = encodeMessage(conn, "BlockResponse", blockResponse)
	if err != nil {
		log.Fatal(err)
	}
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
