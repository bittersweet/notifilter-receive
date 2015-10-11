package notifiers

type NotifierResponse struct {
	response *slackResponse
	error    error
}

// Every notifier we create needs to adhere to this interface, so we can
// substitute another one when testing
type MessageNotifier interface {
	SendMessage(string, []byte) NotifierResponse
}
