package svc

import (
	"time"

	"github.com/City-Bureau/chicovidchat/pkg/chat"
	"github.com/sfreiberg/gotwilio"
)

// TwilioClient generalizes access to Twilio
type TwilioClient interface {
	SendSMS(string, string, string, string, string) (*gotwilio.SmsResponse, *gotwilio.Exception, error)
}

// TwilioChat implements the chat.Provider interface for Twilio messaging
type TwilioChat struct {
	Client         TwilioClient
	From           string // The Twilio automated number
	To             string // The user sending SMS
	StatusCallback string
	ApplicationSid string
	Channel        string // TODO: Should be SMS, MMS, WhatsApp?
}

// NewTwilioChat is a constructor for Twilio Chat structs
func NewTwilioChat(client TwilioClient, from, to, statusCallback, applicationSid string) *TwilioChat {
	return &TwilioChat{
		Client:         client,
		From:           from,
		To:             to,
		StatusCallback: statusCallback,
		ApplicationSid: applicationSid,
		Channel:        "",
	}
}

func (c *TwilioChat) SendSMS(body string) (*gotwilio.SmsResponse, *gotwilio.Exception, error) {
	return c.Client.SendSMS(c.From, c.To, body, c.StatusCallback, c.ApplicationSid)
}

// HandleSMSWebhook manages incoming SMS webhooks and converts them to chat.Message structs
func (c *TwilioChat) HandleSMSWebhook(data gotwilio.SMSWebhook) (chat.Message, error) {
	createdAt := time.Now()
	message := chat.Message{
		ID:        data.MessageSid,
		Sender:    data.From,
		Recipient: data.To,
		Body:      data.Body,
		CreatedAt: &createdAt,
	}
	return message, nil
}
