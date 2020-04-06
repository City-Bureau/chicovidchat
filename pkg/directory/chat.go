package directory

import (
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"

	"github.com/jinzhu/gorm"
	"github.com/jinzhu/gorm/dialects/postgres"

	"github.com/City-Bureau/chicovidchat/pkg/chat"
)

type chatState string

const (
	started     chatState = "started"
	setLanguage chatState = "set_language"
	setWhat     chatState = "set_what"
	setWho      chatState = "set_who"
	setZIP      chatState = "set_zip"
	results     chatState = "results"
)

const pageSize int = 3

// DirectoryChat manages chat conversations for directory filtering
type DirectoryChat struct {
	chat.Chat
	State  chatState     `json:"state"`
	Params *FilterParams `json:"params"`
	Page   int           `json:"page"`
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
		State:  started,
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

func languageOptions() []string {
	return []string{"en", "es"}
}

// Values should be IDs for i18n messages
func whatOptions() []string {
	return []string{"All", "Money", "Food", "Employment"}
}

func whoOptions() []string {
	return []string{"All", "LGBTQI", "Artists"}
}

// HandleMessage updates chat state based on message
func (c *DirectoryChat) HandleMessage(message chat.Message) ([]chat.Message, error) {
	var replies []chat.Message
	var bodies []string
	var err error
	switch c.State {
	case started:
		bodies, err = c.handleStarted(message.Body)
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

func (c *DirectoryChat) handleStarted(body string) ([]string, error) {
	c.State = setLanguage
	return c.buildLanguageMessage(), nil
}

func (c *DirectoryChat) buildLanguageMessage() []string {
	bodyStr := "Reply with the number of the language\n"
	for idx, val := range languageOptions() {
		bodyStr += fmt.Sprintf("\n%d: %s", idx, val)
	}
	return []string{bodyStr}
}

func (c *DirectoryChat) handleSetLanguage(body string) ([]string, error) {
	for idx, val := range languageOptions() {
		if strings.Contains(body, strconv.Itoa(idx)) {
			c.Language = val
		}
	}
	if c.Language == "" {
		return []string{"Please enter one of the options"}, nil
	}
	c.State = setWhat
	return c.buildWhatMessage(), nil
}

func (c *DirectoryChat) buildWhatMessage() []string {
	bodyStr := "Reply with all numbers for resources\n"
	for idx, val := range whatOptions() {
		bodyStr += fmt.Sprintf("\n%d: %s", idx, val)
	}
	return []string{bodyStr}
}

func (c *DirectoryChat) handleSetWhat(body string) ([]string, error) {
	hasMatch := false
	for idx, val := range whatOptions() {
		if strings.Contains(body, strconv.Itoa(idx)) {
			hasMatch = true
			// 0 is option for all
			if idx == 0 {
				c.State = setWho
				return c.buildWhoMessage(), nil
			}
			// TODO: Will need to be correct value for filtering
			c.Params.What = append(c.Params.What, val)
		}
	}
	if !hasMatch {
		return []string{"Please enter one of the options"}, nil
	}
	c.State = setWho
	return c.buildWhoMessage(), nil
}

func (c *DirectoryChat) buildWhoMessage() []string {
	bodyStr := "Reply with all numbers for groups\n"
	for idx, val := range whoOptions() {
		bodyStr += fmt.Sprintf("\n%d: %s", idx, val)
	}
	return []string{bodyStr}
}

func (c *DirectoryChat) handleSetWho(body string) ([]string, error) {
	hasMatch := false
	for idx, val := range whoOptions() {
		if strings.Contains(body, strconv.Itoa(idx)) {
			hasMatch = true
			// 0 is option for all
			if idx == 0 {
				c.State = setZIP
				return c.buildZIPMessage(), nil
			}
			// TODO: Will need to be correct value for filtering
			c.Params.Who = append(c.Params.Who, val)
		}
	}
	if !hasMatch {
		return []string{"Please enter one of the options"}, nil
	}
	c.State = setZIP
	return c.buildZIPMessage(), nil
}

func (c *DirectoryChat) buildZIPMessage() []string {
	return []string{"Enter your ZIP code"}
}

func (c *DirectoryChat) handleSetZIP(body string) ([]string, error) {
	cleanZIPRe := regexp.MustCompile(`\D`)
	cleanZIPStr := cleanZIPRe.ReplaceAllString(body, ``)
	zipRe := regexp.MustCompile(`\d{5}`)
	zipStr := zipRe.FindString(cleanZIPStr)
	if zipStr == "" {
		return []string{"Please enter a valid ZIP code"}, nil
	}
	c.Params.ZIP = &zipStr
	c.State = results
	return c.handleResults("")
}

func (c *DirectoryChat) handleResults(body string) ([]string, error) {
	var results []Resource

	// If page is not 0 and "1" not in string, ignore
	if !strings.Contains(body, "1") && c.Page != 0 {
		return []string{}, nil
	}

	zipMap := ZIPCodeMap()
	resources, err := LoadResources()
	if err != nil {
		return []string{}, err
	}

	filterJSON, _ := json.Marshal(c.Params)
	log.Println(string(filterJSON))

	for _, resource := range resources {
		if c.Params.MatchesFilters(resource, &zipMap) {
			results = append(results, resource)
		}
	}
	if len(results) == 0 {
		// Increment page so that it won't continue to send on replies
		c.Page++
		return []string{"No results available"}, nil
	}

	bodyStr := ""
	sendResults, hasRemaining := paginateResults(results, c.Page)
	if c.Page == 0 {
		bodyStr += fmt.Sprintf("%d results available\n", len(results))
	}
	for _, result := range sendResults {
		bodyStr += fmt.Sprintf("\n%s", result.Name)
	}
	if hasRemaining {
		bodyStr += "\nReply with 1 to see more"
	}
	c.Page++

	return []string{bodyStr}, nil
}

func paginateResults(resources []Resource, page int) ([]Resource, bool) {
	startIdx := page * pageSize
	endIdx := startIdx + pageSize
	if endIdx >= len(resources) {
		return resources[startIdx:], false
	} else {
		return resources[startIdx:endIdx], true
	}
}
