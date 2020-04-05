package chat

import (
	"time"

	"github.com/jinzhu/gorm/dialects/postgres"
)

// Conversation is the struct for managing database access to Chats
type Conversation struct {
	PK        uint      `gorm:primary_key"`
	CreatedAt time.Time `json:"createdAt" gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP"`
	UpdatedAt time.Time `json:"updatedAt" gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP ON UPDATE CURRENT TIMESTAMP"`
	Data      postgres.Jsonb
}

// Chat is the main struct for managing conversations
type Chat struct {
	ContactID string    `json:"id"`
	Active    bool      `json:"active"`
	Category  string    `json:"category"`
	Language  string    `json:"language"`
	Messages  []Message `json:"messages"`
}
