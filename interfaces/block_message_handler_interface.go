package interfaces

import (
	block_chain "github.com/pabloaaa/GO_BLOCKCHAIN/protos"
)

type BlockMessageHandlerInterface interface {
	HandleBlockMessage(msg *block_chain.BlockMessage)
	SetSenderAddress(address string)
}
