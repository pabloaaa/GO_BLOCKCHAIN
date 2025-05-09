package interfaces

import (
	block_chain "github.com/pabloaaa/GO_BLOCKCHAIN/protos"
	"github.com/pabloaaa/GO_BLOCKCHAIN/types"
)

type BlockMessageHandlerInterface interface {
	HandleBlockMessage(msg *block_chain.BlockMessage)
	BroadcastApprovedBlock(block *types.Block, nodes []int)
}
