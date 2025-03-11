package src

import (
	"fmt"

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
	case *block_chain.BlockchainSyncRequest:
		blockMessage = &block_chain.BlockMessage{
			BlockMessageType: &block_chain.BlockMessage_BlockchainSyncRequest{
				BlockchainSyncRequest: msg,
			},
		}
	case *block_chain.ApprovedBlock:
		blockMessage = &block_chain.BlockMessage{
			BlockMessageType: &block_chain.BlockMessage_ApprovedBlock{
				ApprovedBlock: msg,
			},
		}
	case *block_chain.BlockResponse:
		blockMessage = &block_chain.BlockMessage{
			BlockMessageType: &block_chain.BlockMessage_BlockResponse{
				BlockResponse: msg,
			},
		}
	case *block_chain.BlockRequest:
		blockMessage = &block_chain.BlockMessage{
			BlockMessageType: &block_chain.BlockMessage_BlockRequest{
				BlockRequest: msg,
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

// CreateCustomMessage creates a custom message.
func (f *MessageFactory) CreateCustomMessage(message proto.Message) (*block_chain.MainMessage, error) {
	customMessage, ok := message.(*block_chain.Message)
	if !ok {
		return nil, fmt.Errorf("invalid message type")
	}

	return &block_chain.MainMessage{
		MessageType: &block_chain.MainMessage_CustomMessage{
			CustomMessage: customMessage,
		},
	}, nil
}
