package messages

type Handler struct {
	sender    Sender
	receivers map[string]Receiver
}

func NewHandler(sender Sender) Handler {
	receivers := make(map[string]Receiver)
	return Handler{
		sender:    sender,
		receivers: receivers,
	}
}
