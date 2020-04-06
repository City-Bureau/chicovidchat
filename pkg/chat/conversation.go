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
	// Mark any conversations as inactive that haven't been updated in one week
	lastWeek := time.Now().Add(time.Hour * -(24 * 7))
	db.Model(&Conversation{}).Where("active = ? AND updated_at < ?", true, lastWeek).Update("active", false)
}
