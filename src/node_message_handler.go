package src

import (
	"bytes"
	"log"

	block_chain "github.com/pabloaaa/GO_BLOCKCHAIN/protos"
	"google.golang.org/protobuf/proto"
)

// NodeMessageHandlerImpl handles node-related messages.
type NodeMessageHandlerImpl struct {
	node *Node
}

// NewNodeMessageHandler creates a new NodeMessageHandlerImpl.
func NewNodeMessageHandler(node *Node) *NodeMessageHandlerImpl {
	return &NodeMessageHandlerImpl{node: node}
}

// HandleNodeMessage processes incoming node messages.
func (h *NodeMessageHandlerImpl) HandleNodeMessage(msg *block_chain.NodeMessage) {
	switch nodeMsg := msg.NodeMessageType.(type) {
	case *block_chain.NodeMessage_WelcomeRequest:
		h.node.handleWelcomeRequest(nodeMsg.WelcomeRequest.Message)
	case *block_chain.NodeMessage_WelcomeResponse:
		h.node.handleWelcomeResponse(nodeMsg.WelcomeResponse.Message)
	}
}

// handleWelcomeRequest processes a welcome request message.
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

// handleWelcomeResponse processes a welcome response message.
func (n *Node) handleWelcomeResponse(data []byte) {
	welcomeResponse := &block_chain.WelcomeResponse{}
	err := proto.Unmarshal(data, welcomeResponse)
	if err != nil {
		log.Println(err)
		return
	}
	n.AddNodes(welcomeResponse.Message)
}

// BroadcastAddress sends the node's address to all known nodes.
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

// SendAddressWelcomeResponse sends a welcome response with the node's address.
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

// AddNodes adds a new node address to the list of known nodes.
func (n *Node) AddNodes(address []byte) {
	for _, node := range n.nodes {
		if bytes.Equal(node, address) {
			return
		}
	}
	n.nodes = append(n.nodes, address)
}
