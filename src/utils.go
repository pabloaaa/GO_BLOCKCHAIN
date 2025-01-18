package src

import (
	"fmt"

	"google.golang.org/protobuf/proto"
)

// EncodeMessage encodes a protobuf message into a byte slice.
func EncodeMessage(message interface{}) ([]byte, error) {
	protoMessage, ok := message.(proto.Message)
	if !ok {
		return nil, fmt.Errorf("failed to cast message to proto.Message")
	}
	data, err := proto.Marshal(protoMessage)
	if err != nil {
		return nil, err
	}
	return data, nil
}
