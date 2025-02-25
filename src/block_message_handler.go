package src

import (
	"fmt"
	"log"
	"sort"
	"strconv"

	"github.com/pabloaaa/GO_BLOCKCHAIN/interfaces"
	block_chain "github.com/pabloaaa/GO_BLOCKCHAIN/protos"
	"github.com/pabloaaa/GO_BLOCKCHAIN/types"
)

// BlockMessageHandlerImpl handles block-related messages.
type BlockMessageHandlerImpl struct {
	blockchain    interfaces.BlockchainInterface
	messageSender interfaces.MessageSender
	factory       *MessageFactory
	localPort     int
}

// NewBlockMessageHandler creates a new BlockMessageHandlerImpl.
func NewBlockMessageHandler(blockchain interfaces.BlockchainInterface, messageSender interfaces.MessageSender, localPort int) *BlockMessageHandlerImpl {
	return &BlockMessageHandlerImpl{
		blockchain:    blockchain,
		messageSender: messageSender,
		factory:       NewMessageFactory(),
		localPort:     localPort,
	}
}

// HandleBlockMessage processes incoming block messages.
func (h *BlockMessageHandlerImpl) HandleBlockMessage(msg *block_chain.BlockMessage) {
	log.Printf("\033[34mBlockMessageHandlerImpl: Received BlockMessage of type %T\033[0m", msg.BlockMessageType)
	switch blockMsg := msg.BlockMessageType.(type) {
	case *block_chain.BlockMessage_BlockchainSyncRequest:
		log.Println("\033[34mBlockMessageHandlerImpl: Handling BlockchainSyncRequest\033[0m")
		senderPort, _ := strconv.Atoi(string(blockMsg.BlockchainSyncRequest.SenderAddress))
		h.handleBlockchainSyncRequest(blockMsg.BlockchainSyncRequest.Hash, senderPort)
	case *block_chain.BlockMessage_BlocksResponse:
		log.Println("\033[34mBlockMessageHandlerImpl: Handling BlocksResponse\033[0m")
		h.handleBlocksResponse(blockMsg.BlocksResponse.Blocks)
	case *block_chain.BlockMessage_ApprovedBlock:
		log.Println("\033[34mBlockMessageHandlerImpl: Handling ApprovedBlock\033[0m")
		h.handleApprovedBlock(blockMsg.ApprovedBlock)
	default:
		log.Printf("\033[34mBlockMessageHandlerImpl: Unknown BlockMessageType: %T\033[0m", blockMsg)
	}
}

// handleBlockchainSyncRequest processes a blockchain sync request.
func (h *BlockMessageHandlerImpl) handleBlockchainSyncRequest(hash []byte, senderAddress int) {
	blockNode := h.blockchain.GetBlock(hash)
	if blockNode == nil {
		log.Printf("\033[34mBlockMessageHandlerImpl: Block with hash %x not found\033[0m", hash)
		return
	}

	var blocks []*types.Block
	h.blockchain.TraverseTree(func(node *types.BlockNode) bool {
		if blockNode.Block.Index >= node.Block.Index {
			blocks = append(blocks, node.Block)
		}
		return false
	})

	protoBlocks := make([]*block_chain.Block, len(blocks))
	for i, block := range blocks {
		protoBlocks[i] = block.ToProto()
	}

	log.Printf("\033[34mBlockMessageHandlerImpl: Length of protoBlocks before packing: %d\033[0m", len(protoBlocks))

	blocksResponse := &block_chain.BlocksResponse{
		Blocks: protoBlocks,
	}

	data, err := PrepareProtoMessageToSend(h.factory, blocksResponse)
	if err != nil {
		log.Printf("\033[34mBlockMessageHandlerImpl: Failed to encode BlocksResponse: %v\033[0m", err)
		return
	}

	err = h.messageSender.SendMsgToAddress(senderAddress, data, h.localPort)
	if err != nil {
		log.Printf("\033[34mBlockMessageHandlerImpl: Failed to send BlocksResponse: %v\033[0m", err)
	} else {
		log.Println("\033[34mBlockMessageHandlerImpl: Successfully sent BlocksResponse\033[0m")
	}
}

