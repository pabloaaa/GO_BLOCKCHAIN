package interfaces

type MessageSender interface {
	SendMsgToAddress(address string, data []byte) error
	SendMsg(data []byte) error
}
