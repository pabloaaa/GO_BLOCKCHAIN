package src

// import (
// 	"crypto/sha256"
// 	"strconv"
// 	"strings"

// 	pb "github.com/pabloaaa/GO_BLOCKCHAIN/protos"
// 	types "github.com/pabloaaa/GO_BLOCKCHAIN/types"
// )

// func (b *types.Block) CalculateHash() []byte {
// 	var transactionsStrings []string
// 	for _, t := range b.Transactions {
// 		transactionsStrings = append(transactionsStrings, string(t.Sender)+string(t.Receiver)+strconv.FormatFloat(t.Amount, 'f', -1, 32))
// 	}
// 	data := strconv.FormatUint((b.Index), 10) + strconv.FormatUint(b.Timestamp, 10) + strings.Join(transactionsStrings, "") + string(b.PreviousHash) + strconv.FormatUint(b.Data, 10)
// 	hash := sha256.Sum256([]byte(data))
// 	return hash[:]
// }

// func (b *types.Block) SetData(nonce uint64) {
// 	b.Data = nonce
// }

// func BlockFromProto(pbBlock *pb.Block) *types.Block {
// 	transactions := make([]types.Transaction, len(pbBlock.GetTransactions()))
// 	for i, pbTransaction := range pbBlock.GetTransactions() {
// 		transactions[i] = types.Transaction{
// 			Sender:   pbTransaction.GetSender(),
// 			Receiver: pbTransaction.GetReceiver(),
// 			Amount:   pbTransaction.GetAmount(),
// 		}
// 	}

// 	return &types.Block{
// 		Index:        pbBlock.GetIndex(),
// 		Timestamp:    pbBlock.GetTimestamp(),
// 		PreviousHash: pbBlock.GetPreviousHash(),
// 		Transactions: transactions,
// 		Data:         pbBlock.GetData(),
// 		Checkpoint:   pbBlock.GetCheckpoint(),
// 	}
// }

// func (b *types.Block) ToProto() *pb.Block {
// 	pbTransactions := make([]*pb.Transaction, len(b.Transactions))
// 	for i, transaction := range b.Transactions {
// 		pbTransactions[i] = &pb.Transaction{
// 			Sender:   transaction.Sender,
// 			Receiver: transaction.Receiver,
// 			Amount:   transaction.Amount,
// 		}
// 	}

// 	return &pb.Block{
// 		Index:        b.Index,
// 		Timestamp:    b.Timestamp,
// 		PreviousHash: b.PreviousHash,
// 		Transactions: pbTransactions,
// 		Data:         b.Data,
// 		Checkpoint:   b.Checkpoint,
// 	}
// }
