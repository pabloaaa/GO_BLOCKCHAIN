// messageHandler.go
package main

import (
	"net"

	block_chain "github.com/pabloaaa/GO_BLOCKCHAIN/protos"
)

type MessageHandler interface {
	Handle(msg *block_chain.MainMessage, conn net.Conn)
}

type BlockMessageHandler interface {
	HandleBlockMessage(msg *block_chain.BlockMessage, conn net.Conn)
}

type NodeMessageHandler interface {
	HandleNodeMessage(msg *block_chain.NodeMessage, conn net.Conn)
}
