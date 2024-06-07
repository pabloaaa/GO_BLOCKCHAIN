package main

import (
	"encoding/gob"
	"log"
	"net"

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
		log.Fatal(err)
	}
	defer ln.Close()

	n.BroadcastAddress(address)

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Fatal(err)
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
	case "Welcome":
		// Handle Welcome message
		welcomeRequest := &pb.WelcomeRequest{}
		err := proto.Unmarshal(message.Data, welcomeRequest)
		if err != nil {
			log.Println(err)
			return
		}
		n.nodes = append(n.nodes, welcomeRequest.Message)
		n.SendAddress(conn.LocalAddr().String())
	case "WelcomeResponse":
		// Handle WelcomeResponse message
		welcomeResponse := &pb.WelcomeResponse{}
		err := proto.Unmarshal(message.Data, welcomeResponse)
		if err != nil {
			log.Println(err)
			return
		}
		n.nodes = append(n.nodes, welcomeResponse.Message)
	}
}

func (n *Node) BroadcastAddress(address []byte) {
	for _, node := range n.nodes {
		conn, err := net.Dial("tcp", string(node))
		if err != nil {
			log.Fatal(err)
		}
		defer conn.Close()

		welcomeRequest := &pb.WelcomeRequest{
			Message: address,
		}
		data, err := proto.Marshal(welcomeRequest)
		if err != nil {
			log.Fatal(err)
		}

		encoder := gob.NewEncoder(conn)
		err = encoder.Encode(&Message{
			Type: []byte("Welcome"),
			Data: data,
		})
		if err != nil {
			log.Fatal(err)
		}
	}
}

func (n *Node) SendAddress(address string) {
	for _, node := range n.nodes {
		conn, err := net.Dial("tcp", string(node))
		if err != nil {
			log.Fatal(err)
		}
		defer conn.Close()

		welcomeResponse := &pb.WelcomeResponse{
			Message: []byte(address),
		}
		data, err := proto.Marshal(welcomeResponse)
		if err != nil {
			log.Fatal(err)
		}

		encoder := gob.NewEncoder(conn)
		err = encoder.Encode(&Message{
			Type: []byte("WelcomeResponse"),
			Data: data,
		})
		if err != nil {
			log.Fatal(err)
		}
	}
}
