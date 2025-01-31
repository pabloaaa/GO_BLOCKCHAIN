package src

import (
	"log"

	block_chain "github.com/pabloaaa/GO_BLOCKCHAIN/protos"
)

// NodeMessageHandlerImpl handles node-related messages.
type NodeMessageHandlerImpl struct {
	node          *Node
	senderAddress string
	// nodes         [][]byte
}

// NewNodeMessageHandler creates a new NodeMessageHandlerImpl.
func NewNodeMessageHandler(node *Node) *NodeMessageHandlerImpl {
	return &NodeMessageHandlerImpl{node: node}
}

// SetSenderAddress sets the sender address.
func (h *NodeMessageHandlerImpl) SetSenderAddress(address string) {
	h.senderAddress = address
}

// HandleNodeMessage processes incoming node messages.
func (h *NodeMessageHandlerImpl) HandleNodeMessage(msg *block_chain.NodeMessage) {
	log.Printf("Handling node message of type: %T from node: %s with payload: %v", msg.NodeMessageType, h.senderAddress, msg)
	switch nodeMsg := msg.NodeMessageType.(type) {
	case *block_chain.NodeMessage_WelcomeRequest:
		h.handleWelcomeRequest(nodeMsg.WelcomeRequest.Message)
	case *block_chain.NodeMessage_WelcomeResponse:
		h.handleWelcomeResponse(nodeMsg.WelcomeResponse.Message)
	}
}

// handleWelcomeRequest processes a welcome request message.
func (h *NodeMessageHandlerImpl) handleWelcomeRequest(data []byte) {
	// welcomeRequest := &block_chain.WelcomeRequest{}
	// err := proto.Unmarshal(data, welcomeRequest)
	// if err != nil {
	// 	log.Println(err)
	// 	return
	// }
	// n.nodes = append(n.nodes, welcomeRequest.Message)
	// n.SendAddressWelcomeResponse()
}

// handleWelcomeResponse processes a welcome response message.
func (h *NodeMessageHandlerImpl) handleWelcomeResponse(data []byte) {
	// welcomeResponse := &block_chain.WelcomeResponse{}
	// err := proto.Unmarshal(data, welcomeResponse)
	// if err != nil {
	// 	log.Println(err)
	// 	return
	// }
	// n.AddNodes(welcomeResponse.Message)
}

// BroadcastAddress sends the node's address to all known nodes.
func (h *NodeMessageHandlerImpl) BroadcastAddress(address []byte) {
	// for _, node := range n.nodes {
	// 	welcomeRequest := &block_chain.WelcomeRequest{
	// 		Message: address,
	// 	}

	// 	nodeMessage := &block_chain.NodeMessage{
	// 		NodeMessageType: &block_chain.NodeMessage_WelcomeRequest{
	// 			WelcomeRequest: welcomeRequest,
	// 		},
	// 	}

	// 	data, err := EncodeMessage(nodeMessage)
	// 	if err != nil {
	// 		log.Fatal(err)
	// 	}

	// 	err = n.GetMessageSender().SendMsgToAddress(string(node), data)
	// 	if err != nil {
	// 		log.Printf("Failed to send message to node at address %s: %v", string(node), err)
	// 	} else {
	// 		log.Printf("Successfully sent welcome request to node at address %s", string(node))
	// 	}
	// }
}

// SendAddressWelcomeResponse sends a welcome response with the node's address.
func (n *Node) SendAddressWelcomeResponse() {
	// nodes := bytes.Join(n.nodes, []byte(", "))

	// welcomeResponse := &block_chain.WelcomeResponse{
	// 	Message: nodes,
	// }

	// nodeMessage := &block_chain.NodeMessage{
	// 	NodeMessageType: &block_chain.NodeMessage_WelcomeResponse{
	// 		WelcomeResponse: welcomeResponse,
	// 	},
	// }

	// mainMessage, err := WrapMessage(nodeMessage)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// data, err := EncodeMessage(mainMessage)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// err = n.GetMessageSender().SendMsgToAddress(h.senderAddress, data)
	// if err != nil {
	// 	log.Fatal(err)
	// } else {
	// 	log.Println("Successfully sent welcome response")
	// }
}

// AddNodes adds a new node address to the list of known nodes.
func (n *Node) AddNodes(address []byte) {
	// for _, node := range n.nodes {
	// 	if bytes.Equal(node, address) {
	// 		return
	// 	}
	// }
	// n.nodes = append(n.nodes, address)
}
