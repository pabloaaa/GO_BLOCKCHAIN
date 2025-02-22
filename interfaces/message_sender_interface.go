package interfaces

type MessageSender interface {
	SendMsgToAddress(address int, data []byte, senderPort int) error
}
