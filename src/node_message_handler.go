package src

import (
	"log"
	"strconv"

	"github.com/pabloaaa/GO_BLOCKCHAIN/interfaces"
	block_chain "github.com/pabloaaa/GO_BLOCKCHAIN/protos"
)

// NodeMessageHandlerImpl handles node-related messages.
type NodeMessageHandlerImpl struct {
	messageSender interfaces.MessageSender
	nodes         *[]int
	factory       *MessageFactory
	localPort     int
}

// NewNodeMessageHandler creates a new NodeMessageHandlerImpl.
func NewNodeMessageHandler(messageSender interfaces.MessageSender, nodesAddresses *[]int, localPort int) *NodeMessageHandlerImpl {
	return &NodeMessageHandlerImpl{
		messageSender: messageSender,
		nodes:         nodesAddresses,
		factory:       NewMessageFactory(),
		localPort:     localPort,
	}
}

// HandleNodeMessage processes incoming node messages.
func (h *NodeMessageHandlerImpl) HandleNodeMessage(msg *block_chain.NodeMessage) {
	switch nodeMsg := msg.NodeMessageType.(type) {
	case *block_chain.NodeMessage_WelcomeRequest:
		log.Printf("\033[32mNodeMessageHandlerImpl: Handling WelcomeRequest from %s\033[0m", nodeMsg.WelcomeRequest.SenderAddress)
		senderPort, _ := strconv.Atoi(string(nodeMsg.WelcomeRequest.SenderAddress))
		h.handleWelcomeRequest(senderPort)
	case *block_chain.NodeMessage_WelcomeResponse:
		log.Printf("\033[32mNodeMessageHandlerImpl: Handling WelcomeResponse\033[0m")
		h.handleWelcomeResponse(nodeMsg.WelcomeResponse.NodeAdresses)
	}
}

// handleWelcomeRequest processes a welcome request message.
func (h *NodeMessageHandlerImpl) handleWelcomeRequest(senderAddress int) {
	// Build WelcomeResponse message
	nodeAddresses := make([][]byte, len(*h.nodes))
	for i, node := range *h.nodes {
		nodeAddresses[i] = []byte(strconv.Itoa(node))
	}
	welcomeResponse := &block_chain.WelcomeResponse{
		NodeAdresses: nodeAddresses,
	}

	// Prepare message to send
	data, err := PrepareProtoMessageToSend(h.factory, welcomeResponse)
	if err != nil {
		log.Fatal(err)
	}

	// Send WelcomeResponse message to sender
	err = h.messageSender.SendMsgToAddress(senderAddress, data, h.localPort)
	if err != nil {
		log.Printf("\033[32mNodeMessageHandlerImpl: Failed to send welcome response to node at address %d: %v\033[0m", senderAddress, err)
	} else {
		log.Printf("\033[32mNodeMessageHandlerImpl: Successfully sent welcome response to node at address %d\033[0m", senderAddress)
	}

	// Add sender address to nodes list if it is not the local address
	if !containsAddress(*h.nodes, senderAddress) {
		*h.nodes = append(*h.nodes, senderAddress)
		log.Printf("\033[32mNodeMessageHandlerImpl: Added sender address %d to nodes list\033[0m", senderAddress)
	}
}

// handleWelcomeResponse processes a welcome response message.
func (h *NodeMessageHandlerImpl) handleWelcomeResponse(nodes_addresses [][]byte) {
	for _, addr := range nodes_addresses {
		nodeAddress, _ := strconv.Atoi(string(addr))
		if !containsAddress(*h.nodes, nodeAddress) {
			*h.nodes = append(*h.nodes, nodeAddress)
			log.Printf("\033[32mNodeMessageHandlerImpl: Added node address %d to nodes list\033[0m", nodeAddress)
		}
	}
}

// BroadcastAddress sends the node's address to all known nodes.
func (h *NodeMessageHandlerImpl) BroadcastAddress(nodes []int, sender_address int) {
	for _, node := range nodes {
		welcomeRequest := &block_chain.WelcomeRequest{
			SenderAddress: []byte(strconv.Itoa(sender_address)),
		}

		// Prepare message to send
		data, err := PrepareProtoMessageToSend(h.factory, welcomeRequest)
		if err != nil {
			log.Fatal(err)
		}

		// Send WelcomeRequest message to node
		err = h.messageSender.SendMsgToAddress(node, data, h.localPort)
		if err != nil {
			log.Printf("\033[32mNodeMessageHandlerImpl: Failed to send message to node at address %d: %v\033[0m", node, err)
		} else {
			log.Printf("\033[32mNodeMessageHandlerImpl: Successfully sent welcome request to node at address %d\033[0m", node)
		}
	}
}

// containsAddress checks if the given address is in the list of addresses.
func containsAddress(addresses []int, address int) bool {
	for _, addr := range addresses {
		if addr == address {
			return true
		}
	}
	return false
}
