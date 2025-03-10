package src

import (
	"log"
	"strconv"

	"sort"

	"github.com/pabloaaa/GO_BLOCKCHAIN/interfaces"
	block_chain "github.com/pabloaaa/GO_BLOCKCHAIN/protos"
	"github.com/pabloaaa/GO_BLOCKCHAIN/types"
)

// BlockMessageHandlerImpl handles block-related messages.
type BlockMessageHandlerImpl struct {
	blockchain          interfaces.BlockchainInterface
	messageSender       interfaces.MessageSender
	factory             *MessageFactory
	localPort           int
	temporaryBlockChain map[int]map[uint64]*types.Block
}

// NewBlockMessageHandler creates a new BlockMessageHandlerImpl.
func NewBlockMessageHandler(blockchain interfaces.BlockchainInterface, messageSender interfaces.MessageSender, localPort int) *BlockMessageHandlerImpl {
	return &BlockMessageHandlerImpl{
		blockchain:          blockchain,
		messageSender:       messageSender,
		factory:             NewMessageFactory(),
		localPort:           localPort,
		temporaryBlockChain: make(map[int]map[uint64]*types.Block),
	}
}

// HandleBlockMessage processes incoming block messages.
func (h *BlockMessageHandlerImpl) HandleBlockMessage(msg *block_chain.BlockMessage) {
	log.Printf("\033[34mBlockMessageHandlerImpl[%d]: Received BlockMessage of type %T\033[0m", h.localPort, msg.BlockMessageType)
	switch blockMsg := msg.BlockMessageType.(type) {
	case *block_chain.BlockMessage_BlockchainSyncRequest:
		log.Printf("\033[34mBlockMessageHandlerImpl[%d]: Handling BlockchainSyncRequest\033[0m", h.localPort)
		h.handleBlockchainSyncRequest(blockMsg.BlockchainSyncRequest)
	case *block_chain.BlockMessage_ApprovedBlock:
		log.Printf("\033[34mBlockMessageHandlerImpl[%d]: Handling ApprovedBlock\033[0m", h.localPort)
		h.handleApprovedBlock(blockMsg.ApprovedBlock)
	case *block_chain.BlockMessage_BlockResponse:
		log.Printf("\033[34mBlockMessageHandlerImpl[%d]: Handling BlockResponse\033[0m", h.localPort)
		h.handleBlockResponse(blockMsg.BlockResponse)
	case *block_chain.BlockMessage_BlockRequest:
		log.Printf("\033[34mBlockMessageHandlerImpl[%d]: Handling BlockRequest\033[0m", h.localPort)
		h.handleBlockRequest(blockMsg.BlockRequest)
	default:
		log.Printf("\033[34mBlockMessageHandlerImpl[%d]: Unknown BlockMessageType: %T\033[0m", h.localPort, blockMsg)
	}
}

// handleBlockchainSyncRequest processes a blockchain sync request.
func (h *BlockMessageHandlerImpl) handleBlockchainSyncRequest(blockChainSyncRequest *block_chain.BlockchainSyncRequest) {
	bestBlock := h.blockchain.GetLatestBlock()
	log.Printf("\033[34mBlockMessageHandlerImpl[%d]: Best block index: %d\033[0m", h.localPort, bestBlock.Index)

	// Create a BlockResponse for the found block
	blockResponse := &block_chain.BlockResponse{
		Block:         bestBlock.ToProto(),
		SenderAddress: []byte(strconv.Itoa(h.localPort)),
	}

	if blockResponse.Block == nil {
		log.Printf("\033[34mBlockMessageHandlerImpl[%d]: blockResponse.Block is nil after conversion to proto\033[0m", h.localPort)
		return
	}

	// Prepare message to send
	data, err := PrepareProtoMessageToSend(h.factory, blockResponse)
	if err != nil {
		log.Printf("\033[34mBlockMessageHandlerImpl[%d]: Failed to encode BlockResponse: %v\033[0m", h.localPort, err)
		return
	}

	// Log the serialized data for debugging
	log.Printf("\033[34mBlockMessageHandlerImpl[%d]: Serialized BlockResponse data: %x\033[0m", h.localPort, data)

	// Send BlockResponse message to sender
	senderAddress, _ := strconv.Atoi(string(blockChainSyncRequest.SenderAddress))
	log.Printf("\033[34mBlockMessageHandlerImpl[%d]: Sending BlockResponse to address %d\033[0m", h.localPort, senderAddress)
	err = h.messageSender.SendMsgToAddress(senderAddress, data)
	if err != nil {
		log.Printf("\033[34mBlockMessageHandlerImpl[%d]: Failed to send BlockResponse: %v\033[0m", h.localPort, err)
	} else {
		log.Printf("\033[34mBlockMessageHandlerImpl[%d]: Successfully sent BlockResponse\033[0m", h.localPort)
	}
}

