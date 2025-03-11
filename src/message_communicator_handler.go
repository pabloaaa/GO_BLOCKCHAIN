package src

import (
	"fmt"

	"github.com/pabloaaa/GO_BLOCKCHAIN/interfaces"
	block_chain "github.com/pabloaaa/GO_BLOCKCHAIN/protos"
)

// MessageCommunicatorHandler handles message communication.
type MessageCommunicatorHandler struct {
	blockchain       interfaces.BlockchainInterface
	tcpMessageSender *TcpMessageSender
	messageFactory   *MessageFactory
	address          int
}

// NewMessageCommunicatorHandler creates a new MessageCommunicatorHandler.
func NewMessageCommunicatorHandler(blockchain interfaces.BlockchainInterface, tcpMessageSender *TcpMessageSender, address int) *MessageCommunicatorHandler {
	return &MessageCommunicatorHandler{
		blockchain:       blockchain,
		tcpMessageSender: tcpMessageSender,
		address:          address,
		messageFactory:   NewMessageFactory(),
	}
}

// SendMessage sends a message to another node if the node has enough reward.
func (h *MessageCommunicatorHandler) SendMessage(receiver int, content string) error {
	if h.blockchain.GetReward() < 10 {
		return fmt.Errorf("not enough reward to send message")
	}

	message := &block_chain.Message{
		Sender:   []byte(fmt.Sprintf("Node%d", h.address)),
		Receiver: []byte(fmt.Sprintf("Node%d", receiver)),
		Content:  content,
	}

	// Create and encode the message
	data, err := PrepareProtoMessageToSend(h.messageFactory, message)
	if err != nil {
		return err
	}

	err = h.tcpMessageSender.SendMsgToAddress(receiver, data)
	if err != nil {
		return err
	}

	h.blockchain.RewardNode(h.address, ^uint64(9)) // Deduct 10 by adding -10 (two's complement)
	Debug(fmt.Sprintf("MessageCommunicatorHandler: Sent message to node %d, reward decreased to %d", receiver, h.blockchain.GetReward()))
	return nil
}

// HandleCommunicatorMessage handles incoming communicator messages.
func (h *MessageCommunicatorHandler) HandleCommunicatorMessage(message *block_chain.Message) {
	// Handle the received message
	Debug(fmt.Sprintf("MessageCommunicatorHandler: Received message from %s to %s: %s", message.Sender, message.Receiver, message.Content))
}
