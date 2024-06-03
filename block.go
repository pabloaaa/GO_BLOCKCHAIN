package main

import (
	"crypto/sha256"
	"encoding/hex"
	"strconv"
	"strings"

	pb "github.com/pabloaaa/GO_BLOCKCHAIN/protos"
)

type Transaction struct {
	Sender   string  `json:"sender"`
	Receiver string  `json:"receiver"`
	Amount   float64 `json:"amount"`
}

type Block struct {
	Index        uint64        `json:"index"`
	Timestamp    uint64        `json:"timestamp"`
	PreviousHash string        `json:"previous_hash"`
	Hash         string        `json:"hash"`
	Transactions []Transaction `json:"transactions"`
	Data         uint64        `json:"data"`
}

func NewBlock(index uint64, timestamp uint64, transactions []Transaction, previousHash string, nonce uint64) *Block {
	block := &Block{
		Index:        index,
		Timestamp:    timestamp,
		Transactions: transactions,
		PreviousHash: previousHash,
		Hash:         "",
		Data:         nonce,
	}
	block.calculateHash()
	return block
}

func (b *Block) calculateHash() {
	var transactionsStrings []string
	for _, t := range b.Transactions {
		transactionsStrings = append(transactionsStrings, t.Sender+t.Receiver+strconv.FormatFloat(t.Amount, 'f', -1, 32))
	}
	data := strconv.FormatUint((b.Index), 10) + strconv.FormatUint(b.Timestamp, 10) + strings.Join(transactionsStrings, "") + b.PreviousHash + strconv.FormatUint(b.Data, 10)
	hash := sha256.Sum256([]byte(data))
	b.Hash = hex.EncodeToString(hash[:])
}

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
		Hash:         pbBlock.GetHash(),
		Transactions: transactions,
		Data:         pbBlock.GetData(),
	}
}

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
		Hash:         b.Hash,
		Transactions: pbTransactions,
		Data:         b.Data,
	}
}
