package notifiers

// Every notifier we create needs to adhere to this interface, so we can
// substitute another one when testing
type MessageNotifier interface {
	SendMessage(string, string, []byte)
}
