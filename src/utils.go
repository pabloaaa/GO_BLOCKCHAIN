package src

import (
	"fmt"

	block_chain "github.com/pabloaaa/GO_BLOCKCHAIN/protos"
	"google.golang.org/protobuf/proto"
)

// EncodeMessage encodes a protobuf message into a byte slice.
func EncodeMessage(message proto.Message) ([]byte, error) {
	data, err := proto.Marshal(message)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// WrapMessage wraps a message in the appropriate protobuf message type.
func WrapMessage(msg proto.Message) (*block_chain.MainMessage, error) {
	switch m := msg.(type) {
	case *block_chain.BlockMessage:
		return &block_chain.MainMessage{
			MessageType: &block_chain.MainMessage_BlockMessage{
				BlockMessage: m,
			},
		}, nil
	case *block_chain.NodeMessage:
		return &block_chain.MainMessage{
			MessageType: &block_chain.MainMessage_NodeMessage{
				NodeMessage: m,
			},
		}, nil
	default:
		return nil, fmt.Errorf("unknown message type: %T", msg)
	}
}
