package notifiers

// MessageNotifier defines our interface that all our notifications need to adhere too, also handy to swap out in test
type MessageNotifier interface {
	SendMessage(string, string, []byte)
}