// handleBlocksResponse processes blocks response message.
func (h *BlockMessageHandlerImpl) handleBlocksResponse(protoBlocks []*block_chain.Block) {
	// Convert proto blocks to block types
	blocks := make([]*types.Block, len(protoBlocks))
	for i, protoBlock := range protoBlocks {
		blocks[i] = types.BlockFromProto(protoBlock)
	}
	// Sort blocks by index
	sort.Slice(blocks, func(i, j int) bool {
		return blocks[i].Index < blocks[j].Index
	})

	// Add blocks to the blockchain
	for _, block := range blocks {
		log.Printf("\033[34mBlockMessageHandlerImpl: Processing block with index %d\033[0m", block.Index)
		parent := h.blockchain.GetBlock(block.PreviousHash)
		if parent != nil {
			log.Printf("\033[34mBlockMessageHandlerImpl: Adding block with index %d\033[0m", block.Index)
			err := h.blockchain.AddBlock(parent, block)
			if err != nil {
				log.Printf("\033[34mBlockMessageHandlerImpl: Failed to add block with index %d: %v\033[0m", block.Index, err)
			} else {
				log.Printf("\033[34mBlockMessageHandlerImpl: Successfully added block with index %d\033[0m", block.Index)
			}
		} else {
			log.Printf("\033[34mBlockMessageHandlerImpl: Parent block not found for block with index %d\033[0m", block.Index)
		}
	}
}

// handleApprovedBlock processes approved block message.
func (h *BlockMessageHandlerImpl) handleApprovedBlock(approvedBlock *block_chain.ApprovedBlock) {
	block := types.BlockFromProto(approvedBlock.Block)
	if h.blockchain.GetBlockByIndex(block.Index) != nil {
		log.Printf("\033[34mBlockMessageHandlerImpl: Block with index %d already exists\033[0m", block.Index)
		return
	}

	senderAddress, _ := strconv.Atoi(string(approvedBlock.Address))
	log.Printf("\033[34mBlockMessageHandlerImpl: Block with index %d not found, sending sync request to %d\033[0m", block.Index, senderAddress)

	blockchainSyncRequest := &block_chain.BlockchainSyncRequest{
		Hash:          block.CalculateHash(),
		SenderAddress: []byte(fmt.Sprintf("%d", h.localPort)),
	}

	// Prepare message to send
	data, err := PrepareProtoMessageToSend(h.factory, blockchainSyncRequest)
	if err != nil {
		log.Printf("\033[34mBlockMessageHandlerImpl: Failed to encode BlockchainSyncRequest: %v\033[0m", err)
		return
	}

	// Send BlockchainSyncRequest message to sender
	err = h.messageSender.SendMsgToAddress(senderAddress, data, h.localPort)
	if err != nil {
		log.Printf("\033[34mBlockMessageHandlerImpl: Failed to send BlockchainSyncRequest: %v\033[0m", err)
	} else {
		log.Println("\033[34mBlockMessageHandlerImpl: Successfully sent BlockchainSyncRequest\033[0m")
	}
}

// BroadcastApprovedBlock broadcasts an approved block to all known nodes.
func (h *BlockMessageHandlerImpl) BroadcastApprovedBlock(block *types.Block, nodes []int) {
	approvedBlockMessage := &block_chain.ApprovedBlock{
		Block:   block.ToProto(),
		Address: []byte(fmt.Sprintf("%d", h.localPort)),
	}

	// Prepare message to send
	data, err := PrepareProtoMessageToSend(h.factory, approvedBlockMessage)
	if err != nil {
		log.Printf("\033[34mBlockMessageHandlerImpl: Failed to encode ApprovedBlock message: %v\033[0m", err)
		return
	}

	for _, nodeAddress := range nodes {
		err = h.messageSender.SendMsgToAddress(nodeAddress, data, h.localPort)
		if err != nil {
			log.Printf("\033[34mBlockMessageHandlerImpl: Failed to send ApprovedBlock message to node %d: %v\033[0m", nodeAddress, err)
		} else {
			log.Printf("\033[34mBlockMessageHandlerImpl: Successfully sent ApprovedBlock message to node %d\033[0m", nodeAddress)
		}
	}
}
