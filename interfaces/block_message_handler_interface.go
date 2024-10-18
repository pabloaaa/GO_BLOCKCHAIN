package interfaces

import (
	"net"

	block_chain "github.com/pabloaaa/GO_BLOCKCHAIN/protos"
)

type BlockMessageHandlerInterface interface {
	HandleBlockMessage(msg *block_chain.BlockMessage, conn net.Conn)
	BroadcastLatestBlock(nodes [][]byte)
}
