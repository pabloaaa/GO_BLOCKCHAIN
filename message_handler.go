// messageHandler.go
package main

import (
	"net"

	block_chain "github.com/pabloaaa/GO_BLOCKCHAIN/protos"
)

type MessageHandler interface {
	Handle(msg *block_chain.MainMessage, conn net.Conn)
}
