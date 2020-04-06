package chat

// Chat is the main struct for managing conversations
type Chat struct {
	ContactID string    `json:"id"`
	Active    bool      `json:"active"`
	Category  string    `json:"category"`
	Language  string    `json:"language"`
	Messages  []Message `json:"messages"`
}
