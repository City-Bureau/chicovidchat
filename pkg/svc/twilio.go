package svc

import (
	"crypto/hmac"
	"net/url"
	"time"

	"github.com/City-Bureau/chicovidchat/pkg/chat"
	"github.com/sfreiberg/gotwilio"
)

// TwilioClient generalizes access to Twilio
type TwilioClient interface {
	SendSMS(string, string, string, string, string) (*gotwilio.SmsResponse, *gotwilio.Exception, error)
	GenerateSignature(string, url.Values) ([]byte, error)
}

// TwilioChat implements the chat.Provider interface for Twilio messaging
type TwilioChat struct {
	Client  TwilioClient
	From    string // The Twilio automated number
	To      string // The user sending SMS
	Channel string // TODO: Should be SMS, MMS, WhatsApp?
}

// NewTwilioChat is a constructor for Twilio Chat structs
func NewTwilioChat(client TwilioClient, from, to string) *TwilioChat {
	return &TwilioChat{
		Client:  client,
		From:    from,
		To:      to,
		Channel: "",
	}
}

func (c *TwilioChat) SendSMS(body string) (*gotwilio.SmsResponse, *gotwilio.Exception, error) {
	return c.Client.SendSMS(c.From, c.To, body, "", "")
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

func (c *TwilioChat) CheckSignature(url, signature string, values url.Values) (bool, error) {
	expected, err := c.Client.GenerateSignature(url, values)
	if err != nil {
		return false, err
	}

	return hmac.Equal(expected, []byte(signature)), nil
}
