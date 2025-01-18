package types

import (
	"crypto/sha256"
	"strconv"
	"strings"

	pb "github.com/pabloaaa/GO_BLOCKCHAIN/protos"
)

// Block represents a single block in the blockchain.
type Block struct {
	Index        uint64
	Timestamp    uint64
	PreviousHash []byte
	Transactions []Transaction
	Data         uint64
	Checkpoint   bool
}

// BlockNode represents a node in the blockchain tree.
type BlockNode struct {
	Block  *Block
	Parent *BlockNode
	Childs []*BlockNode
}

// Transaction represents a transaction in the blockchain.
type Transaction struct {
	Sender   []byte
	Receiver []byte
	Amount   float64
}

// CalculateHash calculates the SHA-256 hash of the block.
func (b *Block) CalculateHash() []byte {
	var transactionsStrings []string
	for _, t := range b.Transactions {
		transactionsStrings = append(transactionsStrings, string(t.Sender)+string(t.Receiver)+strconv.FormatFloat(t.Amount, 'f', -1, 32))
	}
	data := strconv.FormatUint((b.Index), 10) + strconv.FormatUint(b.Timestamp, 10) + strings.Join(transactionsStrings, "") + string(b.PreviousHash) + strconv.FormatUint(b.Data, 10)
	hash := sha256.Sum256([]byte(data))
	return hash[:]
}

// BlockFromProto converts a protobuf Block to a Block.
func BlockFromProto(pbBlock *pb.Block) *Block {
	transactions := make([]Transaction, len(pbBlock.GetTransactions()))
	for i, pbTransaction := range pbBlock.GetTransactions() {
		transactions[i] = Transaction{
			Sender:   pbTransaction.GetSender(),
			Receiver: pbTransaction.GetReceiver(),
			Amount:   pbTransaction.GetAmount(),
		}
	}

	return &Block{
		Index:        pbBlock.GetIndex(),
		Timestamp:    pbBlock.GetTimestamp(),
		PreviousHash: pbBlock.GetPreviousHash(),
		Transactions: transactions,
		Data:         pbBlock.GetData(),
		Checkpoint:   pbBlock.GetCheckpoint(),
	}
}

// ToProto converts a Block to a protobuf Block.
func (b *Block) ToProto() *pb.Block {
	pbTransactions := make([]*pb.Transaction, len(b.Transactions))
	for i, transaction := range b.Transactions {
		pbTransactions[i] = &pb.Transaction{
			Sender:   transaction.Sender,
			Receiver: transaction.Receiver,
			Amount:   transaction.Amount,
		}
	}

	return &pb.Block{
		Index:        b.Index,
		Timestamp:    b.Timestamp,
		PreviousHash: b.PreviousHash,
		Transactions: pbTransactions,
		Data:         b.Data,
		Checkpoint:   b.Checkpoint,
	}
}
