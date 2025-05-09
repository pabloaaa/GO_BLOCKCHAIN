package src

import (
	"fmt"
	"sort"
	"strconv"

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
	Info(fmt.Sprintf("BlockMessageHandlerImpl[%d]: Received BlockMessage of type %T", h.localPort, msg.BlockMessageType))
	switch blockMsg := msg.BlockMessageType.(type) {
	case *block_chain.BlockMessage_BlockchainSyncRequest:
		Debug(fmt.Sprintf("BlockMessageHandlerImpl[%d]: Handling BlockchainSyncRequest", h.localPort))
		h.handleBlockchainSyncRequest(blockMsg.BlockchainSyncRequest)
	case *block_chain.BlockMessage_ApprovedBlock:
		Debug(fmt.Sprintf("BlockMessageHandlerImpl[%d]: Handling ApprovedBlock", h.localPort))
		h.handleApprovedBlock(blockMsg.ApprovedBlock)
	case *block_chain.BlockMessage_BlockResponse:
		Debug(fmt.Sprintf("BlockMessageHandlerImpl[%d]: Handling BlockResponse", h.localPort))
		h.handleBlockResponse(blockMsg.BlockResponse)
	case *block_chain.BlockMessage_BlockRequest:
		Debug(fmt.Sprintf("BlockMessageHandlerImpl[%d]: Handling BlockRequest", h.localPort))
		h.handleBlockRequest(blockMsg.BlockRequest)
	default:
		Error(fmt.Sprintf("BlockMessageHandlerImpl[%d]: Unknown BlockMessageType: %T", h.localPort, blockMsg))
	}
}

// handleBlockchainSyncRequest processes a blockchain sync request.
func (h *BlockMessageHandlerImpl) handleBlockchainSyncRequest(blockChainSyncRequest *block_chain.BlockchainSyncRequest) {
	bestBlock := h.blockchain.GetLatestBlock()
	Debug(fmt.Sprintf("BlockMessageHandlerImpl[%d]: Best block index: %d", h.localPort, bestBlock.Index))

	// Create a BlockResponse for the found block
	blockResponse := &block_chain.BlockResponse{
		Block:         bestBlock.ToProto(),
		SenderAddress: []byte(strconv.Itoa(h.localPort)),
	}

	if blockResponse.Block == nil {
		Error(fmt.Sprintf("BlockMessageHandlerImpl[%d]: blockResponse.Block is nil after conversion to proto", h.localPort))
		return
	}

	// Prepare message to send
	data, err := PrepareProtoMessageToSend(h.factory, blockResponse)
	if err != nil {
		Error(fmt.Sprintf("BlockMessageHandlerImpl[%d]: Failed to encode BlockResponse: %v", h.localPort, err))
		return
	}

	// Log the serialized data for debugging
	Debug(fmt.Sprintf("BlockMessageHandlerImpl[%d]: Serialized BlockResponse data: %x", h.localPort, data))

	// Send BlockResponse message to sender
	senderAddress, _ := strconv.Atoi(string(blockChainSyncRequest.SenderAddress))
	Debug(fmt.Sprintf("BlockMessageHandlerImpl[%d]: Sending BlockResponse to address %d", h.localPort, senderAddress))
	err = h.messageSender.SendMsgToAddress(senderAddress, data)
	if err != nil {
		Error(fmt.Sprintf("BlockMessageHandlerImpl[%d]: Failed to send BlockResponse: %v", h.localPort, err))
	} else {
		Info(fmt.Sprintf("BlockMessageHandlerImpl[%d]: Successfully sent BlockResponse", h.localPort))
	}
}

