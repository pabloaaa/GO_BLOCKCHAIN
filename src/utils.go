package src

import (
	"fmt"
	"strconv"
	"strings"

	block_chain "github.com/pabloaaa/GO_BLOCKCHAIN/protos"
	"google.golang.org/protobuf/proto"
)

// CreateSendingPort creates a sending port based on the local port and the bootstrap address.
func CreateSendingPort(localPort, bootstrapPort string) string {
	localPortSuffix := strings.TrimPrefix(localPort, "500")
	bootstrapPortSuffix := strings.TrimPrefix(bootstrapPort, "500")
	return fmt.Sprintf("40%s%s", localPortSuffix, bootstrapPortSuffix)
}

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

// ExtractPort extracts the port number from an address string or a port string.
func ExtractPort(address string) (int, error) {
	if strings.Contains(address, ":") {
		parts := strings.Split(address, ":")
		if len(parts) != 2 {
			return 0, fmt.Errorf("invalid address format")
		}
		return strconv.Atoi(parts[1])
	}
	return strconv.Atoi(address)
}
