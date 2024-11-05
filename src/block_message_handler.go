package src

import (
	"log"

	"github.com/pabloaaa/GO_BLOCKCHAIN/interfaces"
	block_chain "github.com/pabloaaa/GO_BLOCKCHAIN/protos"
	"github.com/pabloaaa/GO_BLOCKCHAIN/types"
)

type BlockMessageHandlerImpl struct {
	blockchain    interfaces.BlockchainInterface
	messageSender interfaces.MessageSender
}

func NewBlockMessageHandler(blockchain interfaces.BlockchainInterface, messageSender interfaces.MessageSender) *BlockMessageHandlerImpl {
	return &BlockMessageHandlerImpl{blockchain: blockchain, messageSender: messageSender}
}

func (h *BlockMessageHandlerImpl) HandleBlockMessage(msg *block_chain.BlockMessage) {
	response, err := h.HandleRequest(msg)
	if err != nil {
		log.Println("Error handling block message:", err)
		return
	}
	if response != nil {
		data, err := EncodeMessage(response)
		if err != nil {
			log.Println("Failed to encode response message:", err)
			return
		}
		err = h.messageSender.SendMsg(data)
		if err != nil {
			log.Println("Failed to send response message:", err)
		}
	}
}

func (h *BlockMessageHandlerImpl) HandleRequest(msg *block_chain.BlockMessage) (*block_chain.BlockMessage, error) {
	switch blockMsg := msg.BlockMessageType.(type) {
	case *block_chain.BlockMessage_GetLatestBlockRequest:
		return h.handleGetLatestBlock(blockMsg.GetLatestBlockRequest)
	case *block_chain.BlockMessage_GetBlockRequest_:
		return h.handleGetBlockRequest(blockMsg.GetBlockRequest_.Hash)
	case *block_chain.BlockMessage_BlockResponse:
		return h.handleBlockResponse(blockMsg.BlockResponse)
	}
	return nil, nil
}

func (h *BlockMessageHandlerImpl) handleGetLatestBlock(request *block_chain.GetLatestBlockRequest) (*block_chain.BlockMessage, error) {
	latestBlock := h.blockchain.GetLatestBlock()
	protoBlock := latestBlock.ToProto()
	blockResponse := &block_chain.BlockResponse{
		Success: true,
		Message: []byte("Latest block"),
		Block:   protoBlock,
	}

	blockMessage := &block_chain.BlockMessage{
		BlockMessageType: &block_chain.BlockMessage_BlockResponse{
			BlockResponse: blockResponse,
		},
	}
	return blockMessage, nil
}

func (h *BlockMessageHandlerImpl) handleGetBlockRequest(hash []byte) (*block_chain.BlockMessage, error) {
	getBlockRequest := &block_chain.GetBlockRequest{Hash: hash}
	block := h.blockchain.GetBlock(getBlockRequest.GetHash())
	if block != nil {
		h.SendBlock(block)
	}
	return nil, nil
}

func (h *BlockMessageHandlerImpl) handleBlockResponse(blockResponse *block_chain.BlockResponse) (*block_chain.BlockMessage, error) {
	block := types.BlockFromProto(blockResponse.GetBlock())
	blockHash := block.CalculateHash()
	if !h.blockchain.BlockExists(blockHash) {
		h.GetBlock(block.PreviousHash)
		return nil, nil
	}

	parent := h.blockchain.GetBlock(block.PreviousHash)
	if parent == nil {
		return nil, nil
	}

	// Validate the block before adding it to the blockchain
	err := h.blockchain.ValidateBlock(block, parent.Block)
	if err != nil {
		log.Println("Received invalid block: ", err)
		return nil, nil
	}

	err = h.blockchain.AddBlock(parent, block)
	if err != nil {
		log.Println(err)
		return nil, nil
	}

	return nil, nil
}

func (h *BlockMessageHandlerImpl) SendBlock(blockNode *types.BlockNode) {
	protoBlock := blockNode.Block.ToProto()

	blockResponse := &block_chain.BlockResponse{
		Success: true,
		Message: []byte("Block"),
		Block:   protoBlock,
	}

	data, err := EncodeMessage(blockResponse)
	if err != nil {
		log.Fatal(err)
	}

	err = h.messageSender.SendMsg(data)
	if err != nil {
		log.Fatal(err)
	}
}

func (h *BlockMessageHandlerImpl) GetBlock(blockHash []byte) {
	getBlockRequest := &block_chain.GetBlockRequest{
		Hash: blockHash,
	}

	data, err := EncodeMessage(getBlockRequest)
	if err != nil {
		log.Printf("Failed to encode message: %v", err)
		return
	}

	err = h.messageSender.SendMsg(data)
	if err != nil {
		log.Printf("Failed to send message: %v", err)
	}
}

func (h *BlockMessageHandlerImpl) GetLatestBlock() {
	getLatestBlockRequest := &block_chain.GetLatestBlockRequest{}
	data, err := EncodeMessage(getLatestBlockRequest)
	if err != nil {
		log.Printf("Failed to encode message: %v", err)
		return
	}

	err = h.messageSender.SendMsg(data)
	if err != nil {
		log.Printf("Failed to send message: %v", err)
	}
}
