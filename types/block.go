package types

import (
	"crypto/sha256"
	"strconv"

	pb "github.com/pabloaaa/GO_BLOCKCHAIN/protos"
)

// Block represents a single block in the blockchain.
type Block struct {
	Index        uint64
	Timestamp    uint64
	PreviousHash []byte
	Data         uint64
	Checkpoint   bool
	Messages     []string // New field to store messages in the block
}

// BlockNode represents a node in the blockchain tree.
type BlockNode struct {
	Block  *Block
	Parent *BlockNode
	Childs []*BlockNode
}

// Message represents a message sent between nodes.
type Message struct {
	Sender   []byte
	Receiver []byte
	Content  string
}

// CalculateHash calculates the SHA-256 hash of the block.
func (b *Block) CalculateHash() []byte {
	data := strconv.FormatUint(b.Index, 10) + strconv.FormatUint(b.Timestamp, 10) + string(b.PreviousHash) + strconv.FormatUint(b.Data, 10)
	// Include messages in hash calculation
	for _, msg := range b.Messages {
		data += msg
	}
	hash := sha256.Sum256([]byte(data))
	return hash[:]
}

// BlockFromProto converts a protobuf Block to a Block.
func BlockFromProto(pbBlock *pb.Block) *Block {
	// Convert protobuf messages to string array
	var messages []string
	for _, msg := range pbBlock.GetMessages() {
		messages = append(messages, msg.GetContent())
	}
	
	return &Block{
		Index:        pbBlock.GetIndex(),
		Timestamp:    pbBlock.GetTimestamp(),
		PreviousHash: pbBlock.GetPreviousHash(),
		Data:         pbBlock.GetData(),
		Checkpoint:   pbBlock.GetCheckpoint(),
		Messages:     messages,
	}
}

// ToProto converts a Block to a protobuf Block.
func (b *Block) ToProto() *pb.Block {
	// Convert messages to protobuf ChatMessage array
	var messages []*pb.ChatMessage
	for _, msg := range b.Messages {
		chatMsg := &pb.ChatMessage{
			Content: msg,
		}
		messages = append(messages, chatMsg)
	}
	
	return &pb.Block{
		Index:        b.Index,
		Timestamp:    b.Timestamp,
		PreviousHash: b.PreviousHash,
		Data:         b.Data,
		Checkpoint:   b.Checkpoint,
		Messages:     messages,
	}
}
