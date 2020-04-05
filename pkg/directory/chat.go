package directory

import (
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

func languageOptions() map[string]string {
	return map[string]string{
		"1": "en",
		"2": "es",
	}
}

// Values should be IDs for i18n messages
func whatOptions() map[string]string {
	return map[string]string{
		"1": "Money",
		"2": "Food",
		"3": "Employment",
	}
}

func whoOptions() map[string]string {
	return map[string]string{
		"1": "Artists",
		"2": "LGBTQI",
	}
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

// HandleMessage updates chat state based on message
func (c *DirectoryChat) HandleMessage(message chat.Message) (chat.Message, error) {
	var reply chat.Message
	var body string
	var err error
	switch c.State {
	case setLanguage:
		body, err = c.handleSetLanguage(message.Body)
	case setWhat:
		body, err = c.handleSetWhat(message.Body)
	case setWho:
		body, err = c.handleSetWho(message.Body)
	case setZIP:
		body, err = c.handleSetZIP(message.Body)
	case results:
		body, err = c.handleResults(message.Body)
	}
	if body != "" {
		reply = chat.Message{
			Sender:    "",
			Recipient: message.Sender,
			Body:      body,
		}
		// reply = chat.NewMessage("", message.Sender, body, nil)
	}
	return reply, err
}

func (c *DirectoryChat) handleSetLanguage(body string) (string, error) {
	return "", nil
}

func (c *DirectoryChat) handleSetWhat(body string) (string, error) {
	return "", nil
}

func (c *DirectoryChat) handleSetWho(body string) (string, error) {
	return "", nil
}

func (c *DirectoryChat) handleSetZIP(body string) (string, error) {
	return "", nil
}

func (c *DirectoryChat) handleResults(body string) (string, error) {
	return "", nil
}
