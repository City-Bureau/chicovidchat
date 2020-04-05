package chat

import "time"

// Message is a single message exchanged in a Chat
type Message struct {
	ID        string     `json:"id"`
	Sender    string     `json:"sender"`
	Recipient string     `json:"recipient"`
	Body      string     `json:"body"`
	CreatedAt *time.Time `json:"created_at"`
}
