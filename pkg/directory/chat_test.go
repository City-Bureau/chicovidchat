package directory

import (
	"os"
	"path"
	"runtime"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"

	"github.com/City-Bureau/chicovidchat/pkg/chat"
)

// https://brandur.org/fragments/testing-go-project-root
// Updating to use relative paths for loading i18n bundles
func init() {
	_, filename, _, _ := runtime.Caller(0)
	dir := path.Join(path.Dir(filename), "..", "..")
	err := os.Chdir(dir)
	if err != nil {
		panic(err)
	}
}

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

func TestHandleMessageSetMultiple(t *testing.T) {
	dirChat := NewDirectoryChat("test")
	dirChat.State = setWhat
	_, _ = dirChat.HandleMessage(chat.Message{Body: "1 2"})
	if dirChat.State != setWho {
		t.Errorf("State not updated with valid response")
	}
	if len(dirChat.Params.What) != 2 {
		t.Errorf("Multiple numbers not setting multiple options")
	}
	dirChat.State = setWhat
	dirChat.Params.What = []string{}
	_, _ = dirChat.HandleMessage(chat.Message{Body: "123"})
	if len(dirChat.Params.What) != 3 {
		t.Errorf("Multiple numbers without spaces not setting options")
	}
}

func TestHandleMessageInvalid(t *testing.T) {
	dirChat := NewDirectoryChat("test")
	dirChat.State = setWhat
	_, _ = dirChat.HandleMessage(chat.Message{Body: "test"})
	if dirChat.State != setWhat || len(dirChat.Params.What) != 0 {
		t.Errorf("Invalid options accepted")
	}
	_, _ = dirChat.HandleMessage(chat.Message{Body: "9"})
	if dirChat.State != setWhat || len(dirChat.Params.What) != 0 {
		t.Errorf("Invalid option was accepted")
	}
}

func TestHandleSetZIP(t *testing.T) {
	dirChat := NewDirectoryChat("test")
	dirChat.localizer = LoadLocalizer("en")
	dirChat.State = setZIP
	_, _ = dirChat.handleSetZIP(" 12345 ")
	if dirChat.State != results && *dirChat.Params.ZIP != "12345" {
		t.Errorf("ZIP not being set from text with extra spaces")
	}

	dirChat.State = setZIP
	dirChat.Params.ZIP = nil
	_, _ = dirChat.handleSetZIP(" 123 45")
	if dirChat.State != results && *dirChat.Params.ZIP != "12345" {
		t.Errorf("ZIP with extra characters not being accepted")
	}

	dirChat.State = setZIP
	dirChat.Params.ZIP = nil
	_, _ = dirChat.handleSetZIP("1234")
	if dirChat.State != setZIP && dirChat.Params.ZIP != nil {
		t.Errorf("Invalid ZIP value accepted")
	}

	dirChat.State = setZIP
	dirChat.Params.ZIP = nil
	_, _ = dirChat.handleSetZIP("12345-67890")
	if dirChat.State != results && *dirChat.Params.ZIP != "12345" {
		t.Errorf("ZIP with extra digits not being set correctly")
	}
}

func TestPaginateResults(t *testing.T) {
	resources := []Resource{Resource{}, Resource{}, Resource{}, Resource{}}
	results, hasMore := PaginateResults(resources, 0)
	if len(results) != 3 || !hasMore {
		t.Errorf("First page not pulling correct results")
	}
	results, hasMore = PaginateResults(resources, 1)
	if len(results) != 1 || hasMore {
		t.Errorf("Second page not pulling correct results")
	}
	results, hasMore = PaginateResults(resources, 2)
	if len(results) != 0 || hasMore {
		t.Errorf("Third page not pulling correct results")
	}
}
