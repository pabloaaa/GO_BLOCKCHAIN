package src

import (
	"bytes"
	"log"

	block_chain "github.com/pabloaaa/GO_BLOCKCHAIN/protos"
	"google.golang.org/protobuf/proto"
)

type NodeMessageHandlerImpl struct {
	node *Node
}

func NewNodeMessageHandler(node *Node) *NodeMessageHandlerImpl {
	return &NodeMessageHandlerImpl{node: node}
}

func (h *NodeMessageHandlerImpl) HandleNodeMessage(msg *block_chain.NodeMessage) {
	switch nodeMsg := msg.NodeMessageType.(type) {
	case *block_chain.NodeMessage_WelcomeRequest:
		h.node.handleWelcomeRequest(nodeMsg.WelcomeRequest.Message)
	case *block_chain.NodeMessage_WelcomeResponse:
		h.node.handleWelcomeResponse(nodeMsg.WelcomeResponse.Message)
	}
}

func (n *Node) handleWelcomeRequest(data []byte) {
	welcomeRequest := &block_chain.WelcomeRequest{}
	err := proto.Unmarshal(data, welcomeRequest)
	if err != nil {
		log.Println(err)
		return
	}
	n.nodes = append(n.nodes, welcomeRequest.Message)
	n.SendAddressWelcomeResponse()
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
		welcomeRequest := &block_chain.WelcomeRequest{
			Message: address,
		}

		nodeMessage := &block_chain.NodeMessage{
			NodeMessageType: &block_chain.NodeMessage_WelcomeRequest{
				WelcomeRequest: welcomeRequest,
			},
		}

		data, err := EncodeMessage(nodeMessage)
		if err != nil {
			log.Fatal(err)
		}

		err = n.GetMessageSender().SendMsg(data)
		if err != nil {
			log.Printf("Failed to send message to node at address %s: %v", node, err)
		}
	}
}

func (n *Node) SendAddressWelcomeResponse() {
	nodes := bytes.Join(n.nodes, []byte(", "))

	welcomeResponse := &block_chain.WelcomeResponse{
		Message: nodes,
	}

	nodeMessage := &block_chain.NodeMessage{
		NodeMessageType: &block_chain.NodeMessage_WelcomeResponse{
			WelcomeResponse: welcomeResponse,
		},
	}

	data, err := EncodeMessage(nodeMessage)
	if err != nil {
		log.Fatal(err)
	}

	err = n.GetMessageSender().SendMsg(data)
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
