package src

import (
	"log"
	"sort"

	"github.com/pabloaaa/GO_BLOCKCHAIN/interfaces"
	block_chain "github.com/pabloaaa/GO_BLOCKCHAIN/protos"
	"github.com/pabloaaa/GO_BLOCKCHAIN/types"
)

// BlockMessageHandlerImpl handles block-related messages.
type BlockMessageHandlerImpl struct {
	blockchain    interfaces.BlockchainInterface
	messageSender interfaces.MessageSender
	senderAddress []byte
	factory       *MessageFactory
}

// NewBlockMessageHandler creates a new BlockMessageHandlerImpl.
func NewBlockMessageHandler(blockchain interfaces.BlockchainInterface, messageSender interfaces.MessageSender) *BlockMessageHandlerImpl {
	return &BlockMessageHandlerImpl{
		blockchain:    blockchain,
		messageSender: messageSender,
		factory:       NewMessageFactory(),
	}
}

// SetSenderAddress sets the sender address.
func (h *BlockMessageHandlerImpl) SetSenderAddress(address []byte) {
	h.senderAddress = address
}

// HandleBlockMessage processes incoming block messages.
func (h *BlockMessageHandlerImpl) HandleBlockMessage(msg *block_chain.BlockMessage) {
	switch blockMsg := msg.BlockMessageType.(type) {
	case *block_chain.BlockMessage_BlockchainSyncRequest:
		log.Println("Handling BlockchainSyncRequest")
		h.handleBlockchainSyncRequest(blockMsg.BlockchainSyncRequest.Hash, blockMsg.BlockchainSyncRequest.SenderAddress)
	case *block_chain.BlockMessage_BlocksResponse:
		log.Println("Handling BlocksRespone")
		h.handleBlocksResponse(blockMsg.BlocksResponse.Blocks)
	}
}

func (h *BlockMessageHandlerImpl) handleBlockchainSyncRequest(hash []byte, senderAddress []byte) {
	h.senderAddress = senderAddress
	log.Printf("Adres nadawcy: %s", senderAddress)

	blockNode := h.blockchain.GetBlock(hash)
	if blockNode == nil {
		log.Printf("Block with hash %x not found", hash)
		return
	}

	var blocks []*types.Block
	h.blockchain.TraverseTree(func(node *types.BlockNode) bool {
		if node.Block.Index > blockNode.Block.Index {
			blocks = append(blocks, node.Block)
		}
		return false
	})

	protoBlocks := make([]*block_chain.Block, len(blocks))
	for i, block := range blocks {
		protoBlocks[i] = block.ToProto()
	}

	blocksResponse := &block_chain.BlocksResponse{
		Blocks: protoBlocks,
	}

	// Przygotuj wiadomość do wysłania
	data, err := PrepareProtoMessageToSend(h.factory, blocksResponse)
	if err != nil {
		log.Printf("Failed to encode BlocksResponse: %v", err)
		return
	}

	// Wyślij wiadomość BlocksResponse do nadawcy
	err = h.messageSender.SendMsgToAddress(h.senderAddress, data)
	if err != nil {
		log.Printf("Failed to send BlocksResponse: %v", err)
	} else {
		log.Println("Successfully sent BlocksResponse")
	}
}

// handleBlocksResponse processes blocks response message.
func (h *BlockMessageHandlerImpl) handleBlocksResponse(protoBlocks []*block_chain.Block) {
	log.Println("Handling BlocksResponse")

	// Konwertuj proto bloki na typy bloków
	blocks := make([]*types.Block, len(protoBlocks))
	for i, protoBlock := range protoBlocks {
		blocks[i] = types.BlockFromProto(protoBlock)
	}

	// Sortuj bloki według indeksów
	sort.Slice(blocks, func(i, j int) bool {
		return blocks[i].Index < blocks[j].Index
	})

	// Dodaj bloki do blockchaina
	for _, block := range blocks {
		parent := h.blockchain.GetBlock(block.PreviousHash)
		if parent != nil {
			err := h.blockchain.AddBlock(parent, block)
			if err != nil {
				log.Printf("Failed to add block with index %d: %v", block.Index, err)
			} else {
				log.Printf("Successfully added block with index %d", block.Index)
			}
		} else {
			log.Printf("Parent block not found for block with index %d", block.Index)
		}
	}
}
