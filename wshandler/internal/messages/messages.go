package messages

const (
	ENTER_CHAT   = "ENTER_CHAT"
	LEAVE_CHAT   = "LEAVE_CHAT"
	SEND_MESSAGE = "SEND_MESSAGE"
	BROADCAST    = "BROADCAST"
)

type Message struct {
	Body        string   `json:"body"`
	Attachments []string `json:"attachments"`
}

type Event struct {
	Type         string  `json:"type"`
	Msg          Message `json:"message"`
	From         string  `json:"from"`
	To           string  `json:"to"`
	Url          string  `json:"url"`
	ConnectionId string  `json:"connectionId"`
}

type DynamoEvent struct {
	ConnectionId string `json:"partitionKey"`
	To           string `json:"sortKey"`
}
