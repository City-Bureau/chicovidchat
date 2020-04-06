package directory

import (
	"encoding/json"
	"strings"

	"github.com/jinzhu/gorm"
	"github.com/jinzhu/gorm/dialects/postgres"

	"github.com/City-Bureau/chicovidchat/pkg/chat"
)

type chatState string

const (
	setLanguage chatState = "started"
	setWhat     chatState = "set_what"
	setWho      chatState = "set_who"
	setZIP      chatState = "set_zip"
	results     chatState = "results"
)

// DirectoryChat manages chat conversations for directory filtering
type DirectoryChat struct {
	chat.Chat
	State  chatState     `json:"state"`
	Params *FilterParams `json:"params"`
	Page   uint          `json:"page"`
}

// NewDirectoryChat is a constructor for DirectoryChat structs
func NewDirectoryChat(id string) *DirectoryChat {
	return &DirectoryChat{
		Chat: chat.Chat{
			ContactID: id,
			Active:    true,
			Category:  "directory",
			Language:  "",
		},
		State:  setLanguage,
		Params: &FilterParams{},
		Page:   0,
	}
}

func GetOrCreateConversationFromMessage(contact string, message chat.Message, db *gorm.DB) (*chat.Conversation, bool) {
	var conversation chat.Conversation
	if db.Model(&chat.Conversation{}).Where("data ->> 'id' = ? AND (data ->> 'active')::boolean IS TRUE", contact).Last(&conversation).RecordNotFound() {
		directoryChat := NewDirectoryChat(message.Sender)
		directoryChat.Messages = []chat.Message{message}
		_ = UpdateDirectoryChatConversation(directoryChat, &conversation, db)
		return &conversation, true
	}
	return &conversation, false
}

func UpdateDirectoryChatConversation(directoryChat *DirectoryChat, conversation *chat.Conversation, db *gorm.DB) error {
	chatJSON, _ := json.Marshal(directoryChat)
	conversation.Data = postgres.Jsonb{
		RawMessage: json.RawMessage(chatJSON),
	}

	db.Save(conversation)
	return nil
}

func languageOptions() map[string]string {
	return map[string]string{
		"1": "en",
		"2": "es",
	}
}

// Values should be IDs for i18n messages
// nolint
func whatOptions() map[string]string {
	return map[string]string{
		"1": "Money",
		"2": "Food",
		"3": "Employment",
	}
}

// nolint
func whoOptions() map[string]string {
	return map[string]string{
		"1": "Artists",
		"2": "LGBTQI",
	}
}

// HandleMessage updates chat state based on message
func (c *DirectoryChat) HandleMessage(message chat.Message) ([]chat.Message, error) {
	var replies []chat.Message
	var bodies []string
	var err error
	switch c.State {
	case setLanguage:
		bodies, err = c.handleSetLanguage(message.Body)
	case setWhat:
		bodies, err = c.handleSetWhat(message.Body)
	case setWho:
		bodies, err = c.handleSetWho(message.Body)
	case setZIP:
		bodies, err = c.handleSetZIP(message.Body)
	case results:
		bodies, err = c.handleResults(message.Body)
	}
	if len(bodies) > 0 {
		for _, body := range bodies {
			replies = append(replies, chat.Message{
				Sender:    "",
				Recipient: message.Sender,
				Body:      body,
			})
		}
	}
	return replies, err
}

func (c *DirectoryChat) handleSetLanguage(body string) ([]string, error) {
	langOptions := languageOptions()
	for k, v := range langOptions {
		if strings.Contains(body, k) {
			c.Language = v
		}
	}
	c.State = setWhat
	return []string{"Message for what"}, nil
}

func (c *DirectoryChat) handleSetWhat(body string) ([]string, error) {
	return []string{""}, nil
}

func (c *DirectoryChat) handleSetWho(body string) ([]string, error) {
	return []string{""}, nil
}

func (c *DirectoryChat) handleSetZIP(body string) ([]string, error) {
	return []string{""}, nil
}

func (c *DirectoryChat) handleResults(body string) ([]string, error) {
	var results []Resource
	var replies []string
	zipMap := ZIPCodeMap()
	resources, err := LoadResources()
	if err != nil {
		return replies, err
	}
	for _, resource := range resources {
		if c.Params.MatchesFilters(resource, &zipMap) {
			results = append(results, resource)
		}
	}
	if len(results) == 0 {
		return []string{"No results available"}, nil
	}

	// var resources []Resource
	// for _, resource := range
	// if c.Page == 0 {

	// }
	return []string{""}, nil
}
