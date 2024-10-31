package interfaces

import (
	block_chain "github.com/pabloaaa/GO_BLOCKCHAIN/protos"
)

type NodeMessageHandlerInterface interface {
	HandleNodeMessage(msg *block_chain.NodeMessage)
}
