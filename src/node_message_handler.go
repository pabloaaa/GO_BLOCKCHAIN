package src

import (
	"bytes"
	"log"
	"net"

	block_chain "github.com/pabloaaa/GO_BLOCKCHAIN/protos"
	"google.golang.org/protobuf/proto"
)

type NodeMessageHandlerImpl struct {
	node *Node
}

func NewNodeMessageHandler(node *Node) *NodeMessageHandlerImpl {
	return &NodeMessageHandlerImpl{node: node}
}

func (h *NodeMessageHandlerImpl) HandleNodeMessage(msg *block_chain.NodeMessage, conn net.Conn) {
	switch nodeMsg := msg.NodeMessageType.(type) {
	case *block_chain.NodeMessage_WelcomeRequest:
		h.node.handleWelcomeRequest(nodeMsg.WelcomeRequest.Message, conn.LocalAddr().String())
	case *block_chain.NodeMessage_WelcomeResponse:
		h.node.handleWelcomeResponse(nodeMsg.WelcomeResponse.Message)
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

		nodeMessage := &block_chain.NodeMessage{
			NodeMessageType: &block_chain.NodeMessage_WelcomeRequest{
				WelcomeRequest: welcomeRequest,
			},
		}

		err = EncodeMessage(conn, nodeMessage)
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

	nodeMessage := &block_chain.NodeMessage{
		NodeMessageType: &block_chain.NodeMessage_WelcomeResponse{
			WelcomeResponse: welcomeResponse,
		},
	}

	err = EncodeMessage(conn, nodeMessage)
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
