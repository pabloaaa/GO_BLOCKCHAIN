package interfaces

import block_chain "github.com/pabloaaa/GO_BLOCKCHAIN/protos"

// MessageCommunicatorHandlerInterface defines the methods for handling message communication.
type MessageCommunicatorHandlerInterface interface {
	SendMessage(receiver int, content string) error
	HandleCommunicatorMessage(message *block_chain.Message)
}
