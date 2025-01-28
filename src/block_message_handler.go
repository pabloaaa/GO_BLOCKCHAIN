package src

import (
	"log"

	"github.com/pabloaaa/GO_BLOCKCHAIN/interfaces"
	block_chain "github.com/pabloaaa/GO_BLOCKCHAIN/protos"
	"github.com/pabloaaa/GO_BLOCKCHAIN/types"
	"google.golang.org/protobuf/proto"
)

// BlockMessageHandlerImpl handles block-related messages.
type BlockMessageHandlerImpl struct {
	blockchain    interfaces.BlockchainInterface
	messageSender interfaces.MessageSender
}

// NewBlockMessageHandler creates a new BlockMessageHandlerImpl.
func NewBlockMessageHandler(blockchain interfaces.BlockchainInterface, messageSender interfaces.MessageSender) *BlockMessageHandlerImpl {
	return &BlockMessageHandlerImpl{blockchain: blockchain, messageSender: messageSender}
}

// HandleBlockMessage processes incoming block messages.
func (h *BlockMessageHandlerImpl) HandleBlockMessage(msg *block_chain.BlockMessage) {
	switch blockMsg := msg.BlockMessageType.(type) {
	case *block_chain.BlockMessage_GetLatestBlockRequest:
		log.Println("Handling GetLatestBlockRequest")
		h.handleGetLatestBlock(nil)
	case *block_chain.BlockMessage_GetBlockRequest_:
		log.Println("Handling GetBlockRequest")
		h.handleGetBlockRequest(blockMsg.GetBlockRequest_.Hash)
	case *block_chain.BlockMessage_BlockResponse:
		log.Println("Handling BlockResponse")
		h.handleBlockResponse(blockMsg.BlockResponse.Message)
	case *block_chain.BlockMessage_LatestBlockResponse:
		log.Println("Handling LatestBlockResponse")
		h.handleLatestBlockResponse(blockMsg.LatestBlockResponse.Block)
	}
}

// handleGetLatestBlock processes a request for the latest block.
func (h *BlockMessageHandlerImpl) handleGetLatestBlock(data []byte) {
	if data != nil {
		getLatestBlockRequest := &block_chain.GetLatestBlockRequest{}
		err := proto.Unmarshal(data, getLatestBlockRequest)
		if err != nil {
			log.Println(err)
			return
		}
	}
	h.SendLatestBlock()
}

// handleGetBlockRequest processes a request for a specific block.
func (h *BlockMessageHandlerImpl) handleGetBlockRequest(hash []byte) {
	getBlockRequest := &block_chain.GetBlockRequest{Hash: hash}
	block := h.blockchain.GetBlock(getBlockRequest.GetHash())
	if block != nil {
		h.SendBlock(block)
	}
}

// handleBlockResponse processes a block response message.
func (h *BlockMessageHandlerImpl) handleBlockResponse(data []byte) {
	blockResponse := &block_chain.BlockResponse{}
	err := proto.Unmarshal(data, blockResponse)
	if err != nil {
		log.Println(err)
		return
	}
	block := types.BlockFromProto(blockResponse.GetBlock())
	blockHash := block.CalculateHash()
	if !h.blockchain.BlockExists(blockHash) {
		h.GetBlock(block.PreviousHash)
	} else {
		parent := h.blockchain.GetBlock(block.PreviousHash)
		if parent != nil {
			// Validate the block before adding it to the blockchain
			err := h.blockchain.ValidateBlock(block, parent.Block)
			if err != nil {
				log.Println("Received invalid block: ", err)
			} else {
				err := h.blockchain.AddBlock(parent, block)
				if err != nil {
					log.Println(err)
				} else {
					// Send a success message
					h.SendBlock(parent)
				}
			}
		}
	}
}

// handleLatestBlockResponse processes a latest block response message.
func (h *BlockMessageHandlerImpl) handleLatestBlockResponse(protoBlock *block_chain.Block) {
	block := types.BlockFromProto(protoBlock)
	blockHash := block.CalculateHash()
	if !h.blockchain.BlockExists(blockHash) {
		h.blockchain.ReplaceBlocks([]*types.Block{block})
		log.Println("Blockchain synchronized with the latest block")
	} else {
		log.Println("Received block already exists in the blockchain")
	}
}

// SendBlock sends a block to the message sender.
func (h *BlockMessageHandlerImpl) SendBlock(blockNode *types.BlockNode) {
	protoBlock := blockNode.Block.ToProto()

	blockResponse := &block_chain.BlockResponse{
		Success: true,
		Message: []byte("Block"),
		Block:   protoBlock,
	}

	data, err := proto.Marshal(blockResponse)
	if err != nil {
		log.Fatal(err)
	}

	err = h.messageSender.SendMsg(data)
	if err != nil {
		log.Fatal(err)
	}
}

// SendLatestBlock sends the latest block to the message sender.
func (h *BlockMessageHandlerImpl) SendLatestBlock() {
	latestBlock := h.blockchain.GetLatestBlock()
	log.Printf("Sending latest block with index: %d", latestBlock.Index)

	protoBlock := latestBlock.ToProto()

	blockResponse := &block_chain.BlockResponse{
		Success: true,
		Message: []byte("Latest block"),
		Block:   protoBlock,
	}

	data, err := proto.Marshal(blockResponse)
	if err != nil {
		log.Fatal(err)
	}

	err = h.messageSender.SendMsg(data)
	if err != nil {
		log.Fatal(err)
	}
}

// GetBlock requests a block by its hash.
func (h *BlockMessageHandlerImpl) GetBlock(blockHash []byte) {
	getBlockRequest := &block_chain.GetBlockRequest{
		Hash: blockHash,
	}

	data, err := proto.Marshal(getBlockRequest)
	if err != nil {
		log.Printf("Failed to encode message: %v", err)
		return
	}

	err = h.messageSender.SendMsg(data)
	if err != nil {
		log.Printf("Failed to send message: %v", err)
	}
}

// GetLatestBlock requests the latest block.
func (h *BlockMessageHandlerImpl) GetLatestBlock() {
	getLatestBlockRequest := &block_chain.GetLatestBlockRequest{}
	data, err := proto.Marshal(getLatestBlockRequest)
	if err != nil {
		log.Printf("Failed to encode message: %v", err)
		return
	}

	err = h.messageSender.SendMsg(data)
	if err != nil {
		log.Printf("Failed to send message: %v", err)
	}
}

// BroadcastLatestBlock broadcasts the latest block to all nodes.
func (h *BlockMessageHandlerImpl) BroadcastLatestBlock(nodes [][]byte) {
	// ticker := time.NewTicker(5 * time.Second)
	// defer ticker.Stop()

	// for range ticker.C {
	// 	for _, node := range nodes {
	// 		h.GetLatestBlock()
	// 	}
	// }
}
