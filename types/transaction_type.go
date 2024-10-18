package types

type Transaction struct {
	Sender   []byte
	Receiver []byte
	Amount   float64
}