func (h *BlockMessageHandlerImpl) handleBlockResponse(blockResponse *block_chain.BlockResponse) {
	block := types.BlockFromProto(blockResponse.Block)
	if block == nil {
		Error(fmt.Sprintf("BlockMessageHandlerImpl[%d]: block is nil after conversion from proto", h.localPort))
		return
	}
	blockHash := block.CalculateHash()
	Debug(fmt.Sprintf("BlockMessageHandlerImpl[%d]: Calculated block hash: %x", h.localPort, blockHash))
	blockNode := h.blockchain.GetBlock(blockHash)
	if blockNode != nil {
		Debug(fmt.Sprintf("BlockMessageHandlerImpl[%d]: Block with hash %x found", h.localPort, blockHash))
		// Add all blocks from temporaryBlockChain for the given node port in the correct order
		senderAddress, err := strconv.Atoi(string(blockResponse.SenderAddress))
		if err != nil {
			Error(fmt.Sprintf("BlockMessageHandlerImpl[%d]: Error converting sender address %s to int: %v", h.localPort, blockResponse.SenderAddress, err))
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
						Error(fmt.Sprintf("BlockMessageHandlerImpl[%d]: Failed to add block with index %d: %v", h.localPort, tempBlock.Index, err))
					} else {
						Debug(fmt.Sprintf("BlockMessageHandlerImpl[%d]: Added block with index %d from temporaryBlockChain", h.localPort, tempBlock.Index))
					}
				} else {
					Error(fmt.Sprintf("BlockMessageHandlerImpl[%d]: Parent block not found for block with index %d", h.localPort, tempBlock.Index))
				}
			}
			delete(h.temporaryBlockChain, senderAddress)
		}
		return
	}

	// Add the block to the temporaryBlockChain
	senderAddress, err := strconv.Atoi(string(blockResponse.SenderAddress))
	if err != nil {
		Error(fmt.Sprintf("BlockMessageHandlerImpl[%d]: Error converting sender address %s to int: %v", h.localPort, blockResponse.SenderAddress, err))
		return
	}
	if _, exists := h.temporaryBlockChain[senderAddress]; !exists {
		h.temporaryBlockChain[senderAddress] = make(map[uint64]*types.Block)
	}
	h.temporaryBlockChain[senderAddress][block.Index] = block
	Debug(fmt.Sprintf("BlockMessageHandlerImpl[%d]: Added block with index %d to temporaryBlockChain", h.localPort, block.Index))

	blockRequest := &block_chain.BlockRequest{
		Index:         block.Index - 1,
		SenderAddress: []byte(strconv.Itoa(h.localPort)),
	}
	Debug(fmt.Sprintf("BlockMessageHandlerImpl[%d]: Sending BlockRequest for block index %d", h.localPort, block.Index-1))
	// Prepare message to send
	data, err := PrepareProtoMessageToSend(h.factory, blockRequest)
	if err != nil {
		Error(fmt.Sprintf("BlockMessageHandlerImpl[%d]: Failed to encode BlockRequest: %v", h.localPort, err))
		return
	}

	// Send BlockRequest message to sender
	err = h.messageSender.SendMsgToAddress(senderAddress, data)
	if err != nil {
		Error(fmt.Sprintf("BlockMessageHandlerImpl[%d]: Failed to send BlockRequest: %v", h.localPort, err))
	} else {
		Info(fmt.Sprintf("BlockMessageHandlerImpl[%d]: Successfully sent BlockRequest", h.localPort))
	}
}

func (h *BlockMessageHandlerImpl) handleApprovedBlock(approvedBlock *block_chain.ApprovedBlock) {
	Debug(fmt.Sprintf("BlockMessageHandlerImpl[%d]: handleApprovedBlock called for block index %d", h.localPort, approvedBlock.Block.Index))

	senderAddress, _ := strconv.Atoi(string(approvedBlock.SenderAddress))
	blockchainSyncRequest := &block_chain.BlockchainSyncRequest{
		SenderAddress: []byte(strconv.Itoa(h.localPort)),
	}

	// Prepare message to send
	data, err := PrepareProtoMessageToSend(h.factory, blockchainSyncRequest)
	if err != nil {
		Error(fmt.Sprintf("BlockMessageHandlerImpl[%d]: Failed to encode BlockchainSyncRequest: %v", h.localPort, err))
		return
	}

	// Send BlockchainSyncRequest message to sender
	err = h.messageSender.SendMsgToAddress(senderAddress, data)
	if err != nil {
		Error(fmt.Sprintf("BlockMessageHandlerImpl[%d]: Failed to send BlockchainSyncRequest: %v", h.localPort, err))
	} else {
		Info(fmt.Sprintf("BlockMessageHandlerImpl[%d]: Successfully sent BlockchainSyncRequest", h.localPort))
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
		Error(fmt.Sprintf("BlockMessageHandlerImpl[%d]: Failed to encode ApprovedBlock message: %v", h.localPort, err))
		return
	}

	for _, nodeAddress := range nodes {
		Debug(fmt.Sprintf("BlockMessageHandlerImpl[%d]: Sending ApprovedBlock to node %d", h.localPort, nodeAddress))
		err = h.messageSender.SendMsgToAddress(nodeAddress, data)
		if err != nil {
			Error(fmt.Sprintf("BlockMessageHandlerImpl[%d]: Failed to send ApprovedBlock message to node %d: %v", h.localPort, nodeAddress, err))
		} else {
			Info(fmt.Sprintf("BlockMessageHandlerImpl[%d]: Successfully sent ApprovedBlock message to node %d", h.localPort, nodeAddress))
		}
	}
}

func (h *BlockMessageHandlerImpl) handleBlockRequest(blockRequest *block_chain.BlockRequest) {
	Debug(fmt.Sprintf("BlockMessageHandlerImpl[%d]: handleBlockRequest called for block index %d", h.localPort, blockRequest.Index))
	blockNode := h.blockchain.GetBlockByIndex(blockRequest.Index)
	if blockNode == nil {
		Error(fmt.Sprintf("BlockMessageHandlerImpl[%d]: Block with index %d not found", h.localPort, blockRequest.Index))
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
		Error(fmt.Sprintf("BlockMessageHandlerImpl[%d]: Failed to encode BlockResponse: %v", h.localPort, err))
		return
	}

	// Send BlockResponse message to sender
	senderAddress, _ := strconv.Atoi(string(blockRequest.SenderAddress))
	Debug(fmt.Sprintf("BlockMessageHandlerImpl[%d]: Sending BlockResponse to address %d", h.localPort, senderAddress))
	err = h.messageSender.SendMsgToAddress(senderAddress, data)
	if err != nil {
		Error(fmt.Sprintf("BlockMessageHandlerImpl[%d]: Failed to send BlockResponse: %v", h.localPort, err))
	} else {
		Info(fmt.Sprintf("BlockMessageHandlerImpl[%d]: Successfully sent BlockResponse", h.localPort))
	}
}
