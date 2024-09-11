// utils.go
package main

import (
	"net"

	"google.golang.org/protobuf/proto"
)

func encodeMessage(conn net.Conn, messageType string, message proto.Message) error {
	data, err := proto.Marshal(message)
	if err != nil {
		return err
	}

	_, err = conn.Write(data)
	return err
}
