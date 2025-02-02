package src

import (
	"log"

	"github.com/pabloaaa/GO_BLOCKCHAIN/interfaces"
	block_chain "github.com/pabloaaa/GO_BLOCKCHAIN/protos"
)

// NodeMessageHandlerImpl handles node-related messages.
type NodeMessageHandlerImpl struct {
	messageSender interfaces.MessageSender
	nodes         *[][]byte
	factory       *MessageFactory
}

// NewNodeMessageHandler creates a new NodeMessageHandlerImpl.
func NewNodeMessageHandler(messageSender interfaces.MessageSender, nodesAddresses *[][]byte) *NodeMessageHandlerImpl {
	return &NodeMessageHandlerImpl{
		messageSender: messageSender,
		nodes:         nodesAddresses,
		factory:       NewMessageFactory(),
	}
}

// HandleNodeMessage processes incoming node messages.
func (h *NodeMessageHandlerImpl) HandleNodeMessage(msg *block_chain.NodeMessage) {
	switch nodeMsg := msg.NodeMessageType.(type) {
	case *block_chain.NodeMessage_WelcomeRequest:
		h.handleWelcomeRequest(nodeMsg.WelcomeRequest.SenderAddress)
	case *block_chain.NodeMessage_WelcomeResponse:
		h.handleWelcomeResponse(nodeMsg.WelcomeResponse.NodeAdresses)
	}
}

// handleWelcomeRequest processes a welcome request message.
func (h *NodeMessageHandlerImpl) handleWelcomeRequest(senderAddress []byte) {
	// Zbuduj wiadomość WelcomeResponse
	welcomeResponse := &block_chain.WelcomeResponse{
		NodeAdresses: *h.nodes,
	}

	// Przygotuj wiadomość do wysłania
	data, err := PrepareProtoMessageToSend(h.factory, welcomeResponse)
	if err != nil {
		log.Fatal(err)
	}

	// Wyślij wiadomość WelcomeResponse do nadawcy
	err = h.messageSender.SendMsgToAddress(senderAddress, data)
	if err != nil {
		log.Printf("Failed to send welcome response to node at address %s: %v", string(senderAddress), err)
	} else {
		log.Printf("Successfully sent welcome response to node at address %s", string(senderAddress))
	}

	// Dodaj adres nadawcy do listy nodów, jeśli nie jest to adres własny
	if !containsAddress(*h.nodes, senderAddress) {
		*h.nodes = append(*h.nodes, senderAddress)
	}
}

// handleWelcomeResponse processes a welcome response message.
func (h *NodeMessageHandlerImpl) handleWelcomeResponse(nodes_addresses [][]byte) {
	for _, addr := range nodes_addresses {
		if !containsAddress(*h.nodes, addr) {
			*h.nodes = append(*h.nodes, addr)
		}
	}
}

// BroadcastAddress sends the node's address to all known nodes.
func (h *NodeMessageHandlerImpl) BroadcastAddress(nodes [][]byte, sender_address []byte) {
	for _, node := range *h.nodes {
		welcomeRequest := &block_chain.WelcomeRequest{
			SenderAddress: sender_address,
		}

		// Przygotuj wiadomość do wysłania
		data, err := PrepareProtoMessageToSend(h.factory, welcomeRequest)
		if err != nil {
			log.Fatal(err)
		}

		// Wyślij wiadomość WelcomeRequest do węzła
		err = h.messageSender.SendMsgToAddress(node, data)
		if err != nil {
			log.Printf("Failed to send message to node at address %s: %v", node, err)
		} else {
			log.Printf("Successfully sent welcome request to node at address %s", node)
		}
	}
}

// containsAddress checks if the given address is in the list of addresses.
func containsAddress(addresses [][]byte, address []byte) bool {
	for _, addr := range addresses {
		if string(addr) == string(address) {
			return true
		}
	}
	return false
}
