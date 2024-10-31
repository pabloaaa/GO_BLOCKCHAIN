package src

import (
	"google.golang.org/protobuf/proto"
)

func EncodeMessage(message proto.Message) ([]byte, error) {
	data, err := proto.Marshal(message)
	if err != nil {
		return nil, err
	}
	return data, nil
}
