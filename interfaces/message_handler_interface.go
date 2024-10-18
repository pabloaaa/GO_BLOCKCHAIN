package interfaces

import (
	"net"

	block_chain "github.com/pabloaaa/GO_BLOCKCHAIN/protos"
)

type MessageHandlerInterface interface {
	Handle(msg *block_chain.MainMessage, conn net.Conn)
}
