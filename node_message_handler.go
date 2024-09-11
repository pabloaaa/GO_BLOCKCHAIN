package main

import (
	"bytes"
	"log"
	"net"

	block_chain "github.com/pabloaaa/GO_BLOCKCHAIN/protos"
	"google.golang.org/protobuf/proto"
)

type NodeMessageHandler struct {
	node *Node
}

func NewNodeMessageHandler(node *Node) *NodeMessageHandler {
	return &NodeMessageHandler{node: node}
}

func (h *NodeMessageHandler) Handle(msg *block_chain.MainMessage, conn net.Conn) {
	switch nodeMsg := msg.MessageType.(type) {
	case *block_chain.MainMessage_NodeMessage:
		switch nodeMsg.NodeMessage.NodeMessageType.(type) {
		case *block_chain.NodeMessage_WelcomeRequest:
			h.node.handleWelcomeRequest(nodeMsg.NodeMessage.GetWelcomeRequest().Message, conn.LocalAddr().String())
		case *block_chain.NodeMessage_WelcomeResponse:
			h.node.handleWelcomeResponse(nodeMsg.NodeMessage.GetWelcomeResponse().Message)
			// Add other cases for different NodeMessage types
		}
	}
}

func (n *Node) handleWelcomeRequest(data []byte, address string) {
	welcomeRequest := &block_chain.WelcomeRequest{}
	err := proto.Unmarshal(data, welcomeRequest)
	if err != nil {
		log.Println(err)
		return
	}
	n.nodes = append(n.nodes, welcomeRequest.Message)
	n.SendAddressWelcomeResponse(address)
}

func (n *Node) handleWelcomeResponse(data []byte) {
	welcomeResponse := &block_chain.WelcomeResponse{}
	err := proto.Unmarshal(data, welcomeResponse)
	if err != nil {
		log.Println(err)
		return
	}
	n.AddNodes(welcomeResponse.Message)
}

func (n *Node) BroadcastAddress(address []byte) {
	for _, node := range n.nodes {
		conn, err := net.Dial("tcp", string(node))
		if err != nil {
			log.Printf("Failed to connect to node at address %s: %v", node, err)
			continue
		}
		defer conn.Close()

		welcomeRequest := &block_chain.WelcomeRequest{
			Message: address,
		}

		err = encodeMessage(conn, "WelcomeRequest", welcomeRequest)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func (n *Node) SendAddressWelcomeResponse(address string) {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	nodes := bytes.Join(n.nodes, []byte(", "))

	welcomeResponse := &block_chain.WelcomeResponse{
		Message: nodes,
	}

	err = encodeMessage(conn, "WelcomeResponse", welcomeResponse)
	if err != nil {
		log.Fatal(err)
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
