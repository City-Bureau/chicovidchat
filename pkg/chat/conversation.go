package chat

import (
	"time"

	"github.com/jinzhu/gorm"
	"github.com/jinzhu/gorm/dialects/postgres"
)

// Conversation is the struct for managing database access to Chats
type Conversation struct {
	gorm.Model
	Active bool           `gorm:"default:true" json:"active"`
	Data   postgres.Jsonb `json:"data"`
}

func CleanupInactiveConversations(db *gorm.DB) {
	// Mark any conversations as inactive that haven't been updated in 6 hours
	sixHoursAgo := time.Now().Add(time.Hour * -6)
	db.Model(&Conversation{}).Where("active = ? AND updated_at < ?", true, sixHoursAgo).Update("active", false)
}
