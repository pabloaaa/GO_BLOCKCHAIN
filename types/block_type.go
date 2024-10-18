package types

type Block struct {
	Index        uint64
	Timestamp    uint64
	PreviousHash []byte
	Transactions []Transaction
	Data         uint64
	Checkpoint   bool
}
