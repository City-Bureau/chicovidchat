package directory

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"

	"github.com/City-Bureau/chicovidchat/pkg/chat"
)

func TestGetOrCreateConversationFromMessage(t *testing.T) {
	db, dbMock, _ := sqlmock.New()
	message := chat.Message{Sender: "test", Recipient: "test"}
	gormDB, _ := gorm.Open("postgres", db)

	dbMock.ExpectQuery("SELECT (.+) FROM (.+) WHERE (.+) LIMIT 1").
		WithArgs("+1234567890").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	_, created := GetOrCreateConversationFromMessage("+1234567890", message, gormDB)
	if created {
		t.Errorf("Created record instead of pulling latest")
	}
}
