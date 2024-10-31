package src

import (
	"log"

	"github.com/pabloaaa/GO_BLOCKCHAIN/interfaces"
	block_chain "github.com/pabloaaa/GO_BLOCKCHAIN/protos"
	"github.com/pabloaaa/GO_BLOCKCHAIN/types"
	"google.golang.org/protobuf/proto"
)

type BlockMessageHandlerImpl struct {
	blockchain    interfaces.BlockchainInterface
	messageSender interfaces.MessageSender
}

func NewBlockMessageHandler(blockchain interfaces.BlockchainInterface, messageSender interfaces.MessageSender) *BlockMessageHandlerImpl {
	return &BlockMessageHandlerImpl{blockchain: blockchain, messageSender: messageSender}
}

func (h *BlockMessageHandlerImpl) HandleBlockMessage(msg *block_chain.BlockMessage) {
	switch blockMsg := msg.BlockMessageType.(type) {
	case *block_chain.BlockMessage_GetLatestBlockRequest:
		h.handleGetLatestBlock(nil)
	case *block_chain.BlockMessage_GetBlockRequest_:
		h.handleGetBlockRequest(blockMsg.GetBlockRequest_.Hash)
	case *block_chain.BlockMessage_BlockResponse:
		h.handleBlockResponse(blockMsg.BlockResponse.Message)
	}
}

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

func (h *BlockMessageHandlerImpl) handleGetBlockRequest(hash []byte) {
	getBlockRequest := &block_chain.GetBlockRequest{Hash: hash}
	block := h.blockchain.GetBlock(getBlockRequest.GetHash())
	if block != nil {
		h.SendBlock(block)
	}
}

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

func (h *BlockMessageHandlerImpl) SendLatestBlock() {
	latestBlock := h.blockchain.GetLatestBlock()

	protoBlock := latestBlock.ToProto()

	blockResponse := &block_chain.BlockResponse{
		Success: true,
		Message: []byte("Latest block"),
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
	emptyMessage := &block_chain.Empty{}
	data, err := EncodeMessage(emptyMessage)
	if err != nil {
		log.Printf("Failed to encode message: %v", err)
		return
	}

	err = h.messageSender.SendMsg(data)
	if err != nil {
		log.Printf("Failed to send message: %v", err)
	}
}

func (h *BlockMessageHandlerImpl) BroadcastLatestBlock(nodes [][]byte) {
	// ticker := time.NewTicker(5 * time.Second)
	// defer ticker.Stop()

	// for range ticker.C {
	// 	for _, node := range nodes {
	// 		h.GetLatestBlock()
	// 	}
	// }
}
