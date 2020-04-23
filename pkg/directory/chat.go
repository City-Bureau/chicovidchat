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

// Used to force Unicode in SMS so that ñ renders consistently
const punctuationSpace string = " "
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
			Category:  "directory",
			Language:  "en",
		},
		State:  started,
		Params: &FilterParams{},
		Page:   0,
	}
}

func GetOrCreateConversationFromMessage(contact string, message chat.Message, db *gorm.DB) (*chat.Conversation, bool) {
	var conversation chat.Conversation
	if db.Model(&chat.Conversation{}).Where("data ->> 'id' = ? AND active IS TRUE", contact).Last(&conversation).RecordNotFound() {
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
	return []string{"en", "es", "zh", "pl", "ur", "tl", "vi"}
}

// Values should be IDs for i18n messages
func whatOptions() []string {
	return []string{"All", "Money", "Food", "Housing", "Health", "Mental Health", "Utilities", "Legal Help"}
}

func whoOptions() []string {
	return []string{"All", "Families", "Immigrants", "LGBTQI", "Business Owners", "Students", "None"}
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
	bodyStr := fmt.Sprintf(
		"%s\n%s\n\n%s%s\n",
		c.localizer.MustLocalize(&i18n.LocalizeConfig{
			MessageID: "site-title",
		}), c.localizer.MustLocalize(&i18n.LocalizeConfig{
			MessageID: "site-intro",
		}), c.localizer.MustLocalize(&i18n.LocalizeConfig{
			MessageID: "language-prompt",
		}),
		punctuationSpace,
	)
	for idx, val := range languageOptions() {
		bodyStr += fmt.Sprintf("\n%s", c.localizer.MustLocalize(&i18n.LocalizeConfig{
			MessageID:    fmt.Sprintf("option-%s", val),
			TemplateData: map[string]string{"Number": strconv.Itoa(idx)},
		}))
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
		return c.buildLanguageMessage(), nil
	}
	c.State = setWhat
	return c.buildWhatMessage(), nil
}

func (c *DirectoryChat) buildWhatMessage() []string {
	bodyStr := fmt.Sprintf(
		"%s\n%s%s\n",
		c.localizer.MustLocalize(&i18n.LocalizeConfig{
			MessageID: "what-prompt",
		}),
		c.localizer.MustLocalize(&i18n.LocalizeConfig{
			MessageID: "enter-all-numbers",
		}),
		c.unicodeIfNeeded(),
	)
	for idx, val := range whatOptions() {
		bodyStr += fmt.Sprintf("\n%s", c.localizer.MustLocalize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    fmt.Sprintf("option-%s", val),
				Other: fmt.Sprintf("%d %s", idx, val),
			},
			TemplateData: map[string]string{"Number": strconv.Itoa(idx)},
		}))
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
	bodyStr := fmt.Sprintf(
		"%s\n%s%s\n",
		c.localizer.MustLocalize(&i18n.LocalizeConfig{
			MessageID: "who-prompt",
		}),
		c.localizer.MustLocalize(&i18n.LocalizeConfig{
			MessageID: "enter-all-numbers",
		}),
		c.unicodeIfNeeded(),
	)
	for idx, val := range whoOptions() {
		optionKey := "option"
		// Override display of "All" translation
		if idx == 0 {
			optionKey = "who-option"
		}
		bodyStr += fmt.Sprintf("\n%s", c.localizer.MustLocalize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    fmt.Sprintf("%s-%s", optionKey, val),
				Other: fmt.Sprintf("%d %s", idx, val),
			},
			TemplateData: map[string]string{"Number": strconv.Itoa(idx)},
		}))
	}
	return []string{bodyStr}
}

func (c *DirectoryChat) handleSetWho(body string) ([]string, error) {
	hasMatch := false
	whoOpts := whoOptions()
	for idx, val := range whoOpts {
		if strings.Contains(body, strconv.Itoa(idx)) {
			hasMatch = true
			// 0 is option for all
			if idx == 0 {
				c.State = setZIP
				return c.buildZIPMessage(), nil
			} else if idx == len(whoOpts)-1 {
				// Last item is option for none
				c.State = setZIP
				c.Params.Who = []string{"None"}
				return c.buildZIPMessage(), nil
			}
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
	zipPrompt += c.unicodeIfNeeded()
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

	if strings.Contains(body, "2") {
		return c.handleRestart()
	} else if !strings.Contains(body, "1") && c.Page != 0 {
		// If page is not 0 and "1" not in string, ignore
		return []string{}, nil
	}

	zipMap := ZIPCodeMap()
	chiZips := ChiZIPCodes()
	resources, err := LoadResources()
	if err != nil {
		return []string{}, err
	}

	filterJSON, _ := json.Marshal(c.Params)
	log.Println(string(filterJSON))

	for _, resource := range resources {
		if c.Params.MatchesFilters(resource, &zipMap, &chiZips) {
			results = append(results, resource)
		}
	}

	restartPrompt := c.localizer.MustLocalize(&i18n.LocalizeConfig{
		MessageID:    "restart-prompt",
		TemplateData: map[string]string{"Number": "2"},
	})
	restartPrompt += c.unicodeIfNeeded()
	if len(results) == 0 {
		// Increment page so that it won't continue to send on replies
		c.Page++
		replyStr := fmt.Sprintf("\n%s%s", c.localizer.MustLocalize(&i18n.LocalizeConfig{
			MessageID: "no-results",
		}), c.unicodeIfNeeded())
		return []string{replyStr}, nil
	}

	bodyStr := ""

	sendResults, hasRemaining := PaginateResults(results, c.Page)
	// Skip if past pagination limits
	if len(sendResults) == 0 {
		return []string{}, nil
	}

	// Include results header if first page of results
	if c.Page == 0 {
		bodyStr += c.localizer.MustLocalize(&i18n.LocalizeConfig{
			MessageID:   "results-available",
			PluralCount: len(results),
		})
	}

	// Add result text to the message body
	for _, result := range sendResults {
		bodyStr += fmt.Sprintf("\n\n\n%s", result.AsText(c.Language, c.localizer))
	}

	// Show a prompt for paginating if more results available
	if hasRemaining {
		seeMoreStr := c.localizer.MustLocalize(&i18n.LocalizeConfig{
			MessageID:    "see-more-prompt",
			TemplateData: map[string]string{"Number": "1"},
		})
		bodyStr += fmt.Sprintf("\n\n%s\n", seeMoreStr)
	} else {
		// Add padding for restart prompt if see more prompt not included
		bodyStr += "\n\n"
	}
	bodyStr += restartPrompt
	c.Page++

	return []string{bodyStr}, nil
}

// Reset filters, page, go back to setting "what", keep language
func (c *DirectoryChat) handleRestart() ([]string, error) {
	c.Params = &FilterParams{}
	c.State = setWhat
	c.Page = 0
	return c.buildWhatMessage(), nil
}

// Add a unicode punctuation space to ensure non-ASCII characters load if language isn't English
func (c *DirectoryChat) unicodeIfNeeded() string {
	if c.Language != "" && c.Language != "en" {
		return punctuationSpace
	}
	return ""
}

func PaginateResults(resources []Resource, page int) ([]Resource, bool) {
	startIdx := page * pageSize
	endIdx := startIdx + pageSize
	if startIdx >= len(resources) {
		return []Resource{}, false
	} else if endIdx >= len(resources) {
		return resources[startIdx:], false
	} else {
		return resources[startIdx:endIdx], true
	}
}
