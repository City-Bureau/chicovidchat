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
	"github.com/nicksnyder/go-i18n/v2/i18n"

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
	State     chatState     `json:"state"`
	Params    *FilterParams `json:"params"`
	Page      int           `json:"page"`
	localizer *i18n.Localizer
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

	if c.localizer == nil {
		c.localizer = LoadLocalizer(c.Language)
	}

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
	bodyStr := c.localizer.MustLocalize(&i18n.LocalizeConfig{
		MessageID: "language-prompt",
	})
	bodyStr += "\n"
	for idx, val := range languageOptions() {
		langStr := c.localizer.MustLocalize(&i18n.LocalizeConfig{
			MessageID:    fmt.Sprintf("option-%s", val),
			TemplateData: map[string]string{"Number": strconv.Itoa(idx)},
		})
		bodyStr += "\n"
		bodyStr += langStr
	}
	return []string{bodyStr}
}

func (c *DirectoryChat) handleSetLanguage(body string) ([]string, error) {
	for idx, val := range languageOptions() {
		if strings.Contains(body, strconv.Itoa(idx)) {
			c.Language = val
			c.localizer = LoadLocalizer(val)
		}
	}
	if c.Language == "" {
		// TODO: currently just redoing?
		return c.buildLanguageMessage(), nil
	}
	c.State = setWhat
	return c.buildWhatMessage(), nil
}

func (c *DirectoryChat) buildWhatMessage() []string {
	bodyStr := c.localizer.MustLocalize(&i18n.LocalizeConfig{
		MessageID: "what-prompt",
	})
	bodyStr += "\n"
	for idx, val := range whatOptions() {
		whatOption := c.localizer.MustLocalize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{ID: val, Other: val},
		})
		// TODO: Maybe add "Text X for"...
		bodyStr += fmt.Sprintf("\n%d ", idx)
		bodyStr += whatOption
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
		invalidPrompt := c.localizer.MustLocalize(&i18n.LocalizeConfig{
			MessageID: "please-enter-valid-option",
		})
		return []string{invalidPrompt}, nil
	}
	c.State = setWho
	return c.buildWhoMessage(), nil
}

func (c *DirectoryChat) buildWhoMessage() []string {
	bodyStr := c.localizer.MustLocalize(&i18n.LocalizeConfig{
		MessageID: "who-prompt",
	})
	bodyStr += "\n"
	for idx, val := range whoOptions() {
		whatOption := c.localizer.MustLocalize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{ID: val, Other: val},
		})
		bodyStr += fmt.Sprintf("\n%d ", idx)
		bodyStr += whatOption
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
		invalidPrompt := c.localizer.MustLocalize(&i18n.LocalizeConfig{
			MessageID: "please-enter-valid-option",
		})
		return []string{invalidPrompt}, nil
	}
	c.State = setZIP
	return c.buildZIPMessage(), nil
}

func (c *DirectoryChat) buildZIPMessage() []string {
	zipPrompt := c.localizer.MustLocalize(&i18n.LocalizeConfig{
		MessageID: "zip-prompt",
	})
	return []string{zipPrompt}
}

func (c *DirectoryChat) handleSetZIP(body string) ([]string, error) {
	cleanZIPRe := regexp.MustCompile(`\D`)
	cleanZIPStr := cleanZIPRe.ReplaceAllString(body, ``)
	zipRe := regexp.MustCompile(`\d{5}`)
	zipStr := zipRe.FindString(cleanZIPStr)
	if zipStr == "" {
		invalidPrompt := c.localizer.MustLocalize(&i18n.LocalizeConfig{
			MessageID: "please-enter-valid-zip",
		})
		return []string{invalidPrompt}, nil
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
		noResults := c.localizer.MustLocalize(&i18n.LocalizeConfig{
			MessageID: "no-results",
		})
		return []string{noResults}, nil
	}

	bodyStr := ""
	// TODO: Figure out handling when someone is past limit
	sendResults, hasRemaining := paginateResults(results, c.Page)
	if c.Page == 0 {
		resultsStr := c.localizer.MustLocalize(&i18n.LocalizeConfig{
			MessageID:   "results-available",
			PluralCount: len(results),
		})
		bodyStr += fmt.Sprintf("%s\n", resultsStr)
	}
	for _, result := range sendResults {
		bodyStr += fmt.Sprintf("\n%s", result.Name)
	}
	if hasRemaining {
		seeMoreStr := c.localizer.MustLocalize(&i18n.LocalizeConfig{
			MessageID: "see-more-prompt",
		})
		bodyStr += fmt.Sprintf("\n\n%s", seeMoreStr)
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
