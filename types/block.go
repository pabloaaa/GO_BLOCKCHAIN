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
	hash := sha256.Sum256([]byte(data))
	return hash[:]
}

// BlockFromProto converts a protobuf Block to a Block.
func BlockFromProto(pbBlock *pb.Block) *Block {
	return &Block{
		Index:        pbBlock.GetIndex(),
		Timestamp:    pbBlock.GetTimestamp(),
		PreviousHash: pbBlock.GetPreviousHash(),
		Data:         pbBlock.GetData(),
		Checkpoint:   pbBlock.GetCheckpoint(),
	}
}

// ToProto converts a Block to a protobuf Block.
func (b *Block) ToProto() *pb.Block {
	return &pb.Block{
		Index:        b.Index,
		Timestamp:    b.Timestamp,
		PreviousHash: b.PreviousHash,
		Data:         b.Data,
		Checkpoint:   b.Checkpoint,
	}
}