func (h *BlockMessageHandlerImpl) handleBlockResponse(blockResponse *block_chain.BlockResponse) {
	if blockResponse == nil {
		log.Printf("\033[34mBlockMessageHandlerImpl[%d]: blockResponse is nil\033[0m", h.localPort)
		return
	}
	if blockResponse.Block == nil {
		log.Printf("\033[34mBlockMessageHandlerImpl[%d]: blockResponse.Block is nil\033[0m", h.localPort)
		return
	}
	if blockResponse.SenderAddress == nil {
		log.Printf("\033[34mBlockMessageHandlerImpl[%d]: blockResponse.SenderAddress is nil\033[0m", h.localPort)
		return
	}
	log.Printf("\033[34mBlockMessageHandlerImpl[%d]: blockResponse.Block.Index: %d\033[0m", h.localPort, blockResponse.Block.Index)
	log.Printf("\033[34mBlockMessageHandlerImpl[%d]: blockResponse.SenderAddress: %s\033[0m", h.localPort, string(blockResponse.SenderAddress))

	block := types.BlockFromProto(blockResponse.Block)
	if block == nil {
		log.Printf("\033[34mBlockMessageHandlerImpl[%d]: block is nil after conversion from proto\033[0m", h.localPort)
		return
	}
	blockHash := block.CalculateHash()
	log.Printf("\033[34mBlockMessageHandlerImpl[%d]: Calculated block hash: %x\033[0m", h.localPort, blockHash)
	blockNode := h.blockchain.GetBlock(blockHash)
	if blockNode != nil {
		log.Printf("\033[34mBlockMessageHandlerImpl[%d]: Block with hash %x found\033[0m", h.localPort, blockHash)
		// Add all blocks from temporaryBlockChain for the given node port in the correct order
		senderAddress, err := strconv.Atoi(string(blockResponse.SenderAddress))
		if err != nil {
			log.Printf("\033[34mBlockMessageHandlerImpl[%d]: Error converting sender address %s to int: %v\033[0m", h.localPort, blockResponse.SenderAddress, err)
			return
		}
		if blocks, exists := h.temporaryBlockChain[senderAddress]; exists {
			var indices []uint64
			for index := range blocks {
				indices = append(indices, index)
			}
			sort.Slice(indices, func(i, j int) bool { return indices[i] < indices[j] })
			for _, index := range indices {
				tempBlock := blocks[index]
				parentBlockNode := h.blockchain.GetBlock(tempBlock.PreviousHash)
				if parentBlockNode != nil {
					err := h.blockchain.AddBlock(parentBlockNode, tempBlock)
					if err != nil {
						log.Printf("\033[34mBlockMessageHandlerImpl[%d]: Failed to add block with index %d: %v\033[0m", h.localPort, tempBlock.Index, err)
					} else {
						log.Printf("\033[34mBlockMessageHandlerImpl[%d]: Added block with index %d from temporaryBlockChain\033[0m", h.localPort, tempBlock.Index)
					}
				} else {
					log.Printf("\033[34mBlockMessageHandlerImpl[%d]: Parent block not found for block with index %d\033[0m", h.localPort, tempBlock.Index)
				}
			}
			delete(h.temporaryBlockChain, senderAddress)
		}
		return
	}

	// Add the block to the temporaryBlockChain
	senderAddress, err := strconv.Atoi(string(blockResponse.SenderAddress))
	if err != nil {
		log.Printf("\033[34mBlockMessageHandlerImpl[%d]: Error converting sender address %s to int: %v\033[0m", h.localPort, blockResponse.SenderAddress, err)
		return
	}
	if _, exists := h.temporaryBlockChain[senderAddress]; !exists {
		h.temporaryBlockChain[senderAddress] = make(map[uint64]*types.Block)
	}
	h.temporaryBlockChain[senderAddress][block.Index] = block
	log.Printf("\033[34mBlockMessageHandlerImpl[%d]: Added block with index %d to temporaryBlockChain\033[0m", h.localPort, block.Index)

	blockRequest := &block_chain.BlockRequest{
		Index:         block.Index - 1,
		SenderAddress: []byte(strconv.Itoa(h.localPort)),
	}
	log.Printf("\033[34mBlockMessageHandlerImpl[%d]: Sending BlockRequest for block index %d\033[0m", h.localPort, block.Index-1)
	// Prepare message to send
	data, err := PrepareProtoMessageToSend(h.factory, blockRequest)
	if err != nil {
		log.Printf("\033[34mBlockMessageHandlerImpl[%d]: Failed to encode BlockRequest: %v\033[0m", h.localPort, err)
		return
	}

	// Send BlockRequest message to sender
	err = h.messageSender.SendMsgToAddress(senderAddress, data)
	if err != nil {
		log.Printf("\033[34mBlockMessageHandlerImpl[%d]: Failed to send BlockRequest: %v\033[0m", h.localPort, err)
	} else {
		log.Printf("\033[34mBlockMessageHandlerImpl[%d]: Successfully sent BlockRequest\033[0m", h.localPort)
	}
}

