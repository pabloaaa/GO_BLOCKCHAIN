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

func (h *BlockMessageHandlerImpl) handleRequest(msg *block_chain.BlockMessage) (*block_chain.BlockMessage, error) {

  // ideally that should loook more like

  switch blockMsg := msg.BlockMessageType.(type) {
    case *block_chain.BlockMessage_GetLatestBlockRequest:
      return h.handleGetLatestBlock(blockMsg.GetLatestBlockRequest) // 
    case *block_chain.BlockMessage_GetBlockRequest_:
      h.handleGetBlockRequest(blockMsg.GetBlockRequest_.Hash)
      return nil, nil
    case *block_chain.BlockMessage_BlockResponse: {
      h.handleBlockResponse(blockMsg.BlockResponse)
      return nil, nil
    }
  }
  return nil, nil
}


func (h *BlockMessageHandlerImpl) HandleBlockMessage(msg *block_chain.BlockMessage) {


  // hangle request and get response to send back
  response, err := h.handleRequest(msg)
  if err != nil {
    log.Println("Failed to handle request: %v", err)
  }

  // encode and send happens only in this single place instead of being called in every message handler
	data, err := EncodeMessage(response)
  if err != nil {
    log.Println("Failed to encode message: %v", err)
  }
	err = h.messageSender.SendMsg(data)
}

// providing return type for handler fn has  nubmer of benefits:
// - you can not forget to send the message, you need to return sth and then its send in HanleBlockMessage
// - compiler forces you to provide return value
// - function has single responsibility - handling message and returning response
func (h *BlockMessageHandlerImpl) handleGetLatestBlock(req *block_chain.GetLatestBlockRequest) (*block_chain.BlockMessage, error) {

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

func (h *BlockMessageHandlerImpl) handleGetBlockRequest(hash []byte) {
	getBlockRequest := &block_chain.GetBlockRequest{Hash: hash}
	block := h.blockchain.GetBlock(getBlockRequest.GetHash())
	if block != nil {
		h.SendBlock(block)
	}
}

// Hanlde block response should handle blockResponse type not bytes, 
// functioon should only have 1 responsibility, originally this function do:
// - deserialize proto message (and handle failures)
// - handling the message ang generate response
// - serialize the response
// - send the response
//
func (h *BlockMessageHandlerImpl) handleBlockResponse(blockResponse *block_chain.BlockResponse) {
	block := types.BlockFromProto(blockResponse.GetBlock())
	if !h.blockchain.BlockExists(blockResponse.Block.Hash) {
		h.GetBlock(block.PreviousHash)
	} else {
    // prefer to use early return
    parent := h.blockchain.GetBlock(block.PreviousHash)
    if parent == nil {
      return
    }
    // Validate the block before adding it to the blockchain
    err := h.blockchain.ValidateBlock(block, parent.Block)
    if err != nil {
      log.Println("Received invalid block: ", err)
      return;
    } 

    err = h.blockchain.AddBlock(parent, block)
    if err != nil {
      log.Println(err)
      return
    }

    // this is clearly bad, as handling blockResponse returns another block response ????
    h.SendBlock(parent)
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
