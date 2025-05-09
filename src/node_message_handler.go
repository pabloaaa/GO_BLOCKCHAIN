package src

import (
	"fmt"
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
	Debug(fmt.Sprintf("NodeMessageHandlerImpl[%s%d%s]: Received node message of type %T", colorGreen, h.localPort, colorReset, msg.NodeMessageType))
	switch nodeMsg := msg.NodeMessageType.(type) {
	case *block_chain.NodeMessage_WelcomeRequest:
		Debug(fmt.Sprintf("NodeMessageHandlerImpl[%s%d%s]: Handling WelcomeRequest from %s", colorGreen, h.localPort, colorReset, nodeMsg.WelcomeRequest.SenderAddress))
		senderPort, _ := strconv.Atoi(string(nodeMsg.WelcomeRequest.SenderAddress))
		h.handleWelcomeRequest(senderPort)
	case *block_chain.NodeMessage_WelcomeResponse:
		Debug(fmt.Sprintf("NodeMessageHandlerImpl[%s%d%s]: Handling WelcomeResponse", colorGreen, h.localPort, colorReset))
		h.handleWelcomeResponse(nodeMsg.WelcomeResponse.NodeAdresses)
	default:
		Error(fmt.Sprintf("NodeMessageHandlerImpl[%s%d%s]: Unknown node message type: %T", colorGreen, h.localPort, colorReset, nodeMsg))
	}
}

// handleWelcomeRequest processes a welcome request message.
func (h *NodeMessageHandlerImpl) handleWelcomeRequest(senderAddress int) {
	Debug(fmt.Sprintf("NodeMessageHandlerImpl[%s%d%s]: handleWelcomeRequest called with senderAddress: %d", colorGreen, h.localPort, colorReset, senderAddress))
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
		Error(fmt.Sprintf("NodeMessageHandlerImpl[%s%d%s]: Error preparing proto message: %v", colorGreen, h.localPort, colorReset, err))
		return
	}

	// Send WelcomeResponse message to sender
	err = h.messageSender.SendMsgToAddress(senderAddress, data)
	if err != nil {
		Error(fmt.Sprintf("NodeMessageHandlerImpl[%s%d%s]: Failed to send welcome response to node at address %d: %v", colorGreen, h.localPort, colorReset, senderAddress, err))
	} else {
		Info(fmt.Sprintf("NodeMessageHandlerImpl[%s%d%s]: Successfully sent welcome response to node at address %d", colorGreen, h.localPort, colorReset, senderAddress))
	}

	// Add sender address to nodes list if it is not the local address
	if !containsAddress(*h.nodes, senderAddress) {
		*h.nodes = append(*h.nodes, senderAddress)
		Debug(fmt.Sprintf("NodeMessageHandlerImpl[%s%d%s]: Added sender address %d to nodes list", colorGreen, h.localPort, colorReset, senderAddress))
	}
}

// handleWelcomeResponse processes a welcome response message.
func (h *NodeMessageHandlerImpl) handleWelcomeResponse(nodes_addresses [][]byte) {
	Debug(fmt.Sprintf("NodeMessageHandlerImpl[%s%d%s]: handleWelcomeResponse called with nodes_addresses: %v", colorGreen, h.localPort, colorReset, nodes_addresses))
	for _, addr := range nodes_addresses {
		nodeAddress, err := strconv.Atoi(string(addr))
		if err != nil {
			Error(fmt.Sprintf("NodeMessageHandlerImpl[%s%d%s]: Error converting address %s to int: %v", colorGreen, h.localPort, colorReset, addr, err))
			continue
		}
		if !containsAddress(*h.nodes, nodeAddress) {
			*h.nodes = append(*h.nodes, nodeAddress)
			Debug(fmt.Sprintf("NodeMessageHandlerImpl[%s%d%s]: Added node address %d to nodes list", colorGreen, h.localPort, colorReset, nodeAddress))
		}
	}
}

// BroadcastAddress sends the node's address to all known nodes.
func (h *NodeMessageHandlerImpl) BroadcastAddress(nodes []int, sender_address int) {
	for _, node := range nodes {
		Debug(fmt.Sprintf("NodeMessageHandlerImpl[%s%d%s]: Broadcasting address %d to node at address %d", colorGreen, h.localPort, colorReset, sender_address, node))
		welcomeRequest := &block_chain.WelcomeRequest{
			SenderAddress: []byte(strconv.Itoa(sender_address)),
		}

		// Prepare message to send
		data, err := PrepareProtoMessageToSend(h.factory, welcomeRequest)
		if err != nil {
			Error(fmt.Sprintf("NodeMessageHandlerImpl[%s%d%s]: Error preparing proto message: %v", colorGreen, h.localPort, colorReset, err))
			return
		}

		// Send WelcomeRequest message to node
		err = h.messageSender.SendMsgToAddress(node, data)
		if err != nil {
			Error(fmt.Sprintf("NodeMessageHandlerImpl[%s%d%s]: Failed to send message to node at address %d: %v", colorGreen, h.localPort, colorReset, node, err))
		} else {
			Info(fmt.Sprintf("NodeMessageHandlerImpl[%s%d%s]: Successfully sent welcome request to node at address %d", colorGreen, h.localPort, colorReset, node))
		}
	}
}

// SyncNodes synchronizes the node with another node.
func (h *NodeMessageHandlerImpl) SyncNodes(address int) error {
	Debug(fmt.Sprintf("NodeMessageHandlerImpl[%s%d%s]: Synchronizing with node at address: %d from node: %d", colorGreen, h.localPort, colorReset, address, h.localPort))

	blockchainSyncRequest := &block_chain.BlockchainSyncRequest{
		SenderAddress: []byte(strconv.Itoa(h.localPort)),
	}

	// Prepare message to send
	data, err := PrepareProtoMessageToSend(h.factory, blockchainSyncRequest)
	if err != nil {
		Error(fmt.Sprintf("NodeMessageHandlerImpl[%s%d%s]: Error preparing proto message: %v", colorGreen, h.localPort, colorReset, err))
		return err
	}

	// Send BlockchainSyncRequest message to sender
	Debug(fmt.Sprintf("NodeMessageHandlerImpl[%s%d%s]: Sending BlockchainSyncRequest to node at address %d", colorGreen, h.localPort, colorReset, address))
	err = h.messageSender.SendMsgToAddress(address, data)
	if err != nil {
		Error(fmt.Sprintf("NodeMessageHandlerImpl[%s%d%s]: Failed to send BlockchainSyncRequest to node at address %d: %v", colorGreen, h.localPort, colorReset, address, err))
		return err
	}
	Info(fmt.Sprintf("NodeMessageHandlerImpl[%s%d%s]: Successfully sent BlockchainSyncRequest to node at address %d", colorGreen, h.localPort, colorReset, address))
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
