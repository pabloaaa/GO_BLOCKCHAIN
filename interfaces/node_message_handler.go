package interfaces

import (
	"net"

	block_chain "github.com/pabloaaa/GO_BLOCKCHAIN/protos"
)

type NodeMessageHandler interface {
	HandleNodeMessage(msg *block_chain.NodeMessage, conn net.Conn)
}
