package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/sfreiberg/gotwilio"

	"github.com/City-Bureau/chicovidchat/pkg/chat"
	"github.com/City-Bureau/chicovidchat/pkg/svc"
)

func SendMessage(message chat.Message, twilioChat *svc.TwilioChat, snsClient *svc.SNSClient) error {
	twilioRes, twilioErr, sendErr := twilioChat.SendSMS(message.Body)
	if sendErr != nil {
		return sendErr
	}
	if twilioErr != nil {
		return fmt.Errorf("Twilio returned error code: %s", *twilioErr)
	}

	createdAt := time.Now()
	twilioMessage := chat.Message{
		ID:        twilioRes.Sid,
		Sender:    os.Getenv("TWILIO_FROM"),
		Recipient: message.Recipient,
		Body:      message.Body,
		CreatedAt: &createdAt,
	}
	twilioMessageJSON, _ := json.Marshal(twilioMessage)
	return snsClient.Publish(string(twilioMessageJSON), os.Getenv("SNS_TOPIC_ARN"), svc.SentMessageFeed)
}

func handler(request events.SNSEvent) error {
	if len(request.Records) <= 0 {
		return nil
	}
	message := request.Records[0].SNS.Message

	var messages []chat.Message
	err := json.Unmarshal([]byte(message), &messages)
	if err != nil {
		return err
	}

	if len(messages) == 0 {
		return nil
	}

	client := gotwilio.NewTwilioClient(
		os.Getenv("TWILIO_ACCOUNT_SID"),
		os.Getenv("TWILIO_AUTH_TOKEN"),
	)
	twilioChat := svc.NewTwilioChat(
		client,
		os.Getenv("TWILIO_FROM"),
		messages[0].Recipient,
	)
	snsClient := svc.NewSNSClient()

	if len(messages) == 1 {
		return SendMessage(messages[0], twilioChat, snsClient)
	} else {
		msgErr := SendMessage(messages[0], twilioChat, snsClient)
		if msgErr != nil {
			return msgErr
		}
		// To make sure messages are sent in order, only send the top and
		// all other messages are chained
		messages = messages[1:]
		messagesJSON, _ := json.Marshal(messages)
		return snsClient.Publish(string(messagesJSON), os.Getenv("SNS_TOPIC_ARN"), svc.SendSMSFeed)
	}
}

func main() {
	lambda.Start(handler)
}
