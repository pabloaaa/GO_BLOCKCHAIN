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
	log.Printf("\033[32mNodeMessageHandlerImpl[%d]: Received node message of type %T\033[0m", h.localPort, msg.NodeMessageType)
	switch nodeMsg := msg.NodeMessageType.(type) {
	case *block_chain.NodeMessage_WelcomeRequest:
		log.Printf("\033[32mNodeMessageHandlerImpl[%d]: Handling WelcomeRequest from %s\033[0m", h.localPort, nodeMsg.WelcomeRequest.SenderAddress)
		senderPort, _ := strconv.Atoi(string(nodeMsg.WelcomeRequest.SenderAddress))
		h.handleWelcomeRequest(senderPort)
	case *block_chain.NodeMessage_WelcomeResponse:
		log.Printf("\033[32mNodeMessageHandlerImpl[%d]: Handling WelcomeResponse\033[0m", h.localPort)
		h.handleWelcomeResponse(nodeMsg.WelcomeResponse.NodeAdresses)
	default:
		log.Printf("\033[32mNodeMessageHandlerImpl[%d]: Unknown node message type: %T\033[0m", h.localPort, nodeMsg)
	}
}

// handleWelcomeRequest processes a welcome request message.
func (h *NodeMessageHandlerImpl) handleWelcomeRequest(senderAddress int) {
	log.Printf("NodeMessageHandlerImpl[%d]: handleWelcomeRequest called with senderAddress: %d", h.localPort, senderAddress)
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
		log.Printf("NodeMessageHandlerImpl[%d]: Error preparing proto message: %v", h.localPort, err)
		log.Fatal(err)
	}

	// Send WelcomeResponse message to sender
	err = h.messageSender.SendMsgToAddress(senderAddress, data)
	if err != nil {
		log.Printf("NodeMessageHandlerImpl[%d]: Failed to send welcome response to node at address %d: %v", h.localPort, senderAddress, err)
	} else {
		log.Printf("NodeMessageHandlerImpl[%d]: Successfully sent welcome response to node at address %d", h.localPort, senderAddress)
	}

	// Add sender address to nodes list if it is not the local address
	if !containsAddress(*h.nodes, senderAddress) {
		*h.nodes = append(*h.nodes, senderAddress)
		log.Printf("NodeMessageHandlerImpl[%d]: Added sender address %d to nodes list", h.localPort, senderAddress)
	}
}

// handleWelcomeResponse processes a welcome response message.
func (h *NodeMessageHandlerImpl) handleWelcomeResponse(nodes_addresses [][]byte) {
	log.Printf("NodeMessageHandlerImpl[%d]: handleWelcomeResponse called with nodes_addresses: %v", h.localPort, nodes_addresses)
	for _, addr := range nodes_addresses {
		nodeAddress, err := strconv.Atoi(string(addr))
		if err != nil {
			log.Printf("NodeMessageHandlerImpl[%d]: Error converting address %s to int: %v", h.localPort, addr, err)
			continue
		}
		if !containsAddress(*h.nodes, nodeAddress) {
			*h.nodes = append(*h.nodes, nodeAddress)
			log.Printf("NodeMessageHandlerImpl[%d]: Added node address %d to nodes list", h.localPort, nodeAddress)
		}
	}
}

// BroadcastAddress sends the node's address to all known nodes.
func (h *NodeMessageHandlerImpl) BroadcastAddress(nodes []int, sender_address int) {
	for _, node := range nodes {
		log.Printf("\033[32mNodeMessageHandlerImpl[%d]: Broadcasting address %d to node at address %d\033[0m", h.localPort, sender_address, node)
		welcomeRequest := &block_chain.WelcomeRequest{
			SenderAddress: []byte(strconv.Itoa(sender_address)),
		}

		// Prepare message to send
		data, err := PrepareProtoMessageToSend(h.factory, welcomeRequest)
		if err != nil {
			log.Fatal(err)
		}

		// Send WelcomeRequest message to node
		err = h.messageSender.SendMsgToAddress(node, data)
		if err != nil {
			log.Printf("\033[32mNodeMessageHandlerImpl[%d]: Failed to send message to node at address %d: %v\033[0m", h.localPort, node, err)
		} else {
			log.Printf("\033[32mNodeMessageHandlerImpl[%d]: Successfully sent welcome request to node at address %d\033[0m", h.localPort, node)
		}
	}
}

// SyncNodes synchronizes the node with another node.
func (h *NodeMessageHandlerImpl) SyncNodes(address int) error {
	log.Printf("NodeMessageHandlerImpl[%d]: Synchronizing with node at address: %d from node: %d", h.localPort, address, h.localPort)

	blockchainSyncRequest := &block_chain.BlockchainSyncRequest{
		SenderAddress: []byte(strconv.Itoa(h.localPort)),
	}

	// Prepare message to send
	data, err := PrepareProtoMessageToSend(h.factory, blockchainSyncRequest)
	if err != nil {
		return err
	}

	// Send BlockchainSyncRequest message to sender
	log.Printf("NodeMessageHandlerImpl[%d]: Sending BlockchainSyncRequest to node at address %d", h.localPort, address)
	err = h.messageSender.SendMsgToAddress(address, data)
	if err != nil {
		log.Printf("NodeMessageHandlerImpl[%d]: Failed to send BlockchainSyncRequest to node at address %d: %v", h.localPort, address, err)
		return err
	}
	log.Printf("NodeMessageHandlerImpl[%d]: Successfully sent BlockchainSyncRequest to node at address %d", h.localPort, address)
	return nil
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
