package interfaces

import (
	"net"

	block_chain "github.com/pabloaaa/GO_BLOCKCHAIN/protos"
)

type BlockMessageHandler interface {
	HandleBlockMessage(msg *block_chain.BlockMessage, conn net.Conn)
}
