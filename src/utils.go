package src

import (
	"net"

	"google.golang.org/protobuf/proto"
)

func EncodeMessage(conn net.Conn, message proto.Message) error {
	data, err := proto.Marshal(message)
	if err != nil {
		return err
	}

	_, err = conn.Write(data)
	return err
}
