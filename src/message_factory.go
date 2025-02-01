package src

import (
	block_chain "github.com/pabloaaa/GO_BLOCKCHAIN/protos"
	"google.golang.org/protobuf/proto"
)

// MessageFactory is responsible for creating and wrapping proto messages.
type MessageFactory struct{}

// NewMessageFactory creates a new MessageFactory.
func NewMessageFactory() *MessageFactory {
	return &MessageFactory{}
}

// CreateNodeMessage creates and wraps a node message.
func (f *MessageFactory) CreateNodeMessage(message proto.Message) (*block_chain.MainMessage, error) {
	var nodeMessage *block_chain.NodeMessage

	switch msg := message.(type) {
	case *block_chain.WelcomeRequest:
		nodeMessage = &block_chain.NodeMessage{
			NodeMessageType: &block_chain.NodeMessage_WelcomeRequest{
				WelcomeRequest: msg,
			},
		}
	case *block_chain.WelcomeResponse:
		nodeMessage = &block_chain.NodeMessage{
			NodeMessageType: &block_chain.NodeMessage_WelcomeResponse{
				WelcomeResponse: msg,
			},
		}
	}

	mainMessage := &block_chain.MainMessage{
		MessageType: &block_chain.MainMessage_NodeMessage{
			NodeMessage: nodeMessage,
		},
	}

	return mainMessage, nil
}

// CreateBlockMessage creates and wraps a block message.
func (f *MessageFactory) CreateBlockMessage(message proto.Message) (*block_chain.MainMessage, error) {
	var blockMessage *block_chain.BlockMessage

	switch msg := message.(type) {
	case *block_chain.BlocksResponse:
		blockMessage = &block_chain.BlockMessage{
			BlockMessageType: &block_chain.BlockMessage_BlocksResponse{
				BlocksResponse: msg,
			},
		}
	case *block_chain.BlockchainSyncRequest:
		blockMessage = &block_chain.BlockMessage{
			BlockMessageType: &block_chain.BlockMessage_BlockchainSyncRequest{
				BlockchainSyncRequest: msg,
			},
		}
	}

	mainMessage := &block_chain.MainMessage{
		MessageType: &block_chain.MainMessage_BlockMessage{
			BlockMessage: blockMessage,
		},
	}

	return mainMessage, nil
}
