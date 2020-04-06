package chat

import (
	"github.com/jinzhu/gorm"
	"github.com/jinzhu/gorm/dialects/postgres"
)

// Conversation is the struct for managing database access to Chats
type Conversation struct {
	gorm.Model
	Data postgres.Jsonb `json:"data"`
}
