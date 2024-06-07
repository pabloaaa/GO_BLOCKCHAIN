package main

type BlockCreator struct {
}

func NewBlockCreator() *BlockCreator {
	return &BlockCreator{}
}

// func (bc *BlockCreator) Start(blockchain *Blockchain) {
// 	nonce := 0
// 	for {
// 		transactions := bc.generateTransactions()
// 		timestamp := uint64(time.Now().Unix())
// 		blockchain.mux.Lock()
// 		lastBlock := blockchain.Last()
// 		previousHash := lastBlock.Hash
// 		index := lastBlock.Index + 1
// 		blockchain.mux.Unlock()
// 		block := NewBlock(index, timestamp, transactions, previousHash, uint64(nonce))
// 		for {
// 			blockchain.mux.Lock()
// 			err := bc.validator.ValidateAndAddBlock(block, blockchain)
// 			blockchain.mux.Unlock()
// 			if err == nil {
// 				break
// 			}
// 			nonce++
// 			block.Data = uint64(nonce)
// 			block.calculateHash()
// 		}
// 		time.Sleep(time.Second)
// 	}
// }

// func (bc *BlockCreator) generateTransactions() []Transaction {
// 	var transactions []Transaction
// 	rng := rand.Reader
// 	max := big.NewInt(1e6)
// 	for i := 0; i < 10; i++ {
// 		sender, _ := rand.Int(rng, big.NewInt(1000))
// 		receiver, _ := rand.Int(rng, big.NewInt(1000))
// 		amountBig, _ := rand.Int(rng, max)
// 		amount := float64(amountBig.Int64()) / float64(max.Int64())
// 		transactions = append(transactions, Transaction{Sender: sender.String(), Receiver: receiver.String(), Amount: amount})
// 	}
// 	return transactions
// }
