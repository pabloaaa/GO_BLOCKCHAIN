package src

import (
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

// PrepareProtoMessageToSend prepares a proto message to be sent.
func PrepareProtoMessageToSend(factory *MessageFactory, message proto.Message) ([]byte, error) {
	var mainMessage *block_chain.MainMessage
	var err error

	switch message.(type) {
	case *block_chain.WelcomeRequest, *block_chain.WelcomeResponse:
		mainMessage, err = factory.CreateNodeMessage(message)
	case *block_chain.BlocksResponse, *block_chain.BlockchainSyncRequest:
		mainMessage, err = factory.CreateBlockMessage(message)
	}

	if err != nil {
		return nil, err
	}

	return EncodeMessage(mainMessage)
}
