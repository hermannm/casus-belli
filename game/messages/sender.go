package messages

type Sender interface {
	SendMessage(to string, msg any) error
	SendMessageToAll(msg any) error
}