func (h *BlockMessageHandlerImpl) handleApprovedBlock(approvedBlock *block_chain.ApprovedBlock) {
	log.Printf("\033[34mBlockMessageHandlerImpl[%d]: handleApprovedBlock called for block index %d\033[0m", h.localPort, approvedBlock.Block.Index)

	senderAddress, _ := strconv.Atoi(string(approvedBlock.SenderAddress))
	blockchainSyncRequest := &block_chain.BlockchainSyncRequest{
		SenderAddress: []byte(strconv.Itoa(h.localPort)),
	}

	// Prepare message to send
	data, err := PrepareProtoMessageToSend(h.factory, blockchainSyncRequest)
	if err != nil {
		log.Printf("\033[34mBlockMessageHandlerImpl[%d]: Failed to encode BlockchainSyncRequest: %v\033[0m", h.localPort, err)
		return
	}

	// Send BlockchainSyncRequest message to sender
	err = h.messageSender.SendMsgToAddress(senderAddress, data)
	if err != nil {
		log.Printf("\033[34mBlockMessageHandlerImpl[%d]: Failed to send BlockchainSyncRequest: %v\033[0m", h.localPort, err)
	} else {
		log.Printf("\033[34mBlockMessageHandlerImpl[%d]: Successfully sent BlockchainSyncRequest\033[0m", h.localPort)
	}
}

// BroadcastApprovedBlock broadcasts an approved block to all known nodes.
func (h *BlockMessageHandlerImpl) BroadcastApprovedBlock(block *types.Block, nodes []int) {
	approvedBlockMessage := &block_chain.ApprovedBlock{
		Block:         block.ToProto(),
		SenderAddress: []byte(strconv.Itoa(h.localPort)),
	}

	// Prepare message to send
	data, err := PrepareProtoMessageToSend(h.factory, approvedBlockMessage)
	if err != nil {
		log.Printf("\033[34mBlockMessageHandlerImpl[%d]: Failed to encode ApprovedBlock message: %v\033[0m", h.localPort, err)
		return
	}

	for _, nodeAddress := range nodes {
		log.Printf("\033[34mBlockMessageHandlerImpl[%d]: Sending ApprovedBlock to node %d\033[0m", h.localPort, nodeAddress)
		err = h.messageSender.SendMsgToAddress(nodeAddress, data)
		if err != nil {
			log.Printf("\033[34mBlockMessageHandlerImpl[%d]: Failed to send ApprovedBlock message to node %d: %v\033[0m", h.localPort, nodeAddress, err)
		} else {
			log.Printf("\033[34mBlockMessageHandlerImpl[%d]: Successfully sent ApprovedBlock message to node %d\033[0m", h.localPort, nodeAddress)
		}
	}
}

func (h *BlockMessageHandlerImpl) handleBlockRequest(blockRequest *block_chain.BlockRequest) {
	log.Printf("\033[34mBlockMessageHandlerImpl[%d]: handleBlockRequest called for block index %d\033[0m", h.localPort, blockRequest.Index)
	blockNode := h.blockchain.GetBlockByIndex(blockRequest.Index)
	if blockNode == nil {
		log.Printf("\033[34mBlockMessageHandlerImpl[%d]: Block with index %d not found\033[0m", h.localPort, blockRequest.Index)
		return
	}

	// Create a BlockResponse for the found block
	blockResponse := &block_chain.BlockResponse{
		Block:         blockNode.Block.ToProto(),
		SenderAddress: []byte(strconv.Itoa(h.localPort)),
	}

	// Prepare message to send
	data, err := PrepareProtoMessageToSend(h.factory, blockResponse)
	if err != nil {
		log.Printf("\033[34mBlockMessageHandlerImpl[%d]: Failed to encode BlockResponse: %v\033[0m", h.localPort, err)
		return
	}

	// Send BlockResponse message to sender
	senderAddress, _ := strconv.Atoi(string(blockRequest.SenderAddress))
	log.Printf("\033[34mBlockMessageHandlerImpl[%d]: Sending BlockResponse to address %d\033[0m", h.localPort, senderAddress)
	err = h.messageSender.SendMsgToAddress(senderAddress, data)
	if err != nil {
		log.Printf("\033[34mBlockMessageHandlerImpl[%d]: Failed to send BlockResponse: %v\033[0m", h.localPort, err)
	} else {
		log.Printf("\033[34mBlockMessageHandlerImpl[%d]: Successfully sent BlockResponse\033[0m", h.localPort)
	}
}
