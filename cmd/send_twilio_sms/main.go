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

func handler(request events.SNSEvent) error {
	if len(request.Records) < 0 {
		return nil
	}
	message := request.Records[0].SNS.Message

	var data chat.Message
	err := json.Unmarshal([]byte(message), &data)
	if err != nil {
		return err
	}

	client := gotwilio.NewTwilioClient(os.Getenv("TWILIO_KEY"), os.Getenv("TWILIO_KEY"))
	twilioChat := svc.NewTwilioChat(
		client,
		os.Getenv("TWILIO_FROM"),
		data.Recipient,
		os.Getenv("STATUS_CALLBACK"),
		os.Getenv("APPLICATION_SID"),
	)
	twilioRes, twilioErr, sendErr := twilioChat.SendSMS(data.Body)

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
		Recipient: data.Recipient,
		Body:      data.Body,
		CreatedAt: &createdAt,
	}
	twilioMessageJSON, _ := json.Marshal(twilioMessage)
	snsClient := svc.NewSNSClient()
	return snsClient.Publish(string(twilioMessageJSON), os.Getenv("SNS_TOPIC_ARN"), svc.SentMessageFeed)
}

func main() {
	lambda.Start(handler)
}
