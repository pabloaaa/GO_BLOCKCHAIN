package interfaces

type MessageSender interface {
	SendMsg(data []byte) error
}
