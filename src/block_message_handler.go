package src

import (
	"log"
	"net"
	"time"

	block_chain "github.com/pabloaaa/GO_BLOCKCHAIN/protos"
	"google.golang.org/protobuf/proto"
)

type BlockMessageHandlerImpl struct {
	blockchain *Blockchain
}

func NewBlockMessageHandler(blockchain *Blockchain) *BlockMessageHandlerImpl {
	return &BlockMessageHandlerImpl{blockchain: blockchain}
}

func (h *BlockMessageHandlerImpl) HandleBlockMessage(msg *block_chain.BlockMessage, conn net.Conn) {
	switch blockMsg := msg.BlockMessageType.(type) {
	case *block_chain.BlockMessage_GetLatestBlockRequest:
		h.handleGetLatestBlock(nil, conn.LocalAddr().String())
	case *block_chain.BlockMessage_GetBlockRequest_:
		h.handleGetBlockRequest(blockMsg.GetBlockRequest_.Hash, conn.LocalAddr().String())
	case *block_chain.BlockMessage_BlockResponse:
		h.handleBlockResponse(blockMsg.BlockResponse.Message, conn.LocalAddr().String())
	}
}

func (h *BlockMessageHandlerImpl) handleGetLatestBlock(data []byte, address string) {
	getLatestBlockRequest := &block_chain.GetLatestBlockRequest{}
	err := proto.Unmarshal(data, getLatestBlockRequest)
	if err != nil {
		log.Println(err)
		return
	}
	h.SendLatestBlock(address)
}

func (h *BlockMessageHandlerImpl) handleGetBlockRequest(hash []byte, address string) {
	getBlockRequest := &block_chain.GetBlockRequest{Hash: hash}
	block := h.blockchain.GetBlock(getBlockRequest.GetHash())
	if block != nil {
		h.SendBlock(address, block)
	}
}

func (h *BlockMessageHandlerImpl) handleBlockResponse(data []byte, address string) {
	blockResponse := &block_chain.BlockResponse{}
	err := proto.Unmarshal(data, blockResponse)
	if err != nil {
		log.Println(err)
		return
	}
	block := BlockFromProto(blockResponse.GetBlock())
	blockHash := block.CalculateHash()
	if !h.blockchain.BlockExists(blockHash) {
		h.GetBlock(address, block.PreviousHash)
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
				}
			}
		}
	}
}

func (h *BlockMessageHandlerImpl) SendBlock(address string, blockNode *BlockNode) {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	// Convert block to block_chain.Block format
	protoBlock := blockNode.Block.ToProto()

	blockResponse := &block_chain.BlockResponse{
		Success: true,
		Message: []byte("Block"),
		Block:   protoBlock,
	}

	err = EncodeMessage(conn, blockResponse)
	if err != nil {
		log.Fatal(err)
	}
}

func (h *BlockMessageHandlerImpl) SendLatestBlock(address string) {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	latestBlock := h.blockchain.GetLatestBlock()

	// Convert latestBlock to block_chain.Block format
	protoBlock := latestBlock.ToProto()

	blockResponse := &block_chain.BlockResponse{
		Success: true,
		Message: []byte("Latest block"),
		Block:   protoBlock,
	}

	err = EncodeMessage(conn, blockResponse)
	if err != nil {
		log.Fatal(err)
	}
}

func (h *BlockMessageHandlerImpl) GetBlock(address string, blockHash []byte) {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		log.Printf("Failed to dial node at address %s: %v", address, err)
		return
	}
	defer conn.Close()

	getBlockRequest := &block_chain.GetBlockRequest{
		Hash: blockHash,
	}

	err = EncodeMessage(conn, getBlockRequest)
	if err != nil {
		log.Printf("Failed to encode message: %v", err)
	}
}

func (h *BlockMessageHandlerImpl) GetLatestBlock(address string) {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		log.Printf("Failed to dial node at address %s: %v", address, err)
		return
	}
	defer conn.Close()

	emptyMessage := &block_chain.Empty{}
	err = EncodeMessage(conn, emptyMessage)
	if err != nil {
		log.Printf("Failed to encode message: %v", err)
	}
}

func (h *BlockMessageHandlerImpl) BroadcastLatestBlock(nodes [][]byte) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		for _, node := range nodes {
			h.GetLatestBlock(string(node))
		}
	}
}
