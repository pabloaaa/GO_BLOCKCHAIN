package interfaces

type MessageSender interface {
	SendMsgToAddress(address []byte, data []byte) error
	SendMsg(data []byte) error
}
